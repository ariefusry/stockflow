package handler

import (
	"net/http"
	"time"
)

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	var s struct {
		TotalProducts  int `json:"total_products"`
		ActiveProducts int `json:"active_products"`
		StockValue     int `json:"stock_value"`
		LowStock       int `json:"low_stock"`
	}
	err := h.DB.QueryRow(r.Context(),
		`SELECT count(*),
			count(*) FILTER (WHERE is_active),
			coalesce(sum(price::bigint * stock), 0),
			count(*) FILTER (WHERE is_active AND stock <= $1)
		 FROM products`, batasStokRendah).Scan(&s.TotalProducts, &s.ActiveProducts, &s.StockValue, &s.LowStock)
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusOK, map[string]any{"data": s})
}

func (h *Handler) dailySales(w http.ResponseWriter, r *http.Request) {
	tanggal := r.URL.Query().Get("date")
	if tanggal == "" {
		tanggal = time.Now().In(time.FixedZone("WIB", 7*3600)).Format(time.DateOnly)
	} else if _, err := time.Parse(time.DateOnly, tanggal); err != nil {
		galat(w, http.StatusBadRequest, "VALIDATION_ERROR", "date must use the format YYYY-MM-DD", nil)
		return
	}
	var orders, terjual, omzet int
	err := h.DB.QueryRow(r.Context(),
		`SELECT count(DISTINCT o.id), coalesce(sum(i.qty), 0), coalesce(sum(i.qty::bigint * i.unit_price), 0)
		 FROM sales_orders o JOIN sales_order_items i ON i.order_id = o.id
		 WHERE o.status = 'CREATED' AND (o.created_at AT TIME ZONE 'Asia/Jakarta')::date = $1::date`,
		tanggal).Scan(&orders, &terjual, &omzet)
	if err != nil {
		galatServer(w, err)
		return
	}
	kirim(w, http.StatusOK, map[string]any{"data": map[string]any{
		"date": tanggal, "orders": orders, "items_sold": terjual, "revenue": omzet,
	}})
}
