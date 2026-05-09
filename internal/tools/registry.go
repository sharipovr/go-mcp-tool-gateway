package tools

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/sharipovr/go-mcp-tool-gateway/internal/mcp"
)

type ExecuteFunc func(args json.RawMessage) (mcp.ToolsCallResult, error)

type Tool struct {
	Definition mcp.ToolDefinition
	Execute    ExecuteFunc
}

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Definition.Name] = t
}

func (r *Registry) List() []mcp.ToolDefinition {
	defs := make([]mcp.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, t.Definition)
	}
	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Name < defs[j].Name
	})
	return defs
}

func (r *Registry) Call(name string, args json.RawMessage) (mcp.ToolsCallResult, error) {
	t, ok := r.tools[name]
	if !ok {
		return mcp.ToolsCallResult{}, fmt.Errorf("unknown tool: %s", name)
	}
	return t.Execute(args)
}

func (r *Registry) Has(name string) bool {
	_, ok := r.tools[name]
	return ok
}
