package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"red-tldr/internal/config"
	"red-tldr/internal/db"
	"red-tldr/internal/render"
	"red-tldr/internal/search"
	"red-tldr/internal/updater"
)

const banner = `
         /\_/\
     ____/ o o \   For Red Team [TL;DR]
   /~____  =ø= /   Github @Rvn0xsy
  (______)__m_m)   Blog: https://payloads.online
                   Version: 0.5.0
------------------------------------------------
Thank you for Use https://github.com/Rvn0xsy/red-tldr`

var (
	outputFormat string
	cfg          *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "red-tldr [keyword]",
	Short: "Red Team TL;DR - A lightweight red team command lookup tool",
	Long:  banner,
	Args:  cobra.MaximumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg = config.Load()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(banner)
			total := db.GetKeywordsTotal(cfg.GetDatabaseFilePath())
			fmt.Printf("Keywords Total: %d\n", total)
			cmd.Usage()
			return
		}
		searchAndDisplay(args[0])
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Rebuild the database index from local YAML files",
	Run: func(cmd *cobra.Command, args []string) {
		dbDir := cfg.GetDatabasePath()
		indexFile := cfg.GetDatabaseFilePath()
		if err := db.RebuildIndex(dbDir, indexFile); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		indexPath := search.DefaultIndexPath()
		os.RemoveAll(indexPath)
		engine, err := search.OpenOrCreate(indexPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to rebuild Bleve index: %v\n", err)
			return
		}
		defer engine.Close()
		if err := engine.BuildIndex(dbDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to rebuild Bleve index: %v\n", err)
		}
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Download the latest database from GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		if err := updater.FetchLatestFromGithub(cfg.GetDatabasePath()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "text",
		"Output format: text, json, markdown")
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(upgradeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func searchAndDisplay(keyword string) {
	dbDir := cfg.GetDatabasePath()

	if !config.DatabaseExists(cfg) {
		fmt.Println("[Database not found, downloading...]")
		if err := updater.FetchLatestFromGithub(dbDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading database: %v\n", err)
			os.Exit(1)
		}
	}

	r := render.New(outputFormat, cfg.Color)

	indexPath := search.DefaultIndexPath()
	if search.IndexExists(indexPath) {
		bleveSearch(keyword, indexPath, dbDir, r)
		return
	}

	fallbackSearch(keyword, dbDir, r)
}

func bleveSearch(keyword string, indexPath string, dbDir string, r *render.Renderer) {
	engine, err := search.OpenOrCreate(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Bleve unavailable, falling back to basic search\n")
		fallbackSearch(keyword, dbDir, r)
		return
	}
	defer engine.Close()

	results, err := engine.Search(search.SearchParams{Keyword: keyword, Limit: 20})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Bleve search failed, falling back to basic search\n")
		fallbackSearch(keyword, dbDir, r)
		return
	}

	// Exact match found — display directly
	if len(results) > 0 {
		displayResults(results, dbDir, r)
		return
	}

	// No exact match — try fuzzy search (edit distance <= 2)
	results, err = engine.FuzzySearch(keyword, 20)
	if err == nil && len(results) > 0 {
		displayResults(results, dbDir, r)
		return
	}

	// Fuzzy also empty — fall back to basic substring search
	fallbackSearch(keyword, dbDir, r)
}

func fallbackSearch(keyword string, dbDir string, r *render.Renderer) {
	indexFile := cfg.GetDatabaseFilePath()
	results, err := db.Search(indexFile, keyword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	displayResults(results, dbDir, r)
}

func displayResults(results []db.SearchResult, dbDir string, r *render.Renderer) {
	if len(results) == 0 {
		fmt.Println("[Search] Not Found.")
		return
	}

	if len(results) == 1 {
		showEntry(r, dbDir, results[0])
		return
	}

	if outputFormat == "json" || outputFormat == "markdown" || outputFormat == "md" {
		r.RenderResultList(results)
		return
	}

	r.RenderResultList(results)
	fmt.Printf("[Count: %d] > Select Result Number: ", len(results))
	var i int
	if _, err := fmt.Scanf("%d", &i); err != nil || i >= len(results) || i < 0 {
		i = 0
	}
	showEntry(r, dbDir, results[i])
}

func showEntry(r *render.Renderer, dbDir string, result db.SearchResult) {
	entry, err := db.LoadEntry(dbDir, result.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading entry: %v\n", err)
		return
	}
	r.RenderEntry(entry)
}
