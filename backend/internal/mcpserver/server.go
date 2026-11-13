package mcpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/riccardo/freezo/backend/internal/db"
	"github.com/riccardo/freezo/backend/internal/models"
)

const serverVersion = "0.1.0"

// EnrichedItem represents a freezer item with resolved freezer and category names.
type EnrichedItem struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	CategoryID   int64      `json:"category_id"`
	CategoryName string     `json:"category_name"`
	FreezerID    int64      `json:"freezer_id"`
	FreezerName  string     `json:"freezer_name"`
	Weight       string     `json:"weight,omitempty"`
	FrozenDate   *time.Time `json:"frozen_date,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// NewServer constructs the Freezo MCP server with all tools and resources registered.
func NewServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "freezo",
		Version: serverVersion,
	}, nil)

	registerResources(s)
	registerTools(s)
	return s
}

// NewHTTPHandler returns a Streamable HTTP handler for the Freezo MCP server.
func NewHTTPHandler() http.Handler {
	srv := NewServer()
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, nil)
}

func registerResources(s *mcp.Server) {
	s.AddResource(&mcp.Resource{
		URI:         "freezo://inventory",
		Name:        "Freezer Inventory",
		Description: "All items currently stored across all freezers, with resolved freezer and category names.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		items, err := queryEnrichedItems(ctx, "", "", "", 0)
		if err != nil {
			return nil, err
		}
		return jsonResourceResult(req.Params.URI, items)
	})

	s.AddResource(&mcp.Resource{
		URI:         "freezo://freezers",
		Name:        "Freezers",
		Description: "List of all configured freezers.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		freezers, err := queryFreezers(ctx)
		if err != nil {
			return nil, err
		}
		return jsonResourceResult(req.Params.URI, freezers)
	})

	s.AddResource(&mcp.Resource{
		URI:         "freezo://categories",
		Name:        "Categories",
		Description: "List of all food categories.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		categories, err := queryCategories(ctx)
		if err != nil {
			return nil, err
		}
		return jsonResourceResult(req.Params.URI, categories)
	})
}

func jsonResourceResult(uri string, payload any) (*mcp.ReadResourceResult, error) {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      uri,
				MIMEType: "application/json",
				Text:     string(data),
			},
		},
	}, nil
}

// --- Tool Input/Output Types ---

type emptyInput struct{}

type listInventoryInput struct {
	Freezer       string `json:"freezer,omitempty" jsonschema:"Filter by freezer ID or name (case-insensitive)"`
	Category      string `json:"category,omitempty" jsonschema:"Filter by category ID or name (case-insensitive)"`
	Search        string `json:"search,omitempty" jsonschema:"Filter items whose name contains this substring (case-insensitive)"`
	OlderThanDays int    `json:"older_than_days,omitempty" jsonschema:"Only return items frozen at least this many days ago (excludes unknown dates)"`
}

type listInventoryOutput struct {
	Count int            `json:"count"`
	Items []EnrichedItem `json:"items"`
}

type namedCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type inventorySummaryOutput struct {
	TotalItems       int            `json:"total_items"`
	UnknownDateCount int            `json:"unknown_date_count"`
	ByFreezer        []namedCount   `json:"by_freezer"`
	ByCategory       []namedCount   `json:"by_category"`
	OldestItems      []EnrichedItem `json:"oldest_items"`
}

type listFreezersOutput struct {
	Freezers []models.Freezer `json:"freezers"`
}

type createFreezerInput struct {
	Name string `json:"name" jsonschema:"Name of the new freezer (e.g. 'Garage Deep Freeze')"`
}

type deleteByIDInput struct {
	ID int64 `json:"id" jsonschema:"ID of the record to delete"`
}

type messageOutput struct {
	Message string `json:"message"`
}

type listCategoriesOutput struct {
	Categories []models.Category `json:"categories"`
}

type createCategoryInput struct {
	Name string `json:"name" jsonschema:"Name of the new category"`
	Icon string `json:"icon,omitempty" jsonschema:"Optional Lucide icon name (e.g. 'fish', 'beef', 'tag')"`
}

type addItemsInput struct {
	Name        string `json:"name" jsonschema:"Item name (e.g. 'Chicken Breast')"`
	Freezer     string `json:"freezer" jsonschema:"Target freezer ID or name"`
	Category    string `json:"category,omitempty" jsonschema:"Category ID or name (defaults to 'Uncategorized')"`
	Quantity    int    `json:"quantity,omitempty" jsonschema:"Number of identical items to add (defaults to 1)"`
	Weight      string `json:"weight,omitempty" jsonschema:"Optional weight or size per item (e.g. '500g', '1 lb')"`
	FrozenDate  string `json:"frozen_date,omitempty" jsonschema:"Optional frozen date in YYYY-MM-DD format (defaults to today)"`
	UnknownDate bool   `json:"unknown_date,omitempty" jsonschema:"Set true if the frozen date is unknown"`
}

type addItemsOutput struct {
	AddedCount int           `json:"added_count"`
	Items      []models.Item `json:"items"`
}

type consumeItemsInput struct {
	ItemIDs  []int64 `json:"item_ids,omitempty" jsonschema:"Specific item IDs to consume/remove"`
	Name     string  `json:"name,omitempty" jsonschema:"Item name to consume when item_ids is not provided (FIFO: unknown dates first, then oldest)"`
	Freezer  string  `json:"freezer,omitempty" jsonschema:"Optional freezer ID or name to scope consumption by name"`
	Quantity int     `json:"quantity,omitempty" jsonschema:"Number of items to consume when consuming by name (defaults to 1)"`
}

type consumeItemsOutput struct {
	ConsumedCount int     `json:"consumed_count"`
	ConsumedIDs   []int64 `json:"consumed_ids"`
}

type moveItemsInput struct {
	ToFreezer   string  `json:"to_freezer" jsonschema:"Destination freezer ID or name"`
	ItemIDs     []int64 `json:"item_ids,omitempty" jsonschema:"Specific item IDs to move"`
	Name        string  `json:"name,omitempty" jsonschema:"Item name to move when item_ids is not provided"`
	FromFreezer string  `json:"from_freezer,omitempty" jsonschema:"Optional source freezer ID or name when moving by name"`
	Quantity    int     `json:"quantity,omitempty" jsonschema:"Number of items to move when moving by name (defaults to 1)"`
}

type moveItemsOutput struct {
	MovedCount int     `json:"moved_count"`
	MovedIDs   []int64 `json:"moved_ids"`
	ToFreezer  string  `json:"to_freezer"`
}

type updateItemInput struct {
	ID          int64   `json:"id" jsonschema:"ID of the item to update"`
	Name        *string `json:"name,omitempty" jsonschema:"New item name"`
	Freezer     *string `json:"freezer,omitempty" jsonschema:"New freezer ID or name"`
	Category    *string `json:"category,omitempty" jsonschema:"New category ID or name"`
	Weight      *string `json:"weight,omitempty" jsonschema:"New weight string"`
	FrozenDate  *string `json:"frozen_date,omitempty" jsonschema:"New frozen date in YYYY-MM-DD format"`
	UnknownDate bool    `json:"unknown_date,omitempty" jsonschema:"Set true to clear the frozen date"`
}

func registerTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_inventory",
		Description: "List and filter items currently in freezers, with resolved freezer and category names.",
	}, handleListInventory)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_inventory_summary",
		Description: "Get an overview of freezer inventory counts by freezer and category, plus the oldest items to prioritize.",
	}, handleGetInventorySummary)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_freezers",
		Description: "List all configured freezers.",
	}, handleListFreezers)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_freezer",
		Description: "Create a new freezer location.",
	}, handleCreateFreezer)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_freezer",
		Description: "Delete an empty freezer by ID. Fails if the freezer still contains items.",
	}, handleDeleteFreezer)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_categories",
		Description: "List all food categories.",
	}, handleListCategories)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_category",
		Description: "Create a new food category.",
	}, handleCreateCategory)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_category",
		Description: "Delete an unused category by ID. Cannot delete 'Uncategorized' or categories with items.",
	}, handleDeleteCategory)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_items",
		Description: "Add one or more identical food items to a freezer. Resolves freezer and category by ID or name.",
	}, handleAddItems)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "consume_items",
		Description: "Consume (delete) items either by explicit item_ids or by name/freezer/quantity (consumes unknown-date and oldest items first).",
	}, handleConsumeItems)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "move_items",
		Description: "Move items to another freezer either by explicit item_ids or by name/from_freezer/quantity.",
	}, handleMoveItems)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_item",
		Description: "Update an existing item's name, freezer, category, weight, or frozen date.",
	}, handleUpdateItem)
}

// --- Tool Handlers ---

func handleListInventory(ctx context.Context, _ *mcp.CallToolRequest, in listInventoryInput) (*mcp.CallToolResult, listInventoryOutput, error) {
	items, err := queryEnrichedItems(ctx, in.Freezer, in.Category, in.Search, in.OlderThanDays)
	if err != nil {
		return nil, listInventoryOutput{}, err
	}
	return nil, listInventoryOutput{Count: len(items), Items: items}, nil
}

func handleGetInventorySummary(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, inventorySummaryOutput, error) {
	items, err := queryEnrichedItems(ctx, "", "", "", 0)
	if err != nil {
		return nil, inventorySummaryOutput{}, err
	}

	freezerCounts := map[string]int{}
	categoryCounts := map[string]int{}
	unknownCount := 0

	for _, it := range items {
		freezerCounts[it.FreezerName]++
		categoryCounts[it.CategoryName]++
		if it.FrozenDate == nil {
			unknownCount++
		}
	}

	out := inventorySummaryOutput{
		TotalItems:       len(items),
		UnknownDateCount: unknownCount,
		ByFreezer:        mapToNamedCounts(freezerCounts),
		ByCategory:       mapToNamedCounts(categoryCounts),
	}

	// queryEnrichedItems orders unknown dates first, then oldest frozen_date first.
	limit := 10
	if len(items) < limit {
		limit = len(items)
	}
	out.OldestItems = items[:limit]
	return nil, out, nil
}

func handleListFreezers(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, listFreezersOutput, error) {
	freezers, err := queryFreezers(ctx)
	if err != nil {
		return nil, listFreezersOutput{}, err
	}
	return nil, listFreezersOutput{Freezers: freezers}, nil
}

func handleCreateFreezer(ctx context.Context, _ *mcp.CallToolRequest, in createFreezerInput) (*mcp.CallToolResult, models.Freezer, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, models.Freezer{}, fmt.Errorf("freezer name is required")
	}

	res, err := db.DB.ExecContext(ctx, "INSERT INTO freezers (name) VALUES (?)", name)
	if err != nil {
		return nil, models.Freezer{}, err
	}
	id, _ := res.LastInsertId()

	var f models.Freezer
	err = db.DB.QueryRowContext(ctx, "SELECT id, name, created_at, updated_at FROM freezers WHERE id = ?", id).
		Scan(&f.ID, &f.Name, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, models.Freezer{}, err
	}
	return nil, f, nil
}

func handleDeleteFreezer(ctx context.Context, _ *mcp.CallToolRequest, in deleteByIDInput) (*mcp.CallToolResult, messageOutput, error) {
	var count int
	if err := db.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM items WHERE freezer_id = ?", in.ID).Scan(&count); err != nil {
		return nil, messageOutput{}, err
	}
	if count > 0 {
		return nil, messageOutput{}, fmt.Errorf("cannot delete freezer %d: still contains %d item(s)", in.ID, count)
	}
	res, err := db.DB.ExecContext(ctx, "DELETE FROM freezers WHERE id = ?", in.ID)
	if err != nil {
		return nil, messageOutput{}, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, messageOutput{}, fmt.Errorf("freezer %d not found", in.ID)
	}
	return nil, messageOutput{Message: fmt.Sprintf("deleted freezer %d", in.ID)}, nil
}

func handleListCategories(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, listCategoriesOutput, error) {
	categories, err := queryCategories(ctx)
	if err != nil {
		return nil, listCategoriesOutput{}, err
	}
	return nil, listCategoriesOutput{Categories: categories}, nil
}

func handleCreateCategory(ctx context.Context, _ *mcp.CallToolRequest, in createCategoryInput) (*mcp.CallToolResult, models.Category, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, models.Category{}, fmt.Errorf("category name is required")
	}
	res, err := db.DB.ExecContext(ctx, "INSERT INTO categories (name, icon) VALUES (?, ?)", name, strings.TrimSpace(in.Icon))
	if err != nil {
		return nil, models.Category{}, err
	}
	id, _ := res.LastInsertId()
	return nil, models.Category{ID: id, Name: name, Icon: strings.TrimSpace(in.Icon)}, nil
}

func handleDeleteCategory(ctx context.Context, _ *mcp.CallToolRequest, in deleteByIDInput) (*mcp.CallToolResult, messageOutput, error) {
	var name string
	err := db.DB.QueryRowContext(ctx, "SELECT name FROM categories WHERE id = ?", in.ID).Scan(&name)
	if err == sql.ErrNoRows {
		return nil, messageOutput{}, fmt.Errorf("category %d not found", in.ID)
	}
	if err != nil {
		return nil, messageOutput{}, err
	}
	if strings.EqualFold(name, "uncategorized") {
		return nil, messageOutput{}, fmt.Errorf("cannot delete the default Uncategorized category")
	}

	var count int
	if err := db.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM items WHERE category_id = ?", in.ID).Scan(&count); err != nil {
		return nil, messageOutput{}, err
	}
	if count > 0 {
		return nil, messageOutput{}, fmt.Errorf("cannot delete category %d (%s): %d item(s) assigned", in.ID, name, count)
	}

	if _, err := db.DB.ExecContext(ctx, "DELETE FROM categories WHERE id = ?", in.ID); err != nil {
		return nil, messageOutput{}, err
	}
	return nil, messageOutput{Message: fmt.Sprintf("deleted category %d (%s)", in.ID, name)}, nil
}

func handleAddItems(ctx context.Context, _ *mcp.CallToolRequest, in addItemsInput) (*mcp.CallToolResult, addItemsOutput, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, addItemsOutput{}, fmt.Errorf("item name is required")
	}
	qty := in.Quantity
	if qty <= 0 {
		qty = 1
	}

	freezerID, _, err := resolveFreezer(ctx, in.Freezer)
	if err != nil {
		return nil, addItemsOutput{}, err
	}

	catRef := strings.TrimSpace(in.Category)
	if catRef == "" {
		catRef = "Uncategorized"
	}
	categoryID, _, err := resolveCategory(ctx, catRef)
	if err != nil {
		return nil, addItemsOutput{}, err
	}

	var frozenDate *time.Time
	if !in.UnknownDate {
		if strings.TrimSpace(in.FrozenDate) != "" {
			parsed, err := parseDate(in.FrozenDate)
			if err != nil {
				return nil, addItemsOutput{}, err
			}
			frozenDate = &parsed
		} else {
			now := time.Now().UTC()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
			frozenDate = &today
		}
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, addItemsOutput{}, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO items (name, category_id, freezer_id, weight, frozen_date) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return nil, addItemsOutput{}, err
	}
	defer stmt.Close()

	created := make([]models.Item, 0, qty)
	for i := 0; i < qty; i++ {
		res, err := stmt.ExecContext(ctx, name, categoryID, freezerID, in.Weight, frozenDate)
		if err != nil {
			return nil, addItemsOutput{}, err
		}
		id, _ := res.LastInsertId()
		created = append(created, models.Item{
			ID:         id,
			Name:       name,
			CategoryID: categoryID,
			FreezerID:  freezerID,
			Weight:     in.Weight,
			FrozenDate: frozenDate,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, addItemsOutput{}, err
	}

	return nil, addItemsOutput{AddedCount: len(created), Items: created}, nil
}

func handleConsumeItems(ctx context.Context, _ *mcp.CallToolRequest, in consumeItemsInput) (*mcp.CallToolResult, consumeItemsOutput, error) {
	ids := in.ItemIDs
	if len(ids) == 0 {
		if strings.TrimSpace(in.Name) == "" {
			return nil, consumeItemsOutput{}, fmt.Errorf("provide either item_ids or name")
		}
		qty := in.Quantity
		if qty <= 0 {
			qty = 1
		}
		var err error
		ids, err = findItemIDsFIFO(ctx, in.Name, in.Freezer, qty)
		if err != nil {
			return nil, consumeItemsOutput{}, err
		}
		if len(ids) == 0 {
			return nil, consumeItemsOutput{}, fmt.Errorf("no matching items found for %q", in.Name)
		}
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, consumeItemsOutput{}, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "DELETE FROM items WHERE id = ?")
	if err != nil {
		return nil, consumeItemsOutput{}, err
	}
	defer stmt.Close()

	consumed := make([]int64, 0, len(ids))
	for _, id := range ids {
		res, err := stmt.ExecContext(ctx, id)
		if err != nil {
			return nil, consumeItemsOutput{}, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			consumed = append(consumed, id)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, consumeItemsOutput{}, err
	}
	return nil, consumeItemsOutput{ConsumedCount: len(consumed), ConsumedIDs: consumed}, nil
}

func handleMoveItems(ctx context.Context, _ *mcp.CallToolRequest, in moveItemsInput) (*mcp.CallToolResult, moveItemsOutput, error) {
	toFreezerID, toFreezerName, err := resolveFreezer(ctx, in.ToFreezer)
	if err != nil {
		return nil, moveItemsOutput{}, err
	}

	ids := in.ItemIDs
	if len(ids) == 0 {
		if strings.TrimSpace(in.Name) == "" {
			return nil, moveItemsOutput{}, fmt.Errorf("provide either item_ids or name")
		}
		qty := in.Quantity
		if qty <= 0 {
			qty = 1
		}
		ids, err = findItemIDsFIFO(ctx, in.Name, in.FromFreezer, qty)
		if err != nil {
			return nil, moveItemsOutput{}, err
		}
		if len(ids) == 0 {
			return nil, moveItemsOutput{}, fmt.Errorf("no matching items found for %q", in.Name)
		}
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, moveItemsOutput{}, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "UPDATE items SET freezer_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
	if err != nil {
		return nil, moveItemsOutput{}, err
	}
	defer stmt.Close()

	moved := make([]int64, 0, len(ids))
	for _, id := range ids {
		res, err := stmt.ExecContext(ctx, toFreezerID, id)
		if err != nil {
			return nil, moveItemsOutput{}, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			moved = append(moved, id)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, moveItemsOutput{}, err
	}
	return nil, moveItemsOutput{
		MovedCount: len(moved),
		MovedIDs:   moved,
		ToFreezer:  toFreezerName,
	}, nil
}

func handleUpdateItem(ctx context.Context, _ *mcp.CallToolRequest, in updateItemInput) (*mcp.CallToolResult, models.Item, error) {
	var current models.Item
	var frozenDate sql.NullTime
	var weight sql.NullString

	err := db.DB.QueryRowContext(ctx,
		"SELECT id, name, category_id, freezer_id, weight, frozen_date, created_at, updated_at FROM items WHERE id = ?",
		in.ID,
	).Scan(&current.ID, &current.Name, &current.CategoryID, &current.FreezerID, &weight, &frozenDate, &current.CreatedAt, &current.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, models.Item{}, fmt.Errorf("item %d not found", in.ID)
	}
	if err != nil {
		return nil, models.Item{}, err
	}
	if weight.Valid {
		current.Weight = weight.String
	}
	if frozenDate.Valid && !frozenDate.Time.IsZero() && frozenDate.Time.Year() > 1 {
		current.FrozenDate = &frozenDate.Time
	}

	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		current.Name = strings.TrimSpace(*in.Name)
	}
	if in.Weight != nil {
		current.Weight = *in.Weight
	}
	if in.Freezer != nil {
		fid, _, err := resolveFreezer(ctx, *in.Freezer)
		if err != nil {
			return nil, models.Item{}, err
		}
		current.FreezerID = fid
	}
	if in.Category != nil {
		cid, _, err := resolveCategory(ctx, *in.Category)
		if err != nil {
			return nil, models.Item{}, err
		}
		current.CategoryID = cid
	}
	if in.UnknownDate {
		current.FrozenDate = nil
	} else if in.FrozenDate != nil && strings.TrimSpace(*in.FrozenDate) != "" {
		parsed, err := parseDate(*in.FrozenDate)
		if err != nil {
			return nil, models.Item{}, err
		}
		current.FrozenDate = &parsed
	}

	_, err = db.DB.ExecContext(ctx,
		"UPDATE items SET name=?, category_id=?, freezer_id=?, weight=?, frozen_date=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		current.Name, current.CategoryID, current.FreezerID, current.Weight, current.FrozenDate, current.ID,
	)
	if err != nil {
		return nil, models.Item{}, err
	}
	return nil, current, nil
}

// --- Query & Resolution Helpers ---

func queryFreezers(ctx context.Context) ([]models.Freezer, error) {
	rows, err := db.DB.QueryContext(ctx, "SELECT id, name, created_at, updated_at FROM freezers ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	freezers := []models.Freezer{}
	for rows.Next() {
		var f models.Freezer
		if err := rows.Scan(&f.ID, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		freezers = append(freezers, f)
	}
	return freezers, rows.Err()
}

func queryCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := db.DB.QueryContext(ctx, "SELECT id, name, COALESCE(icon, '') FROM categories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func queryEnrichedItems(ctx context.Context, freezerRef, categoryRef, search string, olderThanDays int) ([]EnrichedItem, error) {
	var clauses []string
	var args []any

	if strings.TrimSpace(freezerRef) != "" {
		fid, _, err := resolveFreezer(ctx, freezerRef)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, "i.freezer_id = ?")
		args = append(args, fid)
	}
	if strings.TrimSpace(categoryRef) != "" {
		cid, _, err := resolveCategory(ctx, categoryRef)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, "i.category_id = ?")
		args = append(args, cid)
	}
	if strings.TrimSpace(search) != "" {
		clauses = append(clauses, "LOWER(i.name) LIKE ?")
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(search))+"%")
	}
	if olderThanDays > 0 {
		cutoff := time.Now().UTC().AddDate(0, 0, -olderThanDays)
		clauses = append(clauses, "i.frozen_date IS NOT NULL AND i.frozen_date <= ?")
		args = append(args, cutoff)
	}

	query := `
		SELECT i.id, i.name, i.category_id, COALESCE(c.name, 'Uncategorized'),
		       i.freezer_id, COALESCE(f.name, ''), COALESCE(i.weight, ''),
		       i.frozen_date, i.created_at, i.updated_at
		FROM items i
		LEFT JOIN categories c ON i.category_id = c.id
		LEFT JOIN freezers f ON i.freezer_id = f.id`

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY (i.frozen_date IS NOT NULL) ASC, i.frozen_date ASC, i.id ASC"

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []EnrichedItem{}
	for rows.Next() {
		var it EnrichedItem
		var frozenDate sql.NullTime
		if err := rows.Scan(
			&it.ID, &it.Name, &it.CategoryID, &it.CategoryName,
			&it.FreezerID, &it.FreezerName, &it.Weight,
			&frozenDate, &it.CreatedAt, &it.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if frozenDate.Valid && !frozenDate.Time.IsZero() && frozenDate.Time.Year() > 1 {
			it.FrozenDate = &frozenDate.Time
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func findItemIDsFIFO(ctx context.Context, name, freezerRef string, limit int) ([]int64, error) {
	var clauses []string
	var args []any

	clauses = append(clauses, "LOWER(name) = ?")
	args = append(args, strings.ToLower(strings.TrimSpace(name)))

	if strings.TrimSpace(freezerRef) != "" {
		fid, _, err := resolveFreezer(ctx, freezerRef)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, "freezer_id = ?")
		args = append(args, fid)
	}

	query := fmt.Sprintf(
		"SELECT id FROM items WHERE %s ORDER BY (frozen_date IS NOT NULL) ASC, frozen_date ASC, id ASC LIMIT ?",
		strings.Join(clauses, " AND "),
	)
	args = append(args, limit)

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func resolveFreezer(ctx context.Context, ref string) (int64, string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return 0, "", fmt.Errorf("freezer reference is required")
	}
	var id int64
	var name string
	if numID, err := strconv.ParseInt(ref, 10, 64); err == nil {
		err = db.DB.QueryRowContext(ctx, "SELECT id, name FROM freezers WHERE id = ?", numID).Scan(&id, &name)
		if err == nil {
			return id, name, nil
		}
	}
	err := db.DB.QueryRowContext(ctx, "SELECT id, name FROM freezers WHERE LOWER(name) = LOWER(?)", ref).Scan(&id, &name)
	if err == sql.ErrNoRows {
		return 0, "", fmt.Errorf("freezer %q not found", ref)
	}
	if err != nil {
		return 0, "", err
	}
	return id, name, nil
}

func resolveCategory(ctx context.Context, ref string) (int64, string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return 0, "", fmt.Errorf("category reference is required")
	}
	var id int64
	var name string
	if numID, err := strconv.ParseInt(ref, 10, 64); err == nil {
		err = db.DB.QueryRowContext(ctx, "SELECT id, name FROM categories WHERE id = ?", numID).Scan(&id, &name)
		if err == nil {
			return id, name, nil
		}
	}
	err := db.DB.QueryRowContext(ctx, "SELECT id, name FROM categories WHERE LOWER(name) = LOWER(?)", ref).Scan(&id, &name)
	if err == sql.ErrNoRows {
		return 0, "", fmt.Errorf("category %q not found", ref)
	}
	if err != nil {
		return 0, "", err
	}
	return id, name, nil
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD or RFC3339", s)
}

func mapToNamedCounts(m map[string]int) []namedCount {
	out := make([]namedCount, 0, len(m))
	for k, v := range m {
		out = append(out, namedCount{Name: k, Count: v})
	}
	return out
}

