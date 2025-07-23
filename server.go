package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	invocation := 0

	hooks := &server.Hooks{}

	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		fmt.Printf("beforeAny: %s, %v, %v\n", method, id, message)
	})
	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		fmt.Printf("beforeCallTool: %v, %v\n", id, message)
	})

	// Create a new MCP server
	s := server.NewMCPServer(
		"OpenAPI Error Checker",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
		server.WithHooks(hooks),
	)
	sse := server.NewSSEServer(s)

	// Add a error checking tool
	lintTool := mcp.NewTool("checkForErrors",
		mcp.WithDescription("Checks an OpenAPI YAML definitions for errors"),
		mcp.WithString("openAPIYAML",
			mcp.Required(),
			mcp.Description("The OpenAPI definition as a YAML string"),
		),
	)

	// Add the error checking handler
	s.AddTool(lintTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Using helper functions for type-safe argument access
		_, err := request.RequireString("openAPIYAML")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// This is a simple example so just alternate between returning an error and returning no errors
		invocation = invocation + 1
		if invocation%2 == 0 {
			fmt.Println("Indicating no errors")
			return mcp.NewToolResultText("No errors"), nil
		} else {
			fmt.Println("Indicating an error")
			return mcp.NewToolResultText("ERROR: A function name or path should never be too granular, for example never named \"add\" or \"subtract\" but rather \"calculate\""), nil
		}
	})

	// Start the server
	err := sse.Start(":8081")
	if err != nil {
		slog.Error("Unable to start server", "error", err)
		os.Exit(1)
	}
}
