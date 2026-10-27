package handler

import (
	"encoding/json"
	"errors"
	"log"
	"maps"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/products", h.listProducts)
	mux.HandleFunc("POST /api/products", h.createProduct)
	mux.HandleFunc("PUT /api/products/{id}", h.updateProduct)
	mux.HandleFunc("DELETE /api/products/{id}", h.deleteProduct)
	mux.HandleFunc("POST /api/products/{id}/stock", h.adjustStock)
	mux.HandleFunc("GET /api/products/{id}/movements", h.listMovements)
	mux.HandleFunc("GET /api/orders", h.listOrders)
	mux.HandleFunc("POST /api/orders", h.createOrder)
	mux.HandleFunc("GET /api/orders/{id}", h.getOrder)
	mux.HandleFunc("POST /api/orders/{id}/cancel", h.cancelOrder)
	mux.HandleFunc("GET /api/summary", h.summary)
	mux.HandleFunc("GET /api/reports/daily-sales", h.dailySales)
	return mux
}

func kirim(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func galat(w http.ResponseWriter, status int, code, pesan string, tambahan map[string]any) {
	e := map[string]any{"code": code, "message": pesan}
	maps.Copy(e, tambahan)
	kirim(w, status, map[string]any{"error": e})
}

func galatServer(w http.ResponseWriter, err error) {
	log.Println("db error:", err)
	galat(w, http.StatusInternalServerError, "INTERNAL", "database operation failed", nil)
}

func bacaBody(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		galat(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON or a field has the wrong type", nil)
		return false
	}
	return true
}

func bacaID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		galat(w, http.StatusBadRequest, "INVALID_ID", "id must be a positive integer", nil)
		return 0, false
	}
	return id, true
}

func halaman(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return page, min(limit, 100)
}

func kodePg(err error) string {
	var e *pgconn.PgError
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}
