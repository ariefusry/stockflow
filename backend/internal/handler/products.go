package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Product struct {
	ID             int        `db:"id" json:"id"`
	SKU            string     `db:"sku" json:"sku"`
	Name           string     `db:"name" json:"name"`
	Price          int        `db:"price" json:"price"`
	Stock          int        `db:"stock" json:"stock"`
	IsActive       bool       `db:"is_active" json:"is_active"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	LastMovementAt *time.Time `db:"last_movement_at" json:"last_movement_at"`
}

type Movement struct {
	ID        int       `db:"id" json:"id"`
	ProductID int       `db:"product_id" json:"product_id"`
	Type      string    `db:"type" json:"type"`
	Qty       int       `db:"qty" json:"qty"`
	Note      *string   `db:"note" json:"note"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

const kolomProduct = `p.id, p.sku, p.name, p.price, p.stock, p.is_active, p.created_at,
	(SELECT max(m.created_at) FROM stock_movements m WHERE m.product_id = p.id) AS last_movement_at`

const batasStokRendah = 10

var pengganti = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (h *Handler) cariProduct(ctx context.Context, id int) (Product, error) {
	rows, _ := h.DB.Query(ctx, "SELECT "+kolomProduct+" FROM products p WHERE p.id = $1", id)
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[Product])
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, limit := halaman(r)
	cari := "%" + pengganti.Replace(strings.TrimSpace(q.Get("search"))) + "%"

	where := "(p.sku ILIKE $1 OR p.name ILIKE $1)"
	switch q.Get("status") {
	case "active":
		where += " AND p.is_active"
	case "inactive":
		where += " AND NOT p.is_active"
	case "low":
		where += fmt.Sprintf(" AND p.is_active AND p.stock <= %d", batasStokRendah)
	}

	var total int
	if err := h.DB.QueryRow(r.Context(), "SELECT count(*) FROM products p WHERE "+where, cari).Scan(&total); err != nil {
		galatServer(w, err)
		return
	}
	rows, _ := h.DB.Query(r.Context(),
		"SELECT "+kolomProduct+" FROM products p WHERE "+where+" ORDER BY p.id LIMIT $2 OFFSET $3",
		cari, limit, (page-1)*limit)
	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[Product])
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusOK, map[string]any{"data": data, "page": page, "limit": limit, "total": total})
}

type inputProduct struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Price    *int   `json:"price"`
	Stock    int    `json:"stock"`
	IsActive *bool  `json:"is_active"`
}

func (in *inputProduct) validasi() string {
	in.SKU = strings.ToUpper(strings.TrimSpace(in.SKU))
	in.Name = strings.TrimSpace(in.Name)
	switch {
	case in.SKU == "":
		return "sku is required"
	case in.Name == "":
		return "name is required"
	case in.Price == nil || *in.Price < 0:
		return "price is required and must be an integer >= 0"
	case in.Stock < 0:
		return "stock must be >= 0"
	}
	return ""
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var in inputProduct
	if !bacaBody(w, r, &in) {
		return
	}
	if pesan := in.validasi(); pesan != "" {
		galat(w, http.StatusBadRequest, "VALIDATION_ERROR", pesan, nil)
		return
	}
	aktif := in.IsActive == nil || *in.IsActive

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		galatServer(w, err)
		return
	}
	defer tx.Rollback(r.Context())

	var id int
	err = tx.QueryRow(r.Context(),
		"INSERT INTO products (sku, name, price, stock, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		in.SKU, in.Name, *in.Price, in.Stock, aktif).Scan(&id)
	if kodePg(err) == "23505" {
		galat(w, http.StatusConflict, "DUPLICATE_SKU", "sku "+in.SKU+" already exists", nil)
		return
	}
	if err != nil {
		galatServer(w, err)
		return
	}
	if in.Stock > 0 {
		_, err = tx.Exec(r.Context(),
			"INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, 'in', $2, 'initial stock')", id, in.Stock)
		if err != nil {
			galatServer(w, err)
			return
		}
	}
	if err = tx.Commit(r.Context()); err != nil {
		galatServer(w, err)
		return
	}
	p, err := h.cariProduct(r.Context(), id)
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusCreated, map[string]any{"data": p})
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := bacaID(w, r)
	if !ok {
		return
	}
	var in inputProduct
	if !bacaBody(w, r, &in) {
		return
	}
	if pesan := in.validasi(); pesan != "" {
		galat(w, http.StatusBadRequest, "VALIDATION_ERROR", pesan, nil)
		return
	}
	hasil, err := h.DB.Exec(r.Context(),
		"UPDATE products SET sku = $2, name = $3, price = $4, is_active = COALESCE($5, is_active) WHERE id = $1",
		id, in.SKU, in.Name, *in.Price, in.IsActive)
	if kodePg(err) == "23505" {
		galat(w, http.StatusConflict, "DUPLICATE_SKU", "sku "+in.SKU+" already exists", nil)
		return
	}
	if err != nil {
		galatServer(w, err)
		return
	}
	if hasil.RowsAffected() == 0 {
		galat(w, http.StatusNotFound, "NOT_FOUND", "product not found", nil)
		return
	}
	p, err := h.cariProduct(r.Context(), id)
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusOK, map[string]any{"data": p})
}

