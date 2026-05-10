package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const (
	ProtocolVersion = "2025-03-26"
	ServerName      = "go-mcp-tool-gateway"
	ServerVersion   = "0.1.0"
)

type ToolListener interface {
	List() []ToolDefinition
}

type ToolCaller interface {
	Call(name string, args json.RawMessage) (ToolsCallResult, error)
	Has(name string) bool
}

type ToolRegistry interface {
	ToolListener
	ToolCaller
}

type Handler struct {
	registry ToolRegistry
	log      *slog.Logger
}

func NewHandler(registry ToolRegistry, log *slog.Logger) *Handler {
	return &Handler{registry: registry, log: log}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
	if err != nil {
		h.writeError(w, nil, CodeParseError, "failed to read request body")
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeError(w, nil, CodeParseError, "invalid JSON")
		return
	}

	if req.JSONRPC != "2.0" {
		h.writeError(w, req.ID, CodeInvalidRequest, "jsonrpc must be \"2.0\"")
		return
	}

	h.log.Info("mcp request", "method", req.Method, "id", string(req.ID))

	if req.IsNotification() {
		h.handleNotification(w, &req)
		return
	}

	switch req.Method {
	case "initialize":
		h.handleInitialize(w, &req)
	case "ping":
		h.handlePing(w, &req)
	case "tools/list":
		h.handleToolsList(w, &req)
	case "tools/call":
		h.handleToolsCall(w, &req)
	default:
		h.writeError(w, req.ID, CodeMethodNotFound, fmt.Sprintf("unknown method: %s", req.Method))
	}
}

func (h *Handler) handleNotification(w http.ResponseWriter, req *Request) {
	switch req.Method {
	case "notifications/initialized":
		h.log.Info("client initialized")
	default:
		h.log.Warn("unknown notification", "method", req.Method)
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) handleInitialize(w http.ResponseWriter, req *Request) {
	var params InitializeParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			h.writeError(w, req.ID, CodeInvalidParams, "invalid initialize params")
			return
		}
	}

	h.log.Info("client connected",
		"client", params.ClientInfo.Name,
		"client_version", params.ClientInfo.Version,
		"protocol_version", params.ProtocolVersion,
	)

	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{ListChanged: false},
		},
		ServerInfo: Implementation{
			Name:    ServerName,
			Version: ServerVersion,
		},
		Instructions: "MCP Tool Gateway - a secure intermediary for I agent tool access. Use tools/list to discover available tools.",
	}

	h.writeResult(w, req.ID, result)
}

func (h *Handler) handlePing(w http.ResponseWriter, req *Request) {
	h.writeResult(w, req.ID, map[string]any{})
}

func (h *Handler) handleToolsList(w http.ResponseWriter, req *Request) {
	tools := h.registry.List()
	h.writeResult(w, req.ID, ToolsListResult{Tools: tools})
}

func (h *Handler) handleToolsCall(w http.ResponseWriter, req *Request) {
	var params ToolsCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.writeError(w, req.ID, CodeInvalidParams, "invalid tools/call params")
		return
	}

	if !h.registry.Has(params.Name) {
		h.writeError(w, req.ID, CodeInvalidParams, fmt.Sprintf("unknown tool: %s", params.Name))
		return
	}

	h.log.Info("tool call", "tool", params.Name)

	result, err := h.registry.Call(params.Name, params.Arguments)
	if err != nil {
		h.log.Error("tool execution failed", "tool", params.Name, "error", err)
		h.writeResult(w, req.ID, ToolsCallResult{
			Content: []ContentBlock{TextContent(fmt.Sprintf("internal error: %v", err))},
			IsError: true,
		})
		return
	}

	h.writeResult(w, req.ID, result)
}

func (h *Handler) writeResult(w http.ResponseWriter, id json.RawMessage, result any) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("failed to write response", "error", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, id json.RawMessage, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &RPCError{Code: code, Message: message},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("failed to write error response", "error", err)
	}
}
