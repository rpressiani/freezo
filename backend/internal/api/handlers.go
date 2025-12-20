package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/riccardo/freezo/backend/internal/db"
	"github.com/riccardo/freezo/backend/internal/models"
)

// Helper to respond with JSON
func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// --- Freezers ---

func GetFreezers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name, created_at, updated_at FROM freezers")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	freezers := []models.Freezer{}
	for rows.Next() {
		var f models.Freezer
		if err := rows.Scan(&f.ID, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		freezers = append(freezers, f)
	}
	jsonResponse(w, http.StatusOK, freezers)
}

func CreateFreezer(w http.ResponseWriter, r *http.Request) {
	var f models.Freezer
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.DB.Exec("INSERT INTO freezers (name) VALUES (?)", f.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	f.ID = id
	jsonResponse(w, http.StatusCreated, f)
}

func DeleteFreezer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	// Check if freezer has items
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM items WHERE freezer_id = ?", id).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Cannot delete freezer with items", http.StatusConflict)
		return
	}

	_, err = db.DB.Exec("DELETE FROM freezers WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// --- Items ---

func GetItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name, category_id, freezer_id, weight, frozen_date, created_at, updated_at FROM items")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []models.Item{}
	for rows.Next() {
		var i models.Item
		var frozenDate sql.NullTime
		var weight sql.NullString
		if err := rows.Scan(&i.ID, &i.Name, &i.CategoryID, &i.FreezerID, &weight, &frozenDate, &i.CreatedAt, &i.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if frozenDate.Valid {
			i.FrozenDate = frozenDate.Time
		}
		if weight.Valid {
			i.Weight = weight.String
		}
		items = append(items, i)
	}
	jsonResponse(w, http.StatusOK, items)
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	var i models.Item
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec("INSERT INTO items (name, category_id, freezer_id, weight, frozen_date) VALUES (?, ?, ?, ?, ?)",
		i.Name, i.CategoryID, i.FreezerID, i.Weight, i.FrozenDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	i.ID = id
	jsonResponse(w, http.StatusCreated, i)
}

func CreateItemsBatch(w http.ResponseWriter, r *http.Request) {
	var items []models.Item
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO items (name, category_id, freezer_id, weight, frozen_date) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for i := range items {
		res, err := stmt.Exec(items[i].Name, items[i].CategoryID, items[i].FreezerID, items[i].Weight, items[i].FrozenDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		items[i].ID = id
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusCreated, items)
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	_, err := db.DB.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	var i models.Item
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("UPDATE items SET name=?, category_id=?, freezer_id=?, weight=?, frozen_date=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		i.Name, i.CategoryID, i.FreezerID, i.Weight, i.FrozenDate, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	i.ID = int64(id)
	jsonResponse(w, http.StatusOK, i)
}

type BatchConsumeRequest struct {
	DeleteIDs []int64 `json:"delete_ids"`
}

func ConsumeItemsBatch(w http.ResponseWriter, r *http.Request) {
	var req BatchConsumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Handle Deletions
	if len(req.DeleteIDs) > 0 {
		delStmt, err := tx.Prepare("DELETE FROM items WHERE id = ?")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer delStmt.Close()

		for _, id := range req.DeleteIDs {
			if _, err := delStmt.Exec(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "batch consumed"})
}

type MoveItemsRequest struct {
	ItemIDs      []int64 `json:"item_ids"`
	NewFreezerID int64   `json:"new_freezer_id"`
}

func MoveItems(w http.ResponseWriter, r *http.Request) {
	var req MoveItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.ItemIDs) == 0 {
		http.Error(w, "No items to move", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update freezer_id for all items
	stmt, err := tx.Prepare("UPDATE items SET freezer_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for _, id := range req.ItemIDs {
		if _, err := stmt.Exec(req.NewFreezerID, id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "items moved"})
}

// --- Categories ---

func GetCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, name FROM categories")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		categories = append(categories, c)
	}
	jsonResponse(w, http.StatusOK, categories)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var c models.Category
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.DB.Exec("INSERT INTO categories (name) VALUES (?)", c.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	c.ID = id
	jsonResponse(w, http.StatusCreated, c)
}
