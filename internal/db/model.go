package db

import (
	"os"
	"strings"
)

type Entry struct {
	Name      string    `yaml:"name" json:"name"`
	Tags      []string  `yaml:"tags" json:"tags"`
	Data      string    `yaml:"data" json:"data"`
	Category  string    `yaml:"category,omitempty" json:"category,omitempty"`
	Platforms []string  `yaml:"platforms,omitempty" json:"platforms,omitempty"`
	Metadata  *Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Related   []string  `yaml:"related,omitempty" json:"related,omitempty"`

	MitreAttack *MitreAttack `yaml:"mitre_attack,omitempty" json:"mitre_attack,omitempty"`
}

type MitreAttack struct {
	Tactics    []string `yaml:"tactics,omitempty" json:"tactics,omitempty"`
	Techniques []string `yaml:"techniques,omitempty" json:"techniques,omitempty"`
}

type Metadata struct {
	Author       string `yaml:"author,omitempty" json:"author,omitempty"`
	Source       string `yaml:"source,omitempty" json:"source,omitempty"`
	VerifiedOn   string `yaml:"verified_on,omitempty" json:"verified_on,omitempty"`
	LastVerified string `yaml:"last_verified,omitempty" json:"last_verified,omitempty"`
	Confidence   string `yaml:"confidence,omitempty" json:"confidence,omitempty"`
}

type IndexEntry struct {
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	File      string   `json:"file"`
	Category  string   `json:"category,omitempty"`
	Platforms []string `json:"platforms,omitempty"`
}

type Index struct {
	Data []IndexEntry `json:"data"`
}

type SearchResult struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Tags       []string `json:"tags"`
	Category   string   `json:"category"`
	Platforms  []string `json:"platforms,omitempty"`
	Tactics    []string `json:"tactics,omitempty"`
	Techniques []string `json:"techniques,omitempty"`
	Preview    string   `json:"preview,omitempty"`
	Score      float64  `json:"score,omitempty"`
}

func ExtractCategoryFromPath(filePath string) string {
	parts := strings.Split(filePath, string(os.PathSeparator))
	if len(parts) == 1 {
		parts = strings.Split(filePath, "/")
	}
	if len(parts) >= 2 {
		return parts[len(parts)-2]
	}
	return ""
}

func ExtractPreview(data string, maxLen int) string {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "```") || strings.HasPrefix(line, "%") {
			continue
		}
		if len(line) > maxLen {
			return line[:maxLen] + "..."
		}
		return line
	}
	return ""
}
