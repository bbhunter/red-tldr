package render

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gookit/color"
	"red-tldr/internal/db"
)

type Format string

const (
	FormatText     Format = "text"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

type Renderer struct {
	format   Format
	useColor bool
}

func New(format string, useColor bool) *Renderer {
	f := FormatText
	switch strings.ToLower(format) {
	case "json":
		f = FormatJSON
	case "markdown", "md":
		f = FormatMarkdown
	}
	return &Renderer{format: f, useColor: useColor}
}

func (r *Renderer) RenderEntry(entry *db.Entry) {
	switch r.format {
	case FormatJSON:
		r.renderJSON(entry)
	case FormatMarkdown:
		r.renderMarkdown(entry)
	default:
		r.renderText(entry)
	}
}

func (r *Renderer) RenderResultList(results []db.SearchResult) {
	switch r.format {
	case FormatJSON:
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	case FormatMarkdown:
		fmt.Println("| # | Name | Category | Tags | ATT&CK | Preview |")
		fmt.Println("|---|------|----------|------|--------|---------|")
		for i, res := range results {
			attack := formatAttackBrief(res.Tactics, res.Techniques)
			fmt.Printf("| %d | %s | %s | %s | %s | %s |\n",
				i, res.Name, res.Category,
				strings.Join(res.Tags, ", "),
				attack, res.Preview)
		}
	default:
		r.renderTextList(results)
	}
}

func (r *Renderer) renderTextList(results []db.SearchResult) {
	for i, res := range results {
		attack := formatAttackBrief(res.Tactics, res.Techniques)

		if r.useColor {
			num := color.Style{color.FgWhite, color.OpBold}.Sprintf("%d)", i)
			name := color.Style{color.FgCyan, color.OpBold}.Sprint(res.Name)
			cat := color.Style{color.FgYellow}.Sprintf("[%s]", res.Category)
			tags := color.Style{color.FgGray}.Sprint(strings.Join(res.Tags, ", "))
			fmt.Printf("%s %s %s  %s", num, name, cat, tags)
			if attack != "" {
				fmt.Printf("  %s", color.Style{color.FgMagenta}.Sprint(attack))
			}
			fmt.Println()
			if res.Preview != "" {
				fmt.Printf("   %s\n", color.Style{color.FgWhite}.Sprint(res.Preview))
			}
		} else {
			fmt.Printf("%d) %s [%s]  %s", i, res.Name, res.Category, strings.Join(res.Tags, ", "))
			if attack != "" {
				fmt.Printf("  %s", attack)
			}
			fmt.Println()
			if res.Preview != "" {
				fmt.Printf("   %s\n", res.Preview)
			}
		}
	}
}

func formatAttackBrief(tactics []string, techniques []string) string {
	var parts []string
	if len(tactics) > 0 {
		parts = append(parts, strings.Join(tactics, ","))
	}
	if len(techniques) > 0 {
		parts = append(parts, strings.Join(techniques, ","))
	}
	if len(parts) == 0 {
		return ""
	}
	return "ATT&CK:" + strings.Join(parts, "/")
}

func (r *Renderer) renderJSON(entry *db.Entry) {
	data, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Println(string(data))
}

func (r *Renderer) renderMarkdown(entry *db.Entry) {
	fmt.Printf("# %s\n\n", entry.Name)
	fmt.Println(entry.Data)
}

func (r *Renderer) renderText(entry *db.Entry) {
	if !r.useColor {
		fmt.Println("=================")
		fmt.Println(entry.Name)
		fmt.Println("=================")
		fmt.Println(entry.Data)
		return
	}

	color.Style{color.Red, color.OpBold}.Println(entry.Name)
	lines := strings.Split(entry.Data, "\n")
	inCodeBlock := false
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				inCodeBlock = false
				continue
			}
			fmt.Println()
			inCodeBlock = true
			continue
		}
		if inCodeBlock {
			color.Style{color.Green, color.OpBold}.Println(line)
			continue
		}
		fmt.Println(line)
	}
}
