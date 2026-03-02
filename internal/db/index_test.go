package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestScanDataFiles_FindsBothExtensions(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "files", "Web"), 0755)
	os.WriteFile(filepath.Join(dir, "files", "Web", "test.yaml"), []byte("name: a"), 0644)

	os.MkdirAll(filepath.Join(dir, "files", "Database"), 0755)
	os.WriteFile(filepath.Join(dir, "files", "Database", "test.yml"), []byte("name: b"), 0644)

	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644)

	files, err := ScanDataFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("expected 2 files (.yaml + .yml), got %d: %v", len(files), files)
	}

	foundYaml, foundYml := false, false
	for _, f := range files {
		if filepath.Ext(f) == ".yaml" {
			foundYaml = true
		}
		if filepath.Ext(f) == ".yml" {
			foundYml = true
		}
	}
	if !foundYaml {
		t.Error("expected to find .yaml file")
	}
	if !foundYml {
		t.Error("expected to find .yml file")
	}
}

func TestRebuildIndex(t *testing.T) {
	dir := t.TempDir()

	subDir := filepath.Join(dir, "files", "Test")
	os.MkdirAll(subDir, 0755)

	yaml1 := `name: tool-a
tags:
  - tool
  - alpha
data: |
  Tool A usage
`
	yaml2 := `name: tool-b
tags:
  - tool
  - beta
data: |
  Tool B usage
`
	os.WriteFile(filepath.Join(subDir, "tool-a.yaml"), []byte(yaml1), 0644)
	os.WriteFile(filepath.Join(subDir, "tool-b.yml"), []byte(yaml2), 0644)

	indexDir := filepath.Join(dir, "db")
	indexFile := filepath.Join(indexDir, "db.json")

	err := RebuildIndex(dir, indexFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(indexFile)
	if err != nil {
		t.Fatalf("failed to read index: %v", err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		t.Fatalf("failed to parse index: %v", err)
	}

	if len(idx.Data) != 2 {
		t.Fatalf("expected 2 entries in index, got %d", len(idx.Data))
	}
}

func TestRebuildIndex_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	indexFile := filepath.Join(dir, "db", "db.json")

	err := RebuildIndex(dir, indexFile)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}
