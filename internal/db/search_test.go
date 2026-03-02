package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesTags_ExactMatch(t *testing.T) {
	tags := []string{"mimikatz", "credential", "windows"}
	if !matchesTags(tags, "mimikatz") {
		t.Error("expected exact match to succeed")
	}
}

func TestMatchesTags_SubstringMatch(t *testing.T) {
	tags := []string{"mimikatz", "credential-dumping", "windows"}
	if !matchesTags(tags, "credential") {
		t.Error("expected substring match to succeed")
	}
}

func TestMatchesTags_PrefixMatch(t *testing.T) {
	tags := []string{"mimikatz", "credential", "windows"}
	if !matchesTags(tags, "mimi") {
		t.Error("expected prefix match to succeed")
	}
}

func TestMatchesTags_CaseInsensitive(t *testing.T) {
	tags := []string{"Mimikatz", "Credential", "Windows"}
	if !matchesTags(tags, "mimikatz") {
		t.Error("expected case-insensitive match to succeed")
	}
}

func TestMatchesTags_NoMatch(t *testing.T) {
	tags := []string{"mimikatz", "credential", "windows"}
	if matchesTags(tags, "linux") {
		t.Error("expected no match")
	}
}

func TestSearch_ReturnsResults(t *testing.T) {
	// Create a temporary index file
	dir := t.TempDir()
	indexFile := filepath.Join(dir, "db.json")

	idx := Index{
		Data: []IndexEntry{
			{Name: "mimikatz", Tags: []string{"mimikatz", "credential"}, File: "files/mimikatz.yaml"},
			{Name: "nmap", Tags: []string{"nmap", "scan", "network"}, File: "files/nmap.yaml"},
			{Name: "impacket", Tags: []string{"impacket", "windows"}, File: "files/impacket.yaml"},
		},
	}
	data, _ := json.Marshal(idx)
	os.WriteFile(indexFile, data, 0644)

	results, err := Search(indexFile, "mimikatz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "mimikatz" {
		t.Errorf("expected name 'mimikatz', got '%s'", results[0].Name)
	}
}

func TestSearch_NoResults(t *testing.T) {
	dir := t.TempDir()
	indexFile := filepath.Join(dir, "db.json")

	idx := Index{
		Data: []IndexEntry{
			{Name: "nmap", Tags: []string{"nmap", "scan"}, File: "files/nmap.yaml"},
		},
	}
	data, _ := json.Marshal(idx)
	os.WriteFile(indexFile, data, 0644)

	results, err := Search(indexFile, "mimikatz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_MultipleResults(t *testing.T) {
	dir := t.TempDir()
	indexFile := filepath.Join(dir, "db.json")

	idx := Index{
		Data: []IndexEntry{
			{Name: "nmap-basic", Tags: []string{"nmap", "scan"}, File: "files/nmap-basic.yaml"},
			{Name: "nmap-vuln", Tags: []string{"nmap", "vuln"}, File: "files/nmap-vuln.yaml"},
			{Name: "masscan", Tags: []string{"masscan", "scan"}, File: "files/masscan.yaml"},
		},
	}
	data, _ := json.Marshal(idx)
	os.WriteFile(indexFile, data, 0644)

	results, err := Search(indexFile, "nmap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSearch_FileNotFound(t *testing.T) {
	_, err := Search("/nonexistent/path/db.json", "test")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadEntry(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `name: test-tool
tags:
  - test
  - example
data: |
  # Test Tool
  This is a test entry.
`
	subDir := filepath.Join(dir, "files")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "test.yaml"), []byte(yamlContent), 0644)

	entry, err := LoadEntry(dir, "files/test.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Name != "test-tool" {
		t.Errorf("expected name 'test-tool', got '%s'", entry.Name)
	}
	if len(entry.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(entry.Tags))
	}
}

func TestGetKeywordsTotal(t *testing.T) {
	dir := t.TempDir()
	indexFile := filepath.Join(dir, "db.json")

	idx := Index{
		Data: []IndexEntry{
			{Name: "a", Tags: []string{"a"}, File: "a.yaml"},
			{Name: "b", Tags: []string{"b"}, File: "b.yaml"},
			{Name: "c", Tags: []string{"c"}, File: "c.yaml"},
		},
	}
	data, _ := json.Marshal(idx)
	os.WriteFile(indexFile, data, 0644)

	total := GetKeywordsTotal(indexFile)
	if total != 3 {
		t.Errorf("expected 3, got %d", total)
	}
}

func TestGetKeywordsTotal_MissingFile(t *testing.T) {
	total := GetKeywordsTotal("/nonexistent/db.json")
	if total != 0 {
		t.Errorf("expected 0 for missing file, got %d", total)
	}
}
