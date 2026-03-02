package search

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"

	"red-tldr/internal/db"
)

type BleveEngine struct {
	index bleve.Index
	dbDir string
}

type BleveDocument struct {
	Name       string `json:"name"`
	Tags       string `json:"tags"`
	Data       string `json:"data"`
	Category   string `json:"category"`
	Platforms  string `json:"platforms"`
	Tactics    string `json:"tactics"`
	Techniques string `json:"techniques"`
	File       string `json:"file"`
}

func buildMapping() mapping.IndexMapping {
	im := mapping.NewIndexMapping()

	nameField := mapping.NewTextFieldMapping()
	nameField.Analyzer = "standard"

	tagsField := mapping.NewTextFieldMapping()
	tagsField.Analyzer = "standard"

	dataField := mapping.NewTextFieldMapping()
	dataField.Analyzer = "standard"

	categoryField := mapping.NewTextFieldMapping()
	categoryField.Analyzer = "keyword"

	platformsField := mapping.NewTextFieldMapping()
	platformsField.Analyzer = "keyword"

	tacticsField := mapping.NewTextFieldMapping()
	tacticsField.Analyzer = "keyword"

	techniquesField := mapping.NewTextFieldMapping()
	techniquesField.Analyzer = "keyword"

	fileField := mapping.NewTextFieldMapping()
	fileField.Store = true
	fileField.Index = false

	im.DefaultMapping.AddFieldMappingsAt("name", nameField)
	im.DefaultMapping.AddFieldMappingsAt("tags", tagsField)
	im.DefaultMapping.AddFieldMappingsAt("data", dataField)
	im.DefaultMapping.AddFieldMappingsAt("category", categoryField)
	im.DefaultMapping.AddFieldMappingsAt("platforms", platformsField)
	im.DefaultMapping.AddFieldMappingsAt("tactics", tacticsField)
	im.DefaultMapping.AddFieldMappingsAt("techniques", techniquesField)
	im.DefaultMapping.AddFieldMappingsAt("file", fileField)

	return im
}

func OpenOrCreate(indexPath string) (*BleveEngine, error) {
	idx, err := bleve.Open(indexPath)
	if err == nil {
		return &BleveEngine{index: idx}, nil
	}

	idx, err = bleve.New(indexPath, buildMapping())
	if err != nil {
		return nil, fmt.Errorf("failed to create bleve index: %w", err)
	}
	return &BleveEngine{index: idx}, nil
}

func (e *BleveEngine) Close() error {
	return e.index.Close()
}

func (e *BleveEngine) SetDbDir(dir string) {
	e.dbDir = dir
}

func (e *BleveEngine) BuildIndex(dbDir string) error {
	e.dbDir = dbDir
	files, err := db.ScanDataFiles(dbDir)
	if err != nil {
		return fmt.Errorf("failed to scan data files: %w", err)
	}

	batch := e.index.NewBatch()
	count := 0

	for _, file := range files {
		entry, err := parseEntryFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", file, err)
			continue
		}

		relPath := strings.TrimPrefix(file, dbDir)
		relPath = strings.TrimPrefix(relPath, string(os.PathSeparator))

		category := entry.Category
		if category == "" {
			category = db.ExtractCategoryFromPath(relPath)
		}

		var tactics, techniques string
		if entry.MitreAttack != nil {
			tactics = strings.Join(entry.MitreAttack.Tactics, " ")
			techniques = strings.Join(entry.MitreAttack.Techniques, " ")
		}

		doc := BleveDocument{
			Name:       entry.Name,
			Tags:       strings.Join(entry.Tags, " "),
			Data:       entry.Data,
			Category:   category,
			Platforms:  strings.Join(entry.Platforms, " "),
			Tactics:    tactics,
			Techniques: techniques,
			File:       relPath,
		}

		if err := batch.Index(relPath, doc); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to index %s: %v\n", file, err)
			continue
		}
		count++
	}

	if err := e.index.Batch(batch); err != nil {
		return fmt.Errorf("failed to commit batch: %w", err)
	}

	fmt.Printf("[Bleve index built: %d entries]\n", count)
	return nil
}

