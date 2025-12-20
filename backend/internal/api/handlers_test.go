package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/riccardo/freezo/backend/internal/db"
	"github.com/riccardo/freezo/backend/internal/models"
)

func setupTestDB() {
	// Use in-memory SQLite for testing
	db.InitDB(":memory:")
}

func executeRequest(req *http.Request, r *chi.Mux) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func TestFreezerLifecycle(t *testing.T) {
	setupTestDB()
	r := NewRouter()

	// 1. Create Freezer
	freezer := models.Freezer{Name: "Test Freezer"}
	payload, _ := json.Marshal(freezer)
	req, _ := http.NewRequest("POST", "/api/freezers", bytes.NewBuffer(payload))
	resp := executeRequest(req, r)

	if resp.Code != http.StatusCreated {
		t.Errorf("Expected StatusCreated, got %d", resp.Code)
	}

	var createdFreezer models.Freezer
	json.Unmarshal(resp.Body.Bytes(), &createdFreezer)
	freezerID := createdFreezer.ID

	// 2. Create Item in Freezer
	item := models.Item{Name: "Ice Cream", FreezerID: freezerID}
	payload, _ = json.Marshal(item)
	req, _ = http.NewRequest("POST", "/api/items", bytes.NewBuffer(payload))
	resp = executeRequest(req, r)

	if resp.Code != http.StatusCreated {
		t.Errorf("Expected StatusCreated, got %d", resp.Code)
	}

	var createdItem models.Item
	json.Unmarshal(resp.Body.Bytes(), &createdItem)
	itemID := createdItem.ID

	// 3. Try to Delete Freezer (Should Fail)
	// We need to format the ID as string
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/freezers/%d", freezerID), nil)
	resp = executeRequest(req, r)

	if resp.Code != http.StatusConflict {
		t.Errorf("Expected StatusConflict (409) when deleting non-empty freezer, got %d", resp.Code)
	}

	// 4. Delete Item
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/items/%d", itemID), nil)
	resp = executeRequest(req, r)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected StatusOK when deleting item, got %d", resp.Code)
	}

	// 5. Delete Freezer (Should Succeed)
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/freezers/%d", freezerID), nil)
	resp = executeRequest(req, r)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected StatusOK when deleting empty freezer, got %d", resp.Code)
	}
}
func TestMoveItems(t *testing.T) {
	setupTestDB()
	defer db.DB.Close()

	// 1. Create 2 Freezers
	f1Req := `{"name": "Freezer 1"}`
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(f1Req))
	w := httptest.NewRecorder()
	CreateFreezer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create freezer 1: %d", w.Code)
	}
	var f1 models.Freezer
	json.NewDecoder(w.Body).Decode(&f1)

	f2Req := `{"name": "Freezer 2"}`
	req = httptest.NewRequest("POST", "/", bytes.NewBufferString(f2Req))
	w = httptest.NewRecorder()
	CreateFreezer(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create freezer 2: %d", w.Code)
	}
	var f2 models.Freezer
	json.NewDecoder(w.Body).Decode(&f2)

	// 2. Create 2 Items in Freezer 1
	// item 1
	i1Req := fmt.Sprintf(`{"name": "Item 1", "quantity": 1, "freezer_id": %d}`, f1.ID)
	req = httptest.NewRequest("POST", "/items/", bytes.NewBufferString(i1Req))
	w = httptest.NewRecorder()
	CreateItem(w, req)
	var i1 models.Item
	json.NewDecoder(w.Body).Decode(&i1)

	// item 2
	i2Req := fmt.Sprintf(`{"name": "Item 2", "quantity": 1, "freezer_id": %d}`, f1.ID)
	req = httptest.NewRequest("POST", "/items/", bytes.NewBufferString(i2Req))
	w = httptest.NewRecorder()
	CreateItem(w, req)
	var i2 models.Item
	json.NewDecoder(w.Body).Decode(&i2)

	// 3. Move Items to Freezer 2
	moveReqStruct := MoveItemsRequest{
		ItemIDs:      []int64{i1.ID, i2.ID},
		NewFreezerID: f2.ID,
	}
	moveBody, _ := json.Marshal(moveReqStruct)
	req = httptest.NewRequest("POST", "/items/move", bytes.NewReader(moveBody))
	w = httptest.NewRecorder()
	MoveItems(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to move items: %d - %s", w.Code, w.Body.String())
	}

	// 4. Verify Items are in Freezer 2
	req = httptest.NewRequest("GET", "/items/", nil)
	w = httptest.NewRecorder()
	GetItems(w, req)

	var items []models.Item
	json.NewDecoder(w.Body).Decode(&items)

	countF2 := 0
	for _, item := range items {
		if item.FreezerID == f2.ID {
			countF2++
		}
	}

	if countF2 != 2 {
		t.Errorf("Expected 2 items in Freezer 2, got %d", countF2)
	}
}
