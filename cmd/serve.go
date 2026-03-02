package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	mcpsrv "red-tldr/internal/mcp"
	"red-tldr/internal/search"
)

var (
	httpMode     bool
	httpAddr     string
	httpEndpoint string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start as a server (MCP mode for AI assistants)",
	Long: `Start the MCP server for AI assistant integration.

Modes:
  stdio (default)    — for Claude Desktop, Cursor, etc.
  HTTP  (--http)     — Streamable HTTP for web clients and remote access`,
	Run: func(cmd *cobra.Command, args []string) {
		dbDir := cfg.GetDatabasePath()
		indexPath := search.DefaultIndexPath()

		if !search.IndexExists(indexPath) {
			fmt.Println("[Building search index...]")
			engine, err := search.OpenOrCreate(indexPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating index: %v\n", err)
				os.Exit(1)
			}
			if err := engine.BuildIndex(dbDir); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
				fmt.Fprintf(os.Stderr, "Server will start with empty index. Run 'red-tldr upgrade' or clone the DB first.\n")
			}
			engine.Close()
		}

		engine, err := search.OpenOrCreate(indexPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening index: %v\n", err)
			os.Exit(1)
		}
		defer engine.Close()
		engine.SetDbDir(dbDir)

		srv := mcpsrv.NewServer(engine, dbDir)

		if httpMode {
			if err := srv.ServeHTTP(httpAddr, httpEndpoint); err != nil {
				fmt.Fprintf(os.Stderr, "MCP HTTP Server error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := srv.Serve(); err != nil {
				fmt.Fprintf(os.Stderr, "MCP Server error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	serveCmd.Flags().BoolVar(&httpMode, "http", false, "Use Streamable HTTP transport instead of stdio")
	serveCmd.Flags().StringVar(&httpAddr, "addr", "localhost:8080", "HTTP server listen address")
	serveCmd.Flags().StringVar(&httpEndpoint, "endpoint", "/mcp", "HTTP endpoint path")
	rootCmd.AddCommand(serveCmd)
}
