package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/sharipovr/go-mcp-tool-gateway/internal/mcp"
)

func HashTool() Tool {
	return Tool{
		Definition: mcp.ToolDefinition{
			Name:        "hash",
			Description: "Computes the SHA-256 hash of the provided input string,",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input": map[string]any{
						"type":        "string",
						"description": "The string to hash",
					},
				},
				"required": []string{"input"},
			},
		},
		Execute: func(args json.RawMessage) (mcp.ToolsCallResult, error) {
			var params struct {
				Input string `json:"input"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return mcp.ToolsCallResult{}, fmt.Errorf("invalid arguments: %w", err)
			}
			h := sha256.Sum256([]byte(params.Input))
			return mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{mcp.TextContent(hex.EncodeToString(h[:]))},
			}, nil
		},
	}
}
