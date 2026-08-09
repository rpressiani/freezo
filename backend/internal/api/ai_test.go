package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/riccardo/freezo/backend/internal/ai"
	"github.com/riccardo/freezo/backend/internal/mcp"
)

func TestMCPProtocol(t *testing.T) {
	setupTestDB()
	r := NewRouter()

	// 1. Test initialize
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
	}
	payload, _ := json.Marshal(initReq)
	req, _ := http.NewRequest("POST", "/api/mcp", bytes.NewBuffer(payload))
	resp := executeRequest(req, r)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 for MCP initialize, got %d", resp.Code)
	}

	var initResp mcp.RPCResponse
	json.Unmarshal(resp.Body.Bytes(), &initResp)
	if initResp.Error != nil {
		t.Fatalf("Unexpected RPC error on initialize: %v", initResp.Error)
	}
	resMap, ok := initResp.Result.(map[string]interface{})
	if !ok || resMap["protocolVersion"] != "2024-11-05" {
		t.Errorf("Unexpected initialize result: %v", initResp.Result)
	}

	// 2. Test tools/list
	listReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}
	payload, _ = json.Marshal(listReq)
	req, _ = http.NewRequest("POST", "/api/mcp", bytes.NewBuffer(payload))
	resp = executeRequest(req, r)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 for tools/list, got %d", resp.Code)
	}

	var listResp mcp.RPCResponse
	json.Unmarshal(resp.Body.Bytes(), &listResp)
	toolsResult, ok := listResp.Result.(map[string]interface{})
	toolsList, ok := toolsResult["tools"].([]interface{})
	if !ok || len(toolsList) < 5 {
		t.Errorf("Expected at least 5 tools listed, got %v", toolsResult)
	}

	// 3. Test tools/call: add_items with top-level parameters (e.g. Qwen style)
	callReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "add_items",
			"arguments": map[string]interface{}{
				"name":          "Bacon",
				"freezer_name":  "Garage Freezer",
				"category_name": "Pork",
				"count":         1,
			},
		},
	}
	payload, _ = json.Marshal(callReq)
	req, _ = http.NewRequest("POST", "/api/mcp", bytes.NewBuffer(payload))
	resp = executeRequest(req, r)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 for tools/call, got %d", resp.Code)
	}

	var callResp mcp.RPCResponse
	json.Unmarshal(resp.Body.Bytes(), &callResp)
	if callResp.Error != nil {
		t.Fatalf("Unexpected RPC error on tool call: %v", callResp.Error)
	}

	// 4. Test tools/call: list_items to verify items were added
	callReqList := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "list_items",
			"arguments": map[string]interface{}{},
		},
	}
	payload, _ = json.Marshal(callReqList)
	req, _ = http.NewRequest("POST", "/api/mcp", bytes.NewBuffer(payload))
	resp = executeRequest(req, r)

	var listItemsResp mcp.RPCResponse
	json.Unmarshal(resp.Body.Bytes(), &listItemsResp)
	if listItemsResp.Error != nil {
		t.Fatalf("tools/call list_items failed: %v", listItemsResp.Error)
	}
	resMap, ok = listItemsResp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not map[string]interface{}: %+v (raw body: %s)", listItemsResp.Result, resp.Body.String())
	}
	contentList, ok := resMap["content"].([]interface{})
	if !ok || len(contentList) == 0 {
		t.Fatalf("No content in result: %+v", resMap)
	}
	resContent := contentList[0].(map[string]interface{})["text"].(string)
	if !bytes.Contains([]byte(resContent), []byte("Bacon")) {
		t.Errorf("Expected Bacon in list_items output, got: %s", resContent)
	}
}

func TestAIConfigEndpoints(t *testing.T) {
	setupTestDB()
	r := NewRouter()

	// 1. GET AI Config
	req, _ := http.NewRequest("GET", "/api/ai/config", nil)
	resp := executeRequest(req, r)
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 for GET AI Config, got %d", resp.Code)
	}

	var cfg ai.Config
	json.Unmarshal(resp.Body.Bytes(), &cfg)
	if cfg.BaseURL == "" || cfg.Model == "" {
		t.Errorf("Expected default non-empty BaseURL and Model, got: %+v", cfg)
	}

	// 2. POST Update AI Config
	newCfg := ai.Config{
		BaseURL: "http://localhost:11434/v1",
		APIKey:  "test-key",
		Model:   "llama3.2",
	}
	payload, _ := json.Marshal(newCfg)
	req, _ = http.NewRequest("POST", "/api/ai/config", bytes.NewBuffer(payload))
	resp = executeRequest(req, r)
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 for POST AI Config, got %d", resp.Code)
	}

	var updatedCfg ai.Config
	json.Unmarshal(resp.Body.Bytes(), &updatedCfg)
	if updatedCfg.BaseURL != "http://localhost:11434/v1" || updatedCfg.Model != "llama3.2" || updatedCfg.APIKey != "test-key" {
		t.Errorf("Config was not updated properly: %+v", updatedCfg)
	}
}