type SearchParams struct {
	Keyword   string
	Platform  string
	Category  string
	Tactic    string
	Technique string
	Limit     int
}

func (e *BleveEngine) Search(params SearchParams) ([]db.SearchResult, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}

	nameQ := bleve.NewMatchQuery(params.Keyword)
	nameQ.SetField("name")
	nameQ.SetBoost(5.0)

	tagsQ := bleve.NewMatchQuery(params.Keyword)
	tagsQ.SetField("tags")
	tagsQ.SetBoost(3.0)

	dataQ := bleve.NewMatchQuery(params.Keyword)
	dataQ.SetField("data")
	dataQ.SetBoost(1.0)

	mainQuery := bleve.NewDisjunctionQuery(nameQ, tagsQ, dataQ)

	hasFilter := params.Platform != "" || params.Category != "" || params.Tactic != "" || params.Technique != ""

	req := func() *bleve.SearchRequest {
		if hasFilter {
			boolQ := bleve.NewBooleanQuery()
			boolQ.AddMust(mainQuery)
			if params.Platform != "" {
				pQ := bleve.NewMatchQuery(params.Platform)
				pQ.SetField("platforms")
				boolQ.AddMust(pQ)
			}
			if params.Category != "" {
				cQ := bleve.NewMatchQuery(params.Category)
				cQ.SetField("category")
				boolQ.AddMust(cQ)
			}
			if params.Tactic != "" {
				tQ := bleve.NewMatchQuery(params.Tactic)
				tQ.SetField("tactics")
				boolQ.AddMust(tQ)
			}
			if params.Technique != "" {
				techQ := bleve.NewMatchQuery(params.Technique)
				techQ.SetField("techniques")
				boolQ.AddMust(techQ)
			}
			return bleve.NewSearchRequestOptions(boolQ, params.Limit, 0, false)
		}
		return bleve.NewSearchRequestOptions(mainQuery, params.Limit, 0, false)
	}()
	req.Fields = []string{"name", "tags", "category", "platforms", "tactics", "techniques", "data"}

	searchResult, err := e.index.Search(req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return e.hitsToResults(searchResult), nil
}

func (e *BleveEngine) FuzzySearch(keyword string, limit int) ([]db.SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}

	fuzzyQ := bleve.NewFuzzyQuery(keyword)
	fuzzyQ.SetFuzziness(2)

	req := bleve.NewSearchRequestOptions(fuzzyQ, limit, 0, false)
	req.Fields = []string{"name", "tags", "category", "platforms", "tactics", "techniques", "data"}

	searchResult, err := e.index.Search(req)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search failed: %w", err)
	}

	return e.hitsToResults(searchResult), nil
}

func (e *BleveEngine) hitsToResults(searchResult *bleve.SearchResult) []db.SearchResult {
	var results []db.SearchResult
	for _, hit := range searchResult.Hits {
		name := fieldString(hit.Fields, "name")
		category := fieldString(hit.Fields, "category")
		if category == "" {
			category = db.ExtractCategoryFromPath(hit.ID)
		}
		preview := db.ExtractPreview(fieldString(hit.Fields, "data"), 80)

		results = append(results, db.SearchResult{
			ID:         hit.ID,
			Name:       name,
			Tags:       strings.Fields(fieldString(hit.Fields, "tags")),
			Category:   category,
			Platforms:  splitNonEmpty(fieldString(hit.Fields, "platforms")),
			Tactics:    splitNonEmpty(fieldString(hit.Fields, "tactics")),
			Techniques: splitNonEmpty(fieldString(hit.Fields, "techniques")),
			Preview:    preview,
			Score:      hit.Score,
		})
	}
	return results
}

func fieldString(fields map[string]interface{}, key string) string {
	if v, ok := fields[key].(string); ok {
		return v
	}
	return ""
}

func splitNonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}

func IndexExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func DefaultIndexPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".red-tldr", "bleve.index")
}

func parseEntryFile(file string) (*db.Entry, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var entry db.Entry
	if err := yamlUnmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}
