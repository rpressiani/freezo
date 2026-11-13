# Freezo Agent Rules

## MCP Server Maintenance

Freezo exposes an embedded Model Context Protocol (MCP) server at `/mcp` (implemented in `backend/internal/mcpserver/server.go`).

Whenever you add, modify, or remove a user-facing feature, data model field, or backend REST API endpoint in Freezo:
1. **Update MCP Tools & Resources**: Ensure `backend/internal/mcpserver/server.go` exposes equivalent capabilities (tools, parameters, or resource fields) so AI assistants stay in sync with the web UI and REST API.
2. **Update MCP Tests**: Add or update test coverage in `backend/internal/mcpserver/server_test.go` for the new or changed MCP behavior.
3. **Update Documentation**: Update `README.md` if new MCP tools, resources, or configuration options are introduced.
4. **Verify**: Run `cd backend && go test -v ./...` before finishing.

