package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/riccardo/freezo/backend/internal/db"
	"github.com/riccardo/freezo/backend/internal/models"
)

func setupTestDB() {
	// Create a temp file for testing
	f, _ := os.CreateTemp("", "freezer-test-*.db")
	f.Close()
	db.InitDB(f.Name())
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

func TestExportDatabase(t *testing.T) {
	setupTestDB()
	r := NewRouter()

	req, _ := http.NewRequest("GET", "/api/database/export", nil)
	resp := executeRequest(req, r)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected StatusOK, got %d", resp.Code)
	}

	contentDisposition := resp.Header().Get("Content-Disposition")
	if contentDisposition != "attachment; filename=freezer.db" {
		t.Errorf("Expected attachment header, got %s", contentDisposition)
	}
}

func TestResetDatabase(t *testing.T) {
	setupTestDB()
	r := NewRouter()

	// 1. Add some data
	freezer := models.Freezer{Name: "Test Freezer"}
	payload, _ := json.Marshal(freezer)
	reqFreezer, _ := http.NewRequest("POST", "/api/freezers", bytes.NewBuffer(payload))
	executeRequest(reqFreezer, r)

	// 2. Verify data exists
	reqCheck, _ := http.NewRequest("GET", "/api/freezers", nil)
	respCheck := executeRequest(reqCheck, r)
	var freezers []models.Freezer
	json.Unmarshal(respCheck.Body.Bytes(), &freezers)
	if len(freezers) != 1 {
		t.Errorf("Expected 1 freezer before reset, got %d", len(freezers))
	}

	// 3. Reset
	reqReset, _ := http.NewRequest("POST", "/api/database/reset", nil)
	respReset := executeRequest(reqReset, r)
	if respReset.Code != http.StatusOK {
		t.Errorf("Expected StatusOK after reset, got %d", respReset.Code)
	}

	// 4. Verify data is gone
	reqCheck2, _ := http.NewRequest("GET", "/api/freezers", nil)
	respCheck2 := executeRequest(reqCheck2, r)
	var freezersAfter []models.Freezer
	json.Unmarshal(respCheck2.Body.Bytes(), &freezersAfter)
	if len(freezersAfter) != 0 {
		t.Errorf("Expected 0 freezers after reset, got %d", len(freezersAfter))
	}
}

func TestImportDatabase(t *testing.T) {
	setupTestDB()
	r := NewRouter()

	// 1. Add some data
	freezer := models.Freezer{Name: "Backup Freezer"}
	payload, _ := json.Marshal(freezer)
	reqFreezer, _ := http.NewRequest("POST", "/api/freezers", bytes.NewBuffer(payload))
	executeRequest(reqFreezer, r)

	// 2. Export database
	reqExport, _ := http.NewRequest("GET", "/api/database/export", nil)
	respExport := executeRequest(reqExport, r)
	if respExport.Code != http.StatusOK {
		t.Fatalf("Expected StatusOK on export, got %d", respExport.Code)
	}
	backupData := respExport.Body.Bytes()

	// 3. Reset Database
	reqReset, _ := http.NewRequest("POST", "/api/database/reset", nil)
	executeRequest(reqReset, r)

	// Verify data is gone
	reqCheck, _ := http.NewRequest("GET", "/api/freezers", nil)
	respCheck := executeRequest(reqCheck, r)
	var freezersAfterReset []models.Freezer
	json.Unmarshal(respCheck.Body.Bytes(), &freezersAfterReset)
	if len(freezersAfterReset) != 0 {
		t.Fatalf("Expected 0 freezers after reset, got %d", len(freezersAfterReset))
	}

	// 4. Import Database
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("database", "freezer.db")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(backupData)
	writer.Close()

	reqImport, _ := http.NewRequest("POST", "/api/database/import", body)
	reqImport.Header.Set("Content-Type", writer.FormDataContentType())
	respImport := executeRequest(reqImport, r)

	if respImport.Code != http.StatusOK {
		t.Fatalf("Expected StatusOK on import, got %d - %s", respImport.Code, respImport.Body.String())
	}

	// 5. Verify data is back
	reqCheckFinal, _ := http.NewRequest("GET", "/api/freezers", nil)
	respCheckFinal := executeRequest(reqCheckFinal, r)
	var freezersFinal []models.Freezer
	json.Unmarshal(respCheckFinal.Body.Bytes(), &freezersFinal)
	if len(freezersFinal) != 1 || freezersFinal[0].Name != "Backup Freezer" {
		t.Errorf("Expected restored freezer 'Backup Freezer', got %v", freezersFinal)
	}
}

