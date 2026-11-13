package mcpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/riccardo/freezo/backend/internal/db"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	f, err := os.CreateTemp("", "freezo-mcp-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	path := f.Name()
	f.Close()
	db.InitDB(path)
	t.Cleanup(func() {
		db.DB.Close()
		os.Remove(path)
		os.Remove(path + "-wal")
		os.Remove(path + "-shm")
	})
}

func connectInMemory(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	srv := NewServer()
	st, ct := mcp.NewInMemoryTransports()

	if _, err := srv.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect failed: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}
	t.Cleanup(func() { session.Close() })
	return session
}

func callToolJSON[T any](t *testing.T, session *mcp.ClientSession, name string, args map[string]any) T {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool(%s) failed: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("CallTool(%s) returned tool error: %+v", name, res.Content)
	}
	if len(res.Content) == 0 {
		t.Fatalf("CallTool(%s) returned empty content", name)
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("CallTool(%s) expected TextContent, got %T", name, res.Content[0])
	}
	var out T
	if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
		t.Fatalf("CallTool(%s) failed to unmarshal JSON (%s): %v", name, tc.Text, err)
	}
	return out
}

func TestMCPInventoryLifecycle(t *testing.T) {
	setupTestDB(t)
	session := connectInMemory(t)
	ctx := context.Background()

	// 1. Verify tools & resources are advertised
	toolsRes, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	if len(toolsRes.Tools) < 10 {
		t.Fatalf("expected at least 10 tools, got %d", len(toolsRes.Tools))
	}

	resourcesRes, err := session.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if len(resourcesRes.Resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resourcesRes.Resources))
	}

	// 2. Create two freezers
	callToolJSON[map[string]any](t, session, "create_freezer", map[string]any{"name": "Kitchen"})
	callToolJSON[map[string]any](t, session, "create_freezer", map[string]any{"name": "Garage"})

	// 3. Add items by freezer & category name (including an unknown-date item and dated items)
	added := callToolJSON[addItemsOutput](t, session, "add_items", map[string]any{
		"name":        "Chicken Breast",
		"freezer":     "kitchen",
		"category":    "poultry",
		"quantity":    2,
		"weight":      "500g",
		"frozen_date": "2026-01-10",
	})
	if added.AddedCount != 2 {
		t.Fatalf("expected 2 added items, got %d", added.AddedCount)
	}

	callToolJSON[addItemsOutput](t, session, "add_items", map[string]any{
		"name":         "Chicken Breast",
		"freezer":      "Kitchen",
		"category":     "Poultry",
		"quantity":     1,
		"unknown_date": true,
	})

	// 4. List inventory filtered by freezer & search
	inv := callToolJSON[listInventoryOutput](t, session, "list_inventory", map[string]any{
		"freezer": "Kitchen",
		"search":  "chicken",
	})
	if inv.Count != 3 {
		t.Fatalf("expected 3 Chicken Breast items in Kitchen, got %d", inv.Count)
	}
	// Unknown-date item should be first (FIFO priority)
	if inv.Items[0].FrozenDate != nil {
		t.Fatalf("expected unknown-date item first, got %v", inv.Items[0].FrozenDate)
	}

	// 5. Move 1 item to Garage
	moved := callToolJSON[moveItemsOutput](t, session, "move_items", map[string]any{
		"name":         "Chicken Breast",
		"from_freezer": "Kitchen",
		"to_freezer":   "Garage",
		"quantity":     1,
	})
	if moved.MovedCount != 1 || moved.ToFreezer != "Garage" {
		t.Fatalf("unexpected move result: %+v", moved)
	}

	// 6. Consume 1 item from Kitchen
	consumed := callToolJSON[consumeItemsOutput](t, session, "consume_items", map[string]any{
		"name":     "Chicken Breast",
		"freezer":  "Kitchen",
		"quantity": 1,
	})
	if consumed.ConsumedCount != 1 {
		t.Fatalf("expected 1 consumed item, got %d", consumed.ConsumedCount)
	}

	// 7. Verify summary & resource read
	summary := callToolJSON[inventorySummaryOutput](t, session, "get_inventory_summary", map[string]any{})
	if summary.TotalItems != 2 {
		t.Fatalf("expected 2 total items remaining, got %d", summary.TotalItems)
	}

	readRes, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: "freezo://inventory"})
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}
	if len(readRes.Contents) != 1 || readRes.Contents[0].Text == "" {
		t.Fatalf("expected non-empty resource contents")
	}
}

type handlerRoundTripper struct {
	handler http.Handler
}

func (rt handlerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	rt.handler.ServeHTTP(rec, req)
	return rec.Result(), nil
}

func TestStreamableHTTPHandler(t *testing.T) {
	setupTestDB(t)
	handler := NewHTTPHandler()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "http-test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{
		Endpoint:             "http://freezo.local/mcp",
		HTTPClient:           &http.Client{Transport: handlerRoundTripper{handler: handler}},
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("StreamableClientTransport Connect failed: %v", err)
	}
	defer session.Close()

	cats := callToolJSON[listCategoriesOutput](t, session, "list_categories", map[string]any{})
	if len(cats.Categories) == 0 {
		t.Fatalf("expected seeded categories, got 0")
	}
}
