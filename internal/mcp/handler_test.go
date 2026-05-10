package mcp_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sharipovr/go-mcp-tool-gateway/internal/mcp"
	"github.com/sharipovr/go-mcp-tool-gateway/internal/tools"
)

func setupHandler() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := tools.NewRegistry()
	registry.Register(tools.EchoTool())
	registry.Register(tools.TimeTool())
	return mcp.NewHandler(registry, logger)
}

func postJSON(handler http.Handler, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestInitialize(t *testing.T) {
	h := setupHandler()
	rec := postJSON(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "test-client", "version": "1.0"},
		},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp mcp.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, _ := json.Marshal(resp.Result)
	var initResult mcp.InitializeResult
	json.Unmarshal(result, &initResult)

	if initResult.ProtocolVersion != mcp.ProtocolVersion {
		t.Errorf("expected protocol %s, got %s", mcp.ProtocolVersion, initResult.ProtocolVersion)
	}
	if initResult.ServerInfo.Name != mcp.ServerName {
		t.Errorf("expected server name %s, got %s", mcp.ServerName, initResult.ServerInfo.Name)
	}
}

func TestPing(t *testing.T) {
	h := setupHandler()
	rec := postJSON(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "ping",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp mcp.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
}

func TestToolsList(t *testing.T) {
	h := setupHandler()
	rec := postJSON(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	})

	var resp mcp.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, _ := json.Marshal(resp.Result)
	var listResult mcp.ToolsListResult
	json.Unmarshal(result, &listResult)

	if len(listResult.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(listResult.Tools))
	}

	names := map[string]bool{}
	for _, tool := range listResult.Tools {
		names[tool.Name] = true
	}
	if !names["echo"] || !names["current_time"] {
		t.Errorf("expected echo and current_time tools, got %v", names)
	}
}

func TestToolsCallEcho(t *testing.T) {
	h := setupHandler()
	rec := postJSON(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "echo",
			"arguments": map[string]any{"message": "hello MCP"},
		},
	})

	var resp mcp.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, _ := json.Marshal(resp.Result)
	var callResult mcp.ToolsCallResult
	json.Unmarshal(result, &callResult)

	if callResult.IsError {
		t.Fatal("expected success, got error")
	}
	if len(callResult.Content) == 0 {
		t.Fatal("expected content")
	}
	if callResult.Content[0].Text != "hello MCP" {
		t.Errorf("expected 'hello MCP', got %q", callResult.Content[0].Text)
	}
}

func TestToolsCallTime(t *testing.T) {
	h := setupHandler()
	rec := postJSON(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "current_time",
			"arguments": map[string]any{"timezone": "UTC"},
		},
	})

	var resp mcp.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	result, _ := json.Marshal(resp.Result)
	var callResult mcp.ToolsCallResult
	json.Unmarshal(result, &callResult)

	if callResult.IsError {
		t.Fatal("expected success, got error")
	}
	if len(callResult.Content) == 0 || callResult.Content[0].Text == "" {
		t.Fatal("expected non-empty time response")
	}
}

func TestToolsCallUnknown(t *testing.T) {
	h := setupHandler()
	rec := postJSON(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "nonexistent",
		},
	})

	var resp mcp.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != mcp.CodeInvalidParams {
		t.Errorf("expected code %d, got %d", mcp.CodeInvalidParams, resp.Error.Code)
	}
}

func TestNotification(t *testing.T) {
	h := setupHandler()
	rec := postJSON(h, map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}
}

func TestInvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := tools.NewRegistry()
	h := mcp.NewHandler(registry, logger)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var resp mcp.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil {
		t.Fatal("expected parse error")
	}
	if resp.Error.Code != mcp.CodeParseError {
		t.Errorf("expected code %d, got %d", mcp.CodeParseError, resp.Error.Code)
	}
}

func TestUnknownMethod(t *testing.T) {
	h := setupHandler()
	rec := postJSON(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "foo/bar",
	})

	var resp mcp.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error == nil {
		t.Fatal("expected method not found error")
	}
	if resp.Error.Code != mcp.CodeMethodNotFound {
		t.Errorf("expected code %d, got %d", mcp.CodeMethodNotFound, resp.Error.Code)
	}
}
