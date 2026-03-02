package search

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestIndex(t *testing.T) (*BleveEngine, string) {
	t.Helper()
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "test.bleve")

	engine, err := OpenOrCreate(indexPath)
	if err != nil {
		t.Fatalf("failed to create index: %v", err)
	}
	return engine, dir
}

func createTestData(t *testing.T, dir string) {
	t.Helper()
	filesDir := filepath.Join(dir, "data", "files", "Active_directory")
	os.MkdirAll(filesDir, 0755)

	yaml1 := `name: mimikatz
tags:
  - mimikatz
  - credential
  - windows
  - hash
category: credential_access
platforms:
  - windows
mitre_attack:
  tactics:
    - TA0006
  techniques:
    - T1003.001
data: |
  # Mimikatz
  ## Dump credentials
  ` + "```" + `
  sekurlsa::logonpasswords
  ` + "```" + `
`
	yaml2 := `name: impacket-exec
tags:
  - impacket
  - windows
  - exec
  - lateral
category: lateral_movement
platforms:
  - windows
  - linux
mitre_attack:
  tactics:
    - TA0008
  techniques:
    - T1021.003
data: |
  # Impacket Remote Execution
  ` + "```" + `
  python3 wmiexec.py domain/user:pass@target
  ` + "```" + `
`
	yaml3 := `name: nmap-scan
tags:
  - nmap
  - scan
  - network
category: reconnaissance
platforms:
  - linux
  - windows
  - macos
data: |
  # Nmap Scanning
  ` + "```" + `
  nmap -sV -sC target
  ` + "```" + `
`
	os.WriteFile(filepath.Join(filesDir, "mimikatz.yaml"), []byte(yaml1), 0644)
	os.WriteFile(filepath.Join(filesDir, "impacket-exec.yaml"), []byte(yaml2), 0644)
	os.WriteFile(filepath.Join(filesDir, "nmap-scan.yml"), []byte(yaml3), 0644)
}

func TestBleveEngine_BuildAndSearch(t *testing.T) {
	engine, dir := setupTestIndex(t)
	defer engine.Close()

	dataDir := filepath.Join(dir, "data")
	createTestData(t, dir)

	if err := engine.BuildIndex(dataDir); err != nil {
		t.Fatalf("failed to build index: %v", err)
	}

	results, err := engine.Search(SearchParams{Keyword: "mimikatz", Limit: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 result for 'mimikatz'")
	}
	if results[0].Name != "mimikatz" {
		t.Errorf("expected first result 'mimikatz', got '%s'", results[0].Name)
	}
	if results[0].ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestBleveEngine_SearchWithATTACK(t *testing.T) {
	engine, dir := setupTestIndex(t)
	defer engine.Close()

	dataDir := filepath.Join(dir, "data")
	createTestData(t, dir)
	engine.BuildIndex(dataDir)

	results, err := engine.Search(SearchParams{Keyword: "mimikatz", Limit: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}

	r := results[0]
	if len(r.Tactics) == 0 || r.Tactics[0] != "TA0006" {
		t.Errorf("expected tactic TA0006, got %v", r.Tactics)
	}
	if len(r.Techniques) == 0 || r.Techniques[0] != "T1003.001" {
		t.Errorf("expected technique T1003.001, got %v", r.Techniques)
	}
	if r.Category != "credential_access" {
		t.Errorf("expected category 'credential_access', got '%s'", r.Category)
	}
}

func TestBleveEngine_FilterByTactic(t *testing.T) {
	engine, dir := setupTestIndex(t)
	defer engine.Close()

	dataDir := filepath.Join(dir, "data")
	createTestData(t, dir)
	engine.BuildIndex(dataDir)

	results, err := engine.Search(SearchParams{
		Keyword: "exec",
		Tactic:  "TA0008",
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for tactic TA0008")
	}
	if results[0].Name != "impacket-exec" {
		t.Errorf("expected 'impacket-exec', got '%s'", results[0].Name)
	}
}

func TestBleveEngine_SearchScoreRanking(t *testing.T) {
	engine, dir := setupTestIndex(t)
	defer engine.Close()

	dataDir := filepath.Join(dir, "data")
	createTestData(t, dir)
	engine.BuildIndex(dataDir)

	results, err := engine.Search(SearchParams{Keyword: "windows", Limit: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results for 'windows', got %d", len(results))
	}

	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("results not properly ranked: score[%d]=%f > score[%d]=%f",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestBleveEngine_SearchNoResults(t *testing.T) {
	engine, dir := setupTestIndex(t)
	defer engine.Close()

	dataDir := filepath.Join(dir, "data")
	createTestData(t, dir)
	engine.BuildIndex(dataDir)

	results, err := engine.Search(SearchParams{Keyword: "nonexistenttool12345", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestBleveEngine_FuzzySearch(t *testing.T) {
	engine, dir := setupTestIndex(t)
	defer engine.Close()

	dataDir := filepath.Join(dir, "data")
	createTestData(t, dir)
	engine.BuildIndex(dataDir)

	results, err := engine.FuzzySearch("mimikat", 10)
	if err != nil {
		t.Fatalf("fuzzy search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected fuzzy match for 'mimikat' -> 'mimikatz'")
	}
}

func TestBleveEngine_ContentSearch(t *testing.T) {
	engine, dir := setupTestIndex(t)
	defer engine.Close()

	dataDir := filepath.Join(dir, "data")
	createTestData(t, dir)
	engine.BuildIndex(dataDir)

	results, err := engine.Search(SearchParams{Keyword: "logonpasswords", Limit: 10})
	if err != nil {
		t.Fatalf("content search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected match for content keyword 'logonpasswords'")
	}
}

func TestBleveEngine_PreviewExtracted(t *testing.T) {
	engine, dir := setupTestIndex(t)
	defer engine.Close()

	dataDir := filepath.Join(dir, "data")
	createTestData(t, dir)
	engine.BuildIndex(dataDir)

	results, err := engine.Search(SearchParams{Keyword: "nmap", Limit: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Preview == "" {
		t.Error("expected non-empty preview")
	}
}

func TestIndexExists(t *testing.T) {
	dir := t.TempDir()
	if IndexExists(filepath.Join(dir, "nonexistent")) {
		t.Error("expected false for nonexistent path")
	}

	indexPath := filepath.Join(dir, "test.bleve")
	engine, err := OpenOrCreate(indexPath)
	if err != nil {
		t.Fatalf("failed to create index: %v", err)
	}
	engine.Close()

	if !IndexExists(indexPath) {
		t.Error("expected true for existing index")
	}
}

func TestOpenOrCreate_ReOpensExisting(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "test.bleve")

	engine1, err := OpenOrCreate(indexPath)
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	engine1.Close()

	engine2, err := OpenOrCreate(indexPath)
	if err != nil {
		t.Fatalf("reopening failed: %v", err)
	}
	engine2.Close()
}