func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := bacaID(w, r)
	if !ok {
		return
	}
	hasil, err := h.DB.Exec(r.Context(), "DELETE FROM products WHERE id = $1", id)
	if kodePg(err) == "23503" {
		galat(w, http.StatusConflict, "PRODUCT_HAS_HISTORY", "product has stock movements or orders and cannot be deleted", nil)
		return
	}
	if err != nil {
		galatServer(w, err)
		return
	}
	if hasil.RowsAffected() == 0 {
		galat(w, http.StatusNotFound, "NOT_FOUND", "product not found", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adjustStock(w http.ResponseWriter, r *http.Request) {
	id, ok := bacaID(w, r)
	if !ok {
		return
	}
	var in struct {
		Type string `json:"type"`
		Qty  int    `json:"qty"`
		Note string `json:"note"`
	}
	if !bacaBody(w, r, &in) {
		return
	}
	if in.Type != "in" && in.Type != "out" {
		galat(w, http.StatusBadRequest, "VALIDATION_ERROR", `type must be "in" or "out"`, nil)
		return
	}
	if in.Qty <= 0 {
		galat(w, http.StatusBadRequest, "VALIDATION_ERROR", "qty must be an integer greater than 0", nil)
		return
	}
	delta := in.Qty
	if in.Type == "out" {
		delta = -in.Qty
	}
	var note *string
	if n := strings.TrimSpace(in.Note); n != "" {
		note = &n
	}

	tx, err := h.DB.Begin(r.Context())
	if err != nil {
		galatServer(w, err)
		return
	}
	defer tx.Rollback(r.Context())

	var stok int
	err = tx.QueryRow(r.Context(),
		"UPDATE products SET stock = stock + $2 WHERE id = $1 AND stock + $2 >= 0 RETURNING stock", id, delta).Scan(&stok)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(r.Context(), "SELECT stock FROM products WHERE id = $1", id).Scan(&stok)
		if errors.Is(err, pgx.ErrNoRows) {
			galat(w, http.StatusNotFound, "NOT_FOUND", "product not found", nil)
			return
		}
		if err != nil {
			galatServer(w, err)
			return
		}
		galat(w, http.StatusConflict, "INSUFFICIENT_STOCK",
			fmt.Sprintf("insufficient stock: only %d available", stok), map[string]any{"available": stok})
		return
	}
	if err != nil {
		galatServer(w, err)
		return
	}

	var m Movement
	rows, _ := tx.Query(r.Context(),
		"INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, $2, $3, $4) RETURNING id, product_id, type, qty, note, created_at",
		id, in.Type, in.Qty, note)
	m, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[Movement])
	if err != nil {
		galatServer(w, err)
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusCreated, map[string]any{"data": map[string]any{"movement": m, "stock": stok}})
}

func (h *Handler) listMovements(w http.ResponseWriter, r *http.Request) {
	id, ok := bacaID(w, r)
	if !ok {
		return
	}
	var ada bool
	if err := h.DB.QueryRow(r.Context(), "SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)", id).Scan(&ada); err != nil {
		galatServer(w, err)
		return
	}
	if !ada {
		galat(w, http.StatusNotFound, "NOT_FOUND", "product not found", nil)
		return
	}
	rows, _ := h.DB.Query(r.Context(),
		"SELECT id, product_id, type, qty, note, created_at FROM stock_movements WHERE product_id = $1 ORDER BY created_at DESC, id DESC", id)
	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[Movement])
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusOK, map[string]any{"data": data})
}
