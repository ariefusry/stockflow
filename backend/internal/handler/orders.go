package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type OrderItem struct {
	SKU       string `db:"sku" json:"sku"`
	Name      string `db:"name" json:"name"`
	Qty       int    `db:"qty" json:"qty"`
	UnitPrice int    `db:"unit_price" json:"unit_price"`
}

type Order struct {
	ID          int         `db:"id" json:"id"`
	Customer    string      `db:"customer" json:"customer"`
	Status      string      `db:"status" json:"status"`
	Total       int         `db:"total" json:"total"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	CancelledAt *time.Time  `db:"cancelled_at" json:"cancelled_at"`
	Items       []OrderItem `db:"-" json:"items,omitempty"`
}

type itemInput struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

const kolomOrder = "id, customer, status, total, created_at, cancelled_at"

func (h *Handler) ambilOrder(ctx context.Context, id int) (Order, error) {
	rows, _ := h.DB.Query(ctx, "SELECT "+kolomOrder+" FROM sales_orders WHERE id = $1", id)
	o, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Order])
	if err != nil {
		return o, err
	}
	rows, _ = h.DB.Query(ctx,
		`SELECT i.sku, p.name, i.qty, i.unit_price FROM sales_order_items i
		 JOIN products p ON p.sku = i.sku WHERE i.order_id = $1 ORDER BY i.id`, id)
	o.Items, err = pgx.CollectRows(rows, pgx.RowToStructByName[OrderItem])
	return o, err
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := bacaID(w, r)
	if !ok {
		return
	}
	o, err := h.ambilOrder(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		galat(w, http.StatusNotFound, "NOT_FOUND", "order not found", nil)
		return
	}
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusOK, map[string]any{"data": o})
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	page, limit := halaman(r)
	var total int
	if err := h.DB.QueryRow(r.Context(), "SELECT count(*) FROM sales_orders").Scan(&total); err != nil {
		galatServer(w, err)
		return
	}
	rows, _ := h.DB.Query(r.Context(),
		"SELECT "+kolomOrder+" FROM sales_orders ORDER BY id DESC LIMIT $1 OFFSET $2", limit, (page-1)*limit)
	data, err := pgx.CollectRows(rows, pgx.RowToStructByName[Order])
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusOK, map[string]any{"data": data, "page": page, "limit": limit, "total": total})
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Customer string      `json:"customer"`
		Items    []itemInput `json:"items"`
	}
	if !bacaBody(w, r, &in) {
		return
	}
	in.Customer = strings.TrimSpace(in.Customer)
	if in.Customer == "" {
		galat(w, http.StatusBadRequest, "VALIDATION_ERROR", "customer is required", nil)
		return
	}
	if len(in.Items) == 0 {
		galat(w, http.StatusBadRequest, "VALIDATION_ERROR", "items must not be empty", nil)
		return
	}
	skus := make([]string, 0, len(in.Items))
	for i := range in.Items {
		it := &in.Items[i]
		it.SKU = strings.ToUpper(strings.TrimSpace(it.SKU))
		if it.SKU == "" {
			galat(w, http.StatusBadRequest, "VALIDATION_ERROR", fmt.Sprintf("item %d: sku is required", i+1), nil)
			return
		}
		if it.Qty <= 0 {
			galat(w, http.StatusBadRequest, "VALIDATION_ERROR", fmt.Sprintf("item %d (%s): qty must be greater than 0", i+1, it.SKU), nil)
			return
		}
		if slices.Contains(skus, it.SKU) {
			galat(w, http.StatusBadRequest, "VALIDATION_ERROR", "duplicate sku in order: "+it.SKU, nil)
			return
		}
		skus = append(skus, it.SKU)
	}

	ctx := r.Context()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		galatServer(w, err)
		return
	}
	defer tx.Rollback(ctx)

	kunci := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if kunci != "" {
		hasil, err := tx.Exec(ctx, "INSERT INTO idempotency_keys (key) VALUES ($1) ON CONFLICT DO NOTHING", kunci)
		if err != nil {
			galatServer(w, err)
			return
		}
		if hasil.RowsAffected() == 0 {
			var lama int
			if err := tx.QueryRow(ctx, "SELECT order_id FROM idempotency_keys WHERE key = $1", kunci).Scan(&lama); err != nil {
				galatServer(w, err)
				return
			}
			o, err := h.ambilOrder(ctx, lama)
			if err != nil {
				galatServer(w, err)
				return
			}
			kirim(w, http.StatusOK, map[string]any{"data": o})
			return
		}
	}

	rows, _ := tx.Query(ctx, "SELECT sku, is_active FROM products WHERE sku = ANY($1)", skus)
	cek, err := pgx.CollectRows(rows, pgx.RowToStructByName[struct {
		SKU      string `db:"sku"`
		IsActive bool   `db:"is_active"`
	}])
	if err != nil {
		galatServer(w, err)
		return
	}
	ada := map[string]bool{}
	for _, c := range cek {
		ada[c.SKU] = c.IsActive
	}
	for _, s := range skus {
		aktif, ketemu := ada[s]
		if !ketemu {
			galat(w, http.StatusBadRequest, "VALIDATION_ERROR", "unknown sku: "+s, map[string]any{"sku": s})
			return
		}
		if !aktif {
			galat(w, http.StatusBadRequest, "VALIDATION_ERROR", "product is inactive: "+s, map[string]any{"sku": s})
			return
		}
	}

	var idOrder int
	err = tx.QueryRow(ctx, "INSERT INTO sales_orders (customer, total) VALUES ($1, 0) RETURNING id", in.Customer).Scan(&idOrder)
	if err != nil {
		galatServer(w, err)
		return
	}

	slices.SortFunc(in.Items, func(a, b itemInput) int { return strings.Compare(a.SKU, b.SKU) })
	total := 0
	for _, it := range in.Items {
		var idProd, harga int
		err = tx.QueryRow(ctx,
			"UPDATE products SET stock = stock - $2 WHERE sku = $1 AND stock >= $2 RETURNING id, price", it.SKU, it.Qty).Scan(&idProd, &harga)
		if errors.Is(err, pgx.ErrNoRows) {
			var sisa int
			if err := tx.QueryRow(ctx, "SELECT stock FROM products WHERE sku = $1", it.SKU).Scan(&sisa); err != nil {
				galatServer(w, err)
				return
			}
			galat(w, http.StatusConflict, "INSUFFICIENT_STOCK",
				fmt.Sprintf("insufficient stock for %s: requested %d, available %d", it.SKU, it.Qty, sisa),
				map[string]any{"sku": it.SKU, "available": sisa})
			return
		}
		if err != nil {
			galatServer(w, err)
			return
		}
		_, err = tx.Exec(ctx,
			"INSERT INTO sales_order_items (order_id, sku, qty, unit_price) VALUES ($1, $2, $3, $4)", idOrder, it.SKU, it.Qty, harga)
		if err == nil {
			_, err = tx.Exec(ctx,
				"INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, 'out', $2, $3)",
				idProd, it.Qty, fmt.Sprintf("order #%d", idOrder))
		}
		if err != nil {
			galatServer(w, err)
			return
		}
		total += harga * it.Qty
	}

	_, err = tx.Exec(ctx, "UPDATE sales_orders SET total = $2 WHERE id = $1", idOrder, total)
	if err == nil && kunci != "" {
		_, err = tx.Exec(ctx, "UPDATE idempotency_keys SET order_id = $2 WHERE key = $1", kunci, idOrder)
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		galatServer(w, err)
		return
	}
	o, err := h.ambilOrder(ctx, idOrder)
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusCreated, map[string]any{"data": o})
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := bacaID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		galatServer(w, err)
		return
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM sales_orders WHERE id = $1 FOR UPDATE", id).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		galat(w, http.StatusNotFound, "NOT_FOUND", "order not found", nil)
		return
	}
	if err != nil {
		galatServer(w, err)
		return
	}
	if status == "CANCELLED" {
		galat(w, http.StatusConflict, "ALREADY_CANCELLED", "order is already cancelled", nil)
		return
	}

	rows, _ := tx.Query(ctx, "SELECT sku, qty FROM sales_order_items WHERE order_id = $1 ORDER BY sku", id)
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[itemInput])
	if err != nil {
		galatServer(w, err)
		return
	}
	for _, it := range items {
		var idProd int
		err = tx.QueryRow(ctx, "UPDATE products SET stock = stock + $2 WHERE sku = $1 RETURNING id", it.SKU, it.Qty).Scan(&idProd)
		if err == nil {
			_, err = tx.Exec(ctx,
				"INSERT INTO stock_movements (product_id, type, qty, note) VALUES ($1, 'in', $2, $3)",
				idProd, it.Qty, fmt.Sprintf("cancel order #%d", id))
		}
		if err != nil {
			galatServer(w, err)
			return
		}
	}
	_, err = tx.Exec(ctx, "UPDATE sales_orders SET status = 'CANCELLED', cancelled_at = now() WHERE id = $1", id)
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		galatServer(w, err)
		return
	}
	o, err := h.ambilOrder(ctx, id)
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusOK, map[string]any{"data": o})
}
