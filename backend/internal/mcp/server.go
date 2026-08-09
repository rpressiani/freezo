package mcp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/riccardo/freezo/backend/internal/db"
	"github.com/riccardo/freezo/backend/internal/models"
)

// JSON-RPC 2.0 Structures
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool Definition Structures
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// List of available MCP tools
func GetToolDefinitions() []Tool {
	return []Tool{
		{
			Name:        "list_freezers",
			Description: "Get list of all freezers in Freezo",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "list_categories",
			Description: "Get list of all food categories in Freezo",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "list_items",
			Description: "Get list of all food items currently in freezer inventory",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "add_items",
			Description: "Add one or more food items to a freezer inventory",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"items": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name":          map[string]interface{}{"type": "string", "description": "Name of the food item"},
								"freezer_name":  map[string]interface{}{"type": "string", "description": "Name of target freezer (e.g., Kitchen Freezer)"},
								"freezer_id":    map[string]interface{}{"type": "integer", "description": "ID of target freezer if known"},
								"category_name": map[string]interface{}{"type": "string", "description": "Category name (e.g., Pork, Beef, Seafood, Poultry, Bread, Uncategorized)"},
								"category_id":   map[string]interface{}{"type": "integer", "description": "Category ID if known"},
								"weight":        map[string]interface{}{"type": "string", "description": "Optional weight or package size e.g. 2 lbs"},
								"count":         map[string]interface{}{"type": "integer", "description": "Number of items to add (default 1)"},
								"frozen_date":   map[string]interface{}{"type": "string", "description": "ISO date string e.g. 2026-08-09"},
							},
							"required": []string{"name"},
						},
					},
				},
				"required": []string{"items"},
			},
		},
		{
			Name:        "consume_items",
			Description: "Remove/consume food items from inventory matching item name, freezer, or item IDs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"item_name":    map[string]interface{}{"type": "string", "description": "Name of the food item to consume"},
					"freezer_name": map[string]interface{}{"type": "string", "description": "Optional name of freezer to consume from"},
					"item_ids":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "integer"}, "description": "Exact item IDs to delete"},
					"count":        map[string]interface{}{"type": "integer", "description": "Number of items to consume (default 1)"},
				},
			},
		},
		{
			Name:        "move_items",
			Description: "Move items from one freezer to another",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"item_name":          map[string]interface{}{"type": "string", "description": "Name of items to move"},
					"target_freezer_name": map[string]interface{}{"type": "string", "description": "Destination freezer name"},
					"target_freezer_id":   map[string]interface{}{"type": "integer", "description": "Destination freezer ID if known"},
					"item_ids":           map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "integer"}, "description": "Specific item IDs to move"},
					"count":              map[string]interface{}{"type": "integer", "description": "Number of items to move (default all matching)"},
				},
			},
		},
		{
			Name:        "update_items_category",
			Description: "Update category of items by item name or item IDs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"item_name":     map[string]interface{}{"type": "string", "description": "Name of items to recategorize"},
					"category_name": map[string]interface{}{"type": "string", "description": "Target category name"},
					"category_id":   map[string]interface{}{"type": "integer", "description": "Target category ID if known"},
					"item_ids":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "integer"}, "description": "Specific item IDs to update"},
				},
			},
		},
	}
}

