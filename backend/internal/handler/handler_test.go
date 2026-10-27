package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func siapkan(t *testing.T) http.Handler {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL belum diisi")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	skema := fmt.Sprintf("test_%d", time.Now().UnixNano())
	skrip, err := os.ReadFile("../../migrations/001_schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+skema); err != nil {
		t.Fatal(err)
	}
	cfg, _ := pgxpool.ParseConfig(url)
	cfg.ConnConfig.RuntimeParams["search_path"] = skema
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, string(skrip)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		admin.Exec(ctx, "DROP SCHEMA "+skema+" CASCADE")
		admin.Close()
	})
	return (&Handler{DB: db}).Routes()
}

func panggil(h http.Handler, method, path string, body any) (int, map[string]any) {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, &buf))
	var hasil map[string]any
	json.Unmarshal(rec.Body.Bytes(), &hasil)
	return rec.Code, hasil
}

func buatProduct(t *testing.T, h http.Handler, sku string, harga, stok int) int {
	code, res := panggil(h, "POST", "/api/products", map[string]any{"sku": sku, "name": "Produk " + sku, "price": harga, "stock": stok})
	if code != http.StatusCreated {
		t.Fatalf("buat produk %s gagal: %d %v", sku, code, res)
	}
	return int(res["data"].(map[string]any)["id"].(float64))
}

func stokSekarang(t *testing.T, h http.Handler, id int) int {
	_, res := panggil(h, "GET", "/api/products?search=", nil)
	for _, p := range res["data"].([]any) {
		m := p.(map[string]any)
		if int(m["id"].(float64)) == id {
			return int(m["stock"].(float64))
		}
	}
	t.Fatalf("produk %d tidak ditemukan", id)
	return 0
}

func jumlahBaris(h http.Handler, path string) int {
	_, res := panggil(h, "GET", path, nil)
	return len(res["data"].([]any))
}

func TestStokKurangDitolakTanpaMenulisApaPun(t *testing.T) {
	h := siapkan(t)
	a := buatProduct(t, h, "AAA", 1000, 5)
	b := buatProduct(t, h, "BBB", 2000, 3)

	code, res := panggil(h, "POST", "/api/orders", map[string]any{
		"customer": "Toko Makmur Jaya",
		"items":    []map[string]any{{"sku": "AAA", "qty": 2}, {"sku": "BBB", "qty": 10}},
	})
	if code != http.StatusConflict {
		t.Fatalf("status %d, mau 409: %v", code, res)
	}
	e := res["error"].(map[string]any)
	if e["code"] != "INSUFFICIENT_STOCK" || e["sku"] != "BBB" || e["available"].(float64) != 3 {
		t.Fatalf("isi error salah: %v", e)
	}
	if stokSekarang(t, h, a) != 5 || stokSekarang(t, h, b) != 3 {
		t.Fatal("stok berubah padahal order ditolak")
	}
	if jumlahBaris(h, "/api/orders") != 0 {
		t.Fatal("ada order tersimpan")
	}
	if jumlahBaris(h, fmt.Sprintf("/api/products/%d/movements", a)) != 1 {
		t.Fatal("ada movement tambahan selain stok awal")
	}
}

func TestOrderBerhasilLaluBatal(t *testing.T) {
	h := siapkan(t)
	id := buatProduct(t, h, "CCC", 15000, 10)

	code, res := panggil(h, "POST", "/api/orders", map[string]any{
		"customer": "Warung Bu Ani",
		"items":    []map[string]any{{"sku": "CCC", "qty": 4, "unit_price": 1}},
	})
	if code != http.StatusCreated {
		t.Fatalf("status %d: %v", code, res)
	}
	o := res["data"].(map[string]any)
	if o["total"].(float64) != 60000 {
		t.Fatalf("total %v, mau 60000", o["total"])
	}
	if stokSekarang(t, h, id) != 6 {
		t.Fatal("stok tidak berkurang jadi 6")
	}
	if jumlahBaris(h, fmt.Sprintf("/api/products/%d/movements", id)) != 2 {
		t.Fatal("movement out tidak tercatat")
	}

	path := fmt.Sprintf("/api/orders/%d/cancel", int(o["id"].(float64)))
	if code, _ = panggil(h, "POST", path, nil); code != http.StatusOK {
		t.Fatalf("cancel status %d", code)
	}
	if stokSekarang(t, h, id) != 10 {
		t.Fatal("stok tidak kembali ke 10")
	}
	if code, _ = panggil(h, "POST", path, nil); code != http.StatusConflict {
		t.Fatalf("cancel kedua status %d, mau 409", code)
	}
	if stokSekarang(t, h, id) != 10 {
		t.Fatal("stok bertambah dua kali")
	}
}

func TestValidasiOrder(t *testing.T) {
	h := siapkan(t)
	buatProduct(t, h, "DDD", 1000, 5)
	kasus := []map[string]any{
		{"customer": "", "items": []map[string]any{{"sku": "DDD", "qty": 1}}},
		{"customer": "Toko A", "items": []map[string]any{}},
		{"customer": "Toko A", "items": []map[string]any{{"sku": "ZZZ", "qty": 1}}},
		{"customer": "Toko A", "items": []map[string]any{{"sku": "DDD", "qty": 0}}},
		{"customer": "Toko A", "items": []map[string]any{{"sku": "DDD", "qty": 1}, {"sku": "ddd", "qty": 1}}},
	}
	for i, k := range kasus {
		if code, res := panggil(h, "POST", "/api/orders", k); code != http.StatusBadRequest {
			t.Fatalf("kasus %d: status %d, mau 400: %v", i, code, res)
		}
	}
}

func TestOrderBersamaanTidakMembuatStokMinus(t *testing.T) {
	h := siapkan(t)
	id := buatProduct(t, h, "EEE", 5000, 10)

	var wg sync.WaitGroup
	var mu sync.Mutex
	hitung := map[int]int{}
	for range 40 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			code, _ := panggil(h, "POST", "/api/orders", map[string]any{
				"customer": "Pembeli Serentak",
				"items":    []map[string]any{{"sku": "EEE", "qty": 1}},
			})
			mu.Lock()
			hitung[code]++
			mu.Unlock()
		}()
	}
	wg.Wait()

	if hitung[http.StatusCreated] != 10 || hitung[http.StatusConflict] != 30 {
		t.Fatalf("hasil %v, mau 10x201 dan 30x409", hitung)
	}
	if stokSekarang(t, h, id) != 0 {
		t.Fatal("stok akhir bukan 0")
	}
	if jumlahBaris(h, "/api/orders?limit=100") != 10 {
		t.Fatal("jumlah order tersimpan bukan 10")
	}
}

func TestIdempotencyKey(t *testing.T) {
	h := siapkan(t)
	id := buatProduct(t, h, "FFF", 1000, 10)
	body, _ := json.Marshal(map[string]any{"customer": "Toko Retry", "items": []map[string]any{{"sku": "FFF", "qty": 3}}})

	kirimUlang := func() int {
		req := httptest.NewRequest("POST", "/api/orders", bytes.NewReader(body))
		req.Header.Set("Idempotency-Key", "abc-123")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	if c := kirimUlang(); c != http.StatusCreated {
		t.Fatalf("pertama %d", c)
	}
	if c := kirimUlang(); c != http.StatusOK {
		t.Fatalf("kedua %d, mau 200", c)
	}
	if stokSekarang(t, h, id) != 7 {
		t.Fatal("stok terpotong dua kali")
	}
}