func TestUpdateItemsCategory(t *testing.T) {
	setupTestDB()
	r := NewRouter()

	// 1. Create Freezer
	freezer := models.Freezer{Name: "Cat Freezer"}
	payload, _ := json.Marshal(freezer)
	req, _ := http.NewRequest("POST", "/api/freezers", bytes.NewBuffer(payload))
	resp := executeRequest(req, r)
	var createdFreezer models.Freezer
	json.Unmarshal(resp.Body.Bytes(), &createdFreezer)

	// 2. Create Item with Category 1
	item := models.Item{Name: "Pork Chop", FreezerID: createdFreezer.ID, CategoryID: 1}
	payload, _ = json.Marshal(item)
	req, _ = http.NewRequest("POST", "/api/items", bytes.NewBuffer(payload))
	resp = executeRequest(req, r)
	var createdItem models.Item
	json.Unmarshal(resp.Body.Bytes(), &createdItem)

	// 3. Update Category to 2 using PUT /api/items/{id}
	createdItem.CategoryID = 2
	updatePayload, _ := json.Marshal(createdItem)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/items/%d", createdItem.ID), bytes.NewBuffer(updatePayload))
	resp = executeRequest(req, r)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected StatusOK (200) when updating item via PUT, got %d", resp.Code)
	}

	// 4. Verify category updated
	req, _ = http.NewRequest("GET", "/api/items", nil)
	resp = executeRequest(req, r)
	var items []models.Item
	json.Unmarshal(resp.Body.Bytes(), &items)
	if len(items) != 1 || items[0].CategoryID != 2 {
		t.Errorf("Expected item category to be 2, got %v", items)
	}
}

func TestCategoryDeleteProtection(t *testing.T) {
	setupTestDB()
	r := NewRouter()

	// 1. Create Category
	catPayload, _ := json.Marshal(models.Category{Name: "Snacks"})
	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(catPayload))
	resp := executeRequest(req, r)
	if resp.Code != http.StatusCreated {
		t.Fatalf("Expected StatusCreated, got %d", resp.Code)
	}
	var createdCat models.Category
	json.Unmarshal(resp.Body.Bytes(), &createdCat)

	// 2. Create Freezer & Item assigned to this Category
	freezerPayload, _ := json.Marshal(models.Freezer{Name: "Cat Test Freezer"})
	req, _ = http.NewRequest("POST", "/api/freezers", bytes.NewBuffer(freezerPayload))
	resp = executeRequest(req, r)
	var createdFreezer models.Freezer
	json.Unmarshal(resp.Body.Bytes(), &createdFreezer)

	itemPayload, _ := json.Marshal(models.Item{Name: "Chips", FreezerID: createdFreezer.ID, CategoryID: createdCat.ID})
	req, _ = http.NewRequest("POST", "/api/items", bytes.NewBuffer(itemPayload))
	resp = executeRequest(req, r)
	var createdItem models.Item
	json.Unmarshal(resp.Body.Bytes(), &createdItem)

	// 3. Try deleting Category with items (Should fail with 409 Conflict)
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/categories/%d", createdCat.ID), nil)
	resp = executeRequest(req, r)
	if resp.Code != http.StatusConflict {
		t.Errorf("Expected StatusConflict (409) when deleting category with items, got %d", resp.Code)
	}

	// 4. Delete the Item first
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/items/%d", createdItem.ID), nil)
	resp = executeRequest(req, r)

	// 5. Delete Category now (Should succeed with 200 OK)
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/categories/%d", createdCat.ID), nil)
	resp = executeRequest(req, r)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected StatusOK (200) when deleting empty category, got %d", resp.Code)
	}

	// 6. Try deleting default Uncategorized category (Should fail with 403 Forbidden)
	req, _ = http.NewRequest("GET", "/api/categories", nil)
	resp = executeRequest(req, r)
	var cats []models.Category
	json.Unmarshal(resp.Body.Bytes(), &cats)
	var uncatID int64
	for _, c := range cats {
		if c.Name == "Uncategorized" {
			uncatID = c.ID
			break
		}
	}
	if uncatID > 0 {
		req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/categories/%d", uncatID), nil)
		resp = executeRequest(req, r)
		if resp.Code != http.StatusForbidden {
			t.Errorf("Expected StatusForbidden (403) when deleting Uncategorized category, got %d", resp.Code)
		}
	}
}


