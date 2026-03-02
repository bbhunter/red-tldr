package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SupportedExtensions lists accepted data-file extensions.
var SupportedExtensions = []string{".yaml", ".yml"}

// RebuildIndex scans all data files under dbDir and writes a new JSON index.
func RebuildIndex(dbDir string, indexFile string) error {
	fileList, err := ScanDataFiles(dbDir)
	if err != nil {
		return fmt.Errorf("failed to scan data files: %w", err)
	}
	if len(fileList) == 0 {
		return fmt.Errorf("no data files found in %s", dbDir)
	}

	var idx Index
	for _, file := range fileList {
		entry, err := parseDataFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", file, err)
			continue
		}
		relPath := strings.TrimPrefix(file, dbDir)
		relPath = strings.TrimPrefix(relPath, string(os.PathSeparator))
		idx.Data = append(idx.Data, IndexEntry{
			Name: entry.Name,
			Tags: entry.Tags,
			File: relPath,
		})
	}

	jsonData, err := json.Marshal(&idx)
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(indexFile), 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	if err := os.WriteFile(indexFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write index file: %w", err)
	}

	fmt.Printf("[Index updated: %d entries]\n", len(idx.Data))
	return nil
}

// ScanDataFiles walks dbDir and returns paths to all .yaml/.yml files.
func ScanDataFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		for _, supported := range SupportedExtensions {
			if ext == supported {
				files = append(files, path)
				break
			}
		}
		return nil
	})
	return files, err
}

func parseDataFile(file string) (*Entry, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var entry Entry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}