// Tool Execution Helpers
func ExecuteTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "list_freezers":
		rows, err := db.DB.Query("SELECT id, name FROM freezers")
		if err != nil {
			return "", err
		}
		defer rows.Close()
		var freezers []models.Freezer
		for rows.Next() {
			var f models.Freezer
			if err := rows.Scan(&f.ID, &f.Name); err != nil {
				return "", err
			}
			freezers = append(freezers, f)
		}
		out, _ := json.Marshal(freezers)
		return string(out), nil

	case "list_categories":
		rows, err := db.DB.Query("SELECT id, name, COALESCE(icon, '') FROM categories")
		if err != nil {
			return "", err
		}
		defer rows.Close()
		var categories []models.Category
		for rows.Next() {
			var c models.Category
			if err := rows.Scan(&c.ID, &c.Name, &c.Icon); err != nil {
				return "", err
			}
			categories = append(categories, c)
		}
		out, _ := json.Marshal(categories)
		return string(out), nil

	case "list_items":
		rows, err := db.DB.Query(`
			SELECT i.id, i.name, COALESCE(i.category_id, 0), COALESCE(c.name, ''), i.freezer_id, COALESCE(f.name, ''), COALESCE(i.weight, ''), i.frozen_date
			FROM items i
			LEFT JOIN categories c ON i.category_id = c.id
			LEFT JOIN freezers f ON i.freezer_id = f.id
		`)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		type ItemWithNames struct {
			ID           int64  `json:"id"`
			Name         string `json:"name"`
			CategoryID   int64  `json:"category_id"`
			CategoryName string `json:"category_name"`
			FreezerID    int64  `json:"freezer_id"`
			FreezerName  string `json:"freezer_name"`
			Weight       string `json:"weight,omitempty"`
			FrozenDate   string `json:"frozen_date,omitempty"`
		}

		var items []ItemWithNames
		for rows.Next() {
			var item ItemWithNames
			var frozenDate sql.NullTime
			if err := rows.Scan(&item.ID, &item.Name, &item.CategoryID, &item.CategoryName, &item.FreezerID, &item.FreezerName, &item.Weight, &frozenDate); err != nil {
				return "", err
			}
			if frozenDate.Valid {
				item.FrozenDate = frozenDate.Time.Format("2006-01-02")
			}
			items = append(items, item)
		}
		out, _ := json.Marshal(items)
		return string(out), nil

	case "add_items":
		var itemsRaw []interface{}
		if rawList, ok := args["items"].([]interface{}); ok && len(rawList) > 0 {
			itemsRaw = rawList
		} else if rawItem, ok := args["item"].(map[string]interface{}); ok {
			itemsRaw = []interface{}{rawItem}
		} else if args != nil {
			if _, ok := args["name"]; ok {
				itemsRaw = []interface{}{args}
			} else if _, ok := args["item_name"]; ok {
				itemsRaw = []interface{}{args}
			}
		}

		if len(itemsRaw) == 0 {
			return "", fmt.Errorf("invalid or missing item parameters")
		}

		var addedItems []models.Item
		for _, raw := range itemsRaw {
			itemMap, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}

			itemName, _ := itemMap["name"].(string)
			if itemName == "" {
				itemName, _ = itemMap["item_name"].(string)
			}
			if itemName == "" {
				continue
			}

			// Resolve freezer_id
			freezerID := int64(0)
			if fid, ok := itemMap["freezer_id"].(float64); ok {
				freezerID = int64(fid)
			}
			if freezerID == 0 {
				fName, _ := itemMap["freezer_name"].(string)
				if fName == "" {
					fName, _ = itemMap["freezer"].(string)
				}
				if fName != "" {
					_ = db.DB.QueryRow("SELECT id FROM freezers WHERE LOWER(name) = LOWER(?)", strings.TrimSpace(fName)).Scan(&freezerID)
				}
			}
			// If freezer not found, select first freezer or create default
			if freezerID == 0 {
				_ = db.DB.QueryRow("SELECT id FROM freezers ORDER BY id LIMIT 1").Scan(&freezerID)
			}
			if freezerID == 0 {
				res, err := db.DB.Exec("INSERT INTO freezers (name) VALUES (?)", "Main Freezer")
				if err == nil {
					freezerID, _ = res.LastInsertId()
				}
			}

			// Resolve category_id
			categoryID := int64(0)
			if cid, ok := itemMap["category_id"].(float64); ok {
				categoryID = int64(cid)
			}
			if categoryID == 0 {
				cName, _ := itemMap["category_name"].(string)
				if cName == "" {
					cName, _ = itemMap["category"].(string)
				}
				if cName != "" {
					_ = db.DB.QueryRow("SELECT id FROM categories WHERE LOWER(name) = LOWER(?)", strings.TrimSpace(cName)).Scan(&categoryID)
				}
			}
			if categoryID == 0 {
				_ = db.DB.QueryRow("SELECT id FROM categories WHERE LOWER(name) = 'uncategorized' LIMIT 1").Scan(&categoryID)
			}
			if categoryID == 0 {
				_ = db.DB.QueryRow("SELECT id FROM categories ORDER BY id LIMIT 1").Scan(&categoryID)
			}

			weight, _ := itemMap["weight"].(string)
			frozenDateStr, _ := itemMap["frozen_date"].(string)
			frozenDate := time.Now()
			if frozenDateStr != "" {
				if parsed, err := time.Parse("2006-01-02", frozenDateStr); err == nil {
					frozenDate = parsed
				}
			}

			count := 1
			if c, ok := itemMap["count"].(float64); ok && c > 1 {
				count = int(c)
			}

			for i := 0; i < count; i++ {
				res, err := db.DB.Exec("INSERT INTO items (name, category_id, freezer_id, weight, frozen_date) VALUES (?, ?, ?, ?, ?)",
					itemName, categoryID, freezerID, weight, frozenDate)
				if err != nil {
					return "", err
				}
				id, _ := res.LastInsertId()
				addedItems = append(addedItems, models.Item{
					ID:         id,
					Name:       itemName,
					CategoryID: categoryID,
					FreezerID:  freezerID,
					Weight:     weight,
					FrozenDate: frozenDate,
				})
			}
		}

		out, _ := json.Marshal(map[string]interface{}{"status": "success", "added_count": len(addedItems), "items": addedItems})
		return string(out), nil

	case "consume_items":
		var deleteIDs []int64

		if idsRaw, ok := args["item_ids"].([]interface{}); ok {
			for _, idVal := range idsRaw {
				if idFloat, ok := idVal.(float64); ok {
					deleteIDs = append(deleteIDs, int64(idFloat))
				}
			}
		}

		if len(deleteIDs) == 0 {
			itemName, _ := args["item_name"].(string)
			if itemName == "" {
				itemName, _ = args["name"].(string)
			}
			if itemName != "" {
				freezerName, _ := args["freezer_name"].(string)
				if freezerName == "" {
					freezerName, _ = args["freezer"].(string)
				}
				count := 1
				if c, ok := args["count"].(float64); ok && c > 0 {
					count = int(c)
				}

				query := "SELECT i.id FROM items i LEFT JOIN freezers f ON i.freezer_id = f.id WHERE LOWER(i.name) LIKE LOWER(?)"
				queryParams := []interface{}{"%" + strings.TrimSpace(itemName) + "%"}
				if freezerName != "" {
					query += " AND LOWER(f.name) = LOWER(?)"
					queryParams = append(queryParams, strings.TrimSpace(freezerName))
				}
				query += " ORDER BY i.id ASC LIMIT ?"
				queryParams = append(queryParams, count)

				rows, err := db.DB.Query(query, queryParams...)
				if err == nil {
					for rows.Next() {
						var id int64
						if rows.Scan(&id) == nil {
							deleteIDs = append(deleteIDs, id)
						}
					}
					rows.Close()
				}
			}
		}

		if len(deleteIDs) == 0 {
			return `{"status":"warning","message":"No matching items found to consume"}`, nil
		}

		for _, id := range deleteIDs {
			_, _ = db.DB.Exec("DELETE FROM items WHERE id = ?", id)
		}

		out, _ := json.Marshal(map[string]interface{}{"status": "success", "consumed_count": len(deleteIDs), "consumed_ids": deleteIDs})
		return string(out), nil

	case "move_items":
		targetFreezerID := int64(0)
		if fid, ok := args["target_freezer_id"].(float64); ok {
			targetFreezerID = int64(fid)
		} else if fName, ok := args["target_freezer_name"].(string); ok && fName != "" {
			_ = db.DB.QueryRow("SELECT id FROM freezers WHERE LOWER(name) = LOWER(?)", strings.TrimSpace(fName)).Scan(&targetFreezerID)
		}

		if targetFreezerID == 0 {
			return "", fmt.Errorf("target freezer not found")
		}

		var moveIDs []int64
		if idsRaw, ok := args["item_ids"].([]interface{}); ok {
			for _, idVal := range idsRaw {
				if idFloat, ok := idVal.(float64); ok {
					moveIDs = append(moveIDs, int64(idFloat))
				}
			}
		}

		if len(moveIDs) == 0 {
			itemName, _ := args["item_name"].(string)
			if itemName != "" {
				count := 999999
				if c, ok := args["count"].(float64); ok && c > 0 {
					count = int(c)
				}
				rows, err := db.DB.Query("SELECT id FROM items WHERE LOWER(name) LIKE LOWER(?) ORDER BY id ASC LIMIT ?", "%"+strings.TrimSpace(itemName)+"%", count)
				if err == nil {
					for rows.Next() {
						var id int64
						if rows.Scan(&id) == nil {
							moveIDs = append(moveIDs, id)
						}
					}
					rows.Close()
				}
			}
		}

		if len(moveIDs) == 0 {
			return `{"status":"warning","message":"No matching items found to move"}`, nil
		}

		for _, id := range moveIDs {
			_, _ = db.DB.Exec("UPDATE items SET freezer_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", targetFreezerID, id)
		}

		out, _ := json.Marshal(map[string]interface{}{"status": "success", "moved_count": len(moveIDs), "moved_ids": moveIDs, "target_freezer_id": targetFreezerID})
		return string(out), nil

	case "update_items_category":
		targetCategoryID := int64(0)
		if cid, ok := args["category_id"].(float64); ok {
			targetCategoryID = int64(cid)
		} else if cName, ok := args["category_name"].(string); ok && cName != "" {
			_ = db.DB.QueryRow("SELECT id FROM categories WHERE LOWER(name) = LOWER(?)", strings.TrimSpace(cName)).Scan(&targetCategoryID)
		}

		if targetCategoryID == 0 {
			return "", fmt.Errorf("target category not found")
		}

		var updateIDs []int64
		if idsRaw, ok := args["item_ids"].([]interface{}); ok {
			for _, idVal := range idsRaw {
				if idFloat, ok := idVal.(float64); ok {
					updateIDs = append(updateIDs, int64(idFloat))
				}
			}
		}

		if len(updateIDs) == 0 {
			itemName, _ := args["item_name"].(string)
			if itemName != "" {
				rows, err := db.DB.Query("SELECT id FROM items WHERE LOWER(name) LIKE LOWER(?)", "%"+strings.TrimSpace(itemName)+"%")
				if err == nil {
					for rows.Next() {
						var id int64
						if rows.Scan(&id) == nil {
							updateIDs = append(updateIDs, id)
						}
					}
					rows.Close()
				}
			}
		}

		if len(updateIDs) == 0 {
			return `{"status":"warning","message":"No matching items found to update category"}`, nil
		}

		for _, id := range updateIDs {
			_, _ = db.DB.Exec("UPDATE items SET category_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", targetCategoryID, id)
		}

		out, _ := json.Marshal(map[string]interface{}{"status": "success", "updated_count": len(updateIDs), "updated_ids": updateIDs, "target_category_id": targetCategoryID})
		return string(out), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// HTTP Handler for /api/mcp
func HandleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	res := RPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		res.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "freezo-mcp",
				"version": "1.0.0",
			},
		}

	case "notifications/initialized":
		res.Result = map[string]interface{}{}

	case "tools/list":
		res.Result = map[string]interface{}{
			"tools": GetToolDefinitions(),
		}

	case "tools/call":
		var callParams ToolCallParams
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			res.Error = &RPCError{Code: -32602, Message: "Invalid params"}
		} else {
			output, err := ExecuteTool(callParams.Name, callParams.Arguments)
			if err != nil {
				res.Error = &RPCError{Code: -32603, Message: err.Error()}
			} else {
				res.Result = map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": output,
						},
					},
				}
			}
		}

	default:
		res.Error = &RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func StrToInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}
