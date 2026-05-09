package tools

import (
	"encoding/json"
	"fmt"

	"github.com/sharipovr/go-mcp-tool-gateway/internal/mcp"
)

func EchoTool() Tool {
	return Tool{
		Definition: mcp.ToolDefinition{
			Name:        "echo",
			Description: "Echoes back the provided message. Useful for testing connectivity.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message": map[string]any{
						"type":        "string",
						"description": "The message to echo back",
					},
				},
				"required": []string{"message"},
			},
		},
		Execute: func(args json.RawMessage) (mcp.ToolsCallResult, error) {
			var params struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return mcp.ToolsCallResult{}, fmt.Errorf("invalid arguments: %w", err)
			}
			if params.Message == "" {
				return mcp.ToolsCallResult{
					Content: []mcp.ContentBlock{mcp.TextContent("error: message is required")},
					IsError: true,
				}, nil
			}
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.TextContent(params.Message)},
			}, nil
		},
	}
}
