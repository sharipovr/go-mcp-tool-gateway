package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sharipovr/go-mcp-tool-gateway/internal/mcp"
)

func TimeTool() Tool {
	return Tool{
		Definition: mcp.ToolDefinition{
			Name:        "current_time",
			Description: "Return the current server time in the specified timezone (default: UTC).",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"timezone": map[string]any{
						"type":        "string",
						"description": "IANA timezone name (e.g., America/New_York, Europe/London). Defaults to UTC.",
						"default":     "UTC",
					},
				},
			},
		},
		Execute: func(args json.RawMessage) (mcp.ToolsCallResult, error) {
			var params struct {
				Timezone string `json:"timezone"`
			}
			if len(args) > 0 {
				if err := json.Unmarshal(args, &params); err != nil {
					return mcp.ToolsCallResult{}, fmt.Errorf("invalid argumants: %w", err)
				}
			}
			if params.Timezone == "" {
				params.Timezone = "UTC"
			}

			loc, err := time.LoadLocation(params.Timezone)
			if err != nil {
				return mcp.ToolsCallResult{
					Content: []mcp.ContentBlock{mcp.TextContent(fmt.Sprintf("error: unknown timezone %q", params.Timezone))},
					IsError: true,
				}, nil
			}

			now := time.Now().In(loc)
			result := fmt.Sprintf("%s (%s)", now.Format(time.RFC3339), params.Timezone)
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.TextContent(result)},
			}, nil
		},
	}
}
