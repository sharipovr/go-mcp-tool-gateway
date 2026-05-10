# Step 01 — MCP Protocol Foundation

## Goal

Build a minimal but fully functional MCP (Model Context Protocol) server in Go that speaks JSON-RPC 2.0 over HTTP. By the end of this step, any MCP client can connect, discover tools, and invoke them.

## What This Step Implements

- **JSON-RPC 2.0 transport** over HTTP POST (`/mcp` endpoint)
- **MCP handshake**: `initialize` → server capabilities → `notifications/initialized`
- **Tool discovery**: `tools/list` returns all registered tools with JSON Schema input definitions
- **Tool execution**: `tools/call` dispatches to the correct tool and returns structured content
- **Health check**: `GET /health` for readiness probes
- **Graceful shutdown**: catches SIGINT/SIGTERM and drains in-flight requests
- **Two demo tools**: `echo` (connectivity test) and `current_time` (timezone-aware)

## Project Structure

```
step-01/
├── cmd/server/main.go           # Entry point, wiring, graceful shutdown
├── internal/
│   ├── mcp/
│   │   ├── types.go             # JSON-RPC 2.0 + MCP protocol types
│   │   ├── handler.go           # Request dispatcher (initialize, tools/list, tools/call)
│   │   └── handler_test.go      # 9 tests covering all MCP methods + error paths
│   └── tools/
│       ├── registry.go          # Tool registry (register, list, call)
│       ├── echo.go              # Echo tool
│       └── time.go              # Current time tool
├── Dockerfile                   # Multi-stage build
├── Makefile                     # run, build, test, docker targets
├── go.mod
└── .gitignore
```

## How to Run

```bash
# Run directly
make run
# or
go run ./cmd/server

# Build and run binary
make build
./bin/mcp-gateway

# Run with Docker
make docker
docker run -p 8080:8080 mcp-tool-gateway:step-01
```

The server starts on `:8080` by default. Set `ADDR=:9090` to change the port.

## Try It — curl Recipes

### 1. Health check

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### 2. Initialize (MCP handshake)

```bash
curl -s http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "clientInfo": {"name": "curl-client", "version": "1.0"}
    }
  }' | jq .
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2025-03-26",
    "capabilities": { "tools": { "listChanged": false } },
    "serverInfo": { "name": "mcp-tool-gateway", "version": "0.1.0" },
    "instructions": "MCP Tool Gateway — a secure intermediary for AI agent tool access. Use tools/list to discover available tools."
  }
}
```

### 3. Send initialized notification

```bash
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "notifications/initialized"}'
# 202
```

### 4. List tools

```bash
curl -s http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}' | jq .
```

### 5. Call the echo tool

```bash
curl -s http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "echo",
      "arguments": {"message": "Hello from MCP!"}
    }
  }' | jq .
```

### 6. Call the current_time tool

```bash
curl -s http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "current_time",
      "arguments": {"timezone": "America/New_York"}
    }
  }' | jq .
```

## Tests

```bash
make test
```

9 tests cover: initialize, ping, tools/list, tools/call (echo, time, unknown tool), notifications, invalid JSON, unknown method.

## Plan for This Step

| #   | Task                                                           | Status |
| --- | -------------------------------------------------------------- | ------ |
| 1   | Define JSON-RPC 2.0 types                                      | Done   |
| 2   | Define MCP protocol types (initialize, tools/list, tools/call) | Done   |
| 3   | Implement request dispatcher with method routing               | Done   |
| 4   | Implement tool registry with register/list/call                | Done   |
| 5   | Create echo and current_time demo tools                        | Done   |
| 6   | Wire up HTTP server with graceful shutdown                     | Done   |
| 7   | Write comprehensive tests                                      | Done   |
| 8   | Add Dockerfile and Makefile                                    | Done   |

## Changes vs Previous Step

This is the **initial step** — there is no previous step to compare against.

**From zero to:**
- 7 Go source files (~400 lines of code)
- 9 passing tests
- Fully working MCP server with 2 tools
- Docker-ready with multi-stage build

## What's Next (Step 02 Preview)

**Tool Registry & Configuration** — instead of hardcoding tools in `main.go`, we'll:
- Load tool definitions from a YAML config file
- Add JSON Schema validation for tool inputs
- Add tool categories and metadata
- Support tool enable/disable without code changes
- Add a `tools/describe` helper for rich tool documentation

## Recommended Commit Message

```
feat: step-01 — MCP protocol foundation with JSON-RPC 2.0 transport and demo tools
```
