package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"red-tldr/internal/db"
	"red-tldr/internal/search"
)

type Server struct {
	engine *search.BleveEngine
	dbDir  string
}

func NewServer(engine *search.BleveEngine, dbDir string) *Server {
	return &Server{engine: engine, dbDir: dbDir}
}

func (s *Server) buildMCPServer() *server.MCPServer {
	mcpServer := server.NewMCPServer(
		"red-tldr",
		"0.5.0",
		server.WithToolCapabilities(true),
	)

	mcpServer.AddTool(s.searchTool(), s.handleSearch)
	mcpServer.AddTool(s.detailsTool(), s.handleDetails)
	mcpServer.AddTool(s.listTool(), s.handleList)

	return mcpServer
}

func (s *Server) Serve() error {
	mcpServer := s.buildMCPServer()
	log.Println("[MCP Server] Starting on stdio...")
	return server.ServeStdio(mcpServer)
}

func (s *Server) ServeHTTP(addr string, endpoint string) error {
	mcpServer := s.buildMCPServer()

	opts := []server.StreamableHTTPOption{}
	if endpoint != "" {
		opts = append(opts, server.WithEndpointPath(endpoint))
	}

	httpServer := server.NewStreamableHTTPServer(mcpServer, opts...)

	ep := "/mcp"
	if endpoint != "" {
		ep = endpoint
	}

	log.Printf("[MCP Server] Streamable HTTP listening on http://%s%s\n", addr, ep)

	return httpServer.Start(addr)
}

func (s *Server) searchTool() mcp.Tool {
	return mcp.NewTool("search_redteam_commands",
		mcp.WithDescription("Search verified red team commands and techniques. Supports keyword search combined with filters for platform, ATT&CK tactic, and ATT&CK technique. Returns ranked results with name, category, tags, ATT&CK mapping, and a command preview."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search keyword (e.g. 'mimikatz', 'lateral movement', 'credential dumping')"),
		),
		mcp.WithString("platform",
			mcp.Description("Filter by OS platform: windows, linux, macos"),
		),
		mcp.WithString("tactic",
			mcp.Description("Filter by MITRE ATT&CK tactic ID (e.g. 'TA0006' for Credential Access)"),
		),
		mcp.WithString("technique",
			mcp.Description("Filter by MITRE ATT&CK technique ID (e.g. 'T1003' for OS Credential Dumping)"),
		),
		mcp.WithString("category",
			mcp.Description("Filter by category (e.g. 'Active_directory', 'Web', 'Linux')"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum results to return (default: 10)"),
		),
	)
}

func (s *Server) detailsTool() mcp.Tool {
	return mcp.NewTool("get_command_details",
		mcp.WithDescription("Get the full details of a specific red team command entry by its ID (from search results). Returns full command usage, examples, and notes."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Entry ID from search results"),
		),
	)
}

func (s *Server) listTool() mcp.Tool {
	return mcp.NewTool("list_techniques",
		mcp.WithDescription("List all available red team technique entries in the database, optionally filtered by a keyword."),
		mcp.WithString("keyword",
			mcp.Description("Optional keyword to filter entries"),
		),
	)
}

func toArgsMap(v any) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func argString(args map[string]interface{}, key string, fallback string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}

func argInt(args map[string]interface{}, key string, fallback int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return fallback
}

func (s *Server) handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := toArgsMap(req.Params.Arguments)

	query := argString(args, "query", "")
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}

	params := search.SearchParams{
		Keyword:   query,
		Platform:  argString(args, "platform", ""),
		Category:  argString(args, "category", ""),
		Tactic:    argString(args, "tactic", ""),
		Technique: argString(args, "technique", ""),
		Limit:     argInt(args, "limit", 10),
	}

	results, err := s.engine.Search(params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No results found."), nil
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleDetails(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := toArgsMap(req.Params.Arguments)
	id := argString(args, "id", "")
	if id == "" {
		return mcp.NewToolResultError("id parameter is required"), nil
	}

	entry, err := db.LoadEntry(s.dbDir, id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to load entry: %v", err)), nil
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format entry: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := toArgsMap(req.Params.Arguments)
	keyword := argString(args, "keyword", "")

	var results []db.SearchResult
	var err error

	if keyword != "" {
		results, err = s.engine.Search(search.SearchParams{Keyword: keyword, Limit: 50})
	} else {
		results, err = s.engine.Search(search.SearchParams{Keyword: "*", Limit: 100})
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText("No entries found."), nil
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format list: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}
