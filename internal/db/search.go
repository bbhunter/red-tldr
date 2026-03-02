package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Search looks up entries whose tags match the given keyword.
func Search(indexFile string, keyword string) ([]SearchResult, error) {
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read index file %s: %w", indexFile, err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("cannot parse index file: %w", err)
	}

	keyword = strings.ToLower(keyword)
	var results []SearchResult
	for _, entry := range idx.Data {
		if matchesTags(entry.Tags, keyword) {
			results = append(results, SearchResult{
				ID:       entry.File,
				Name:     entry.Name,
				Tags:     entry.Tags,
				Category: ExtractCategoryFromPath(entry.File),
			})
		}
	}
	return results, nil
}

// matchesTags returns true if any tag matches the keyword by equality,
// substring, or prefix (case-insensitive).
func matchesTags(tags []string, keyword string) bool {
	for _, tag := range tags {
		t := strings.ToLower(tag)
		if keyword == t || strings.Contains(t, keyword) || strings.HasPrefix(t, keyword) {
			return true
		}
	}
	return false
}

// LoadEntry reads and parses a single YAML entry from disk.
func LoadEntry(dbDir string, relPath string) (*Entry, error) {
	fullPath := filepath.Join(dbDir, relPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read entry file %s: %w", fullPath, err)
	}

	var entry Entry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("cannot parse entry file %s: %w", fullPath, err)
	}
	return &entry, nil
}

// GetKeywordsTotal returns the total number of entries in the index.
func GetKeywordsTotal(indexFile string) int {
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return 0
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return 0
	}
	return len(idx.Data)
}
