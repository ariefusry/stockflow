# StockFlow
Mini ERP untuk inventori dan penjualan: Go + PostgreSQL di backend, React + TypeScript (Vite) di frontend.

## Menjalankan
Semua service sekaligus (Postgres dengan migrasi dan seed, backend, frontend):

    docker compose up --build

- Frontend: http://localhost:5173
- Backend: http://localhost:8080
- Postgres: localhost:5433 (user, password, dan database: stockflow)

Mode dev (hanya database di Docker):

    docker compose up -d db
    cd backend && go run ./cmd/server
    cd frontend && npm install && npm run dev

Vite mem-proxy `/api` ke `localhost:8080`, jadi tidak butuh CORS. Di compose, nginx yang mem-proxy `/api` ke service backend.

Reset data: `docker compose down -v`, lalu `up` lagi.

## Tes
    cd backend
    TEST_DATABASE_URL=postgres://stockflow:stockflow@localhost:5433/stockflow go test -race ./...

Setiap tes membuat schema Postgres sendiri lalu menghapusnya, jadi data seed tidak tersentuh. `-race` butuh cgo (gcc).

    cd frontend
    npm test

## Halaman
- Products: tabel dengan pencarian (debounce), filter All/Active/Inactive/Low stock, paging, adjust stock, riwayat movement, tambah/ubah/hapus produk.
- Orders: daftar order, buat order, detail order dengan tombol cancel.
- Kartu ringkasan di atas: total produk, produk aktif, nilai stok, stok rendah (10 atau kurang), penjualan hari ini.

## Contoh curl
    curl "localhost:8080/api/products?search=beras&page=1&limit=5"

    curl -X POST localhost:8080/api/products -d '{"sku":"SNK-CHT-70","name":"Chitato Sapi Panggang 68 g","price":11500,"stock":30}'

    curl -X POST localhost:8080/api/products/3/stock -d '{"type":"in","qty":20,"note":"kiriman distributor"}'

    curl -X POST localhost:8080/api/orders \
      -H "Idempotency-Key: pesanan-001" \
      -d '{"customer":"Toko Berkah Malang","items":[{"sku":"MIE-IND-GRG","qty":24},{"sku":"TEH-SSR-25","qty":6}]}'

    curl -X POST localhost:8080/api/orders/1/cancel

    curl "localhost:8080/api/reports/daily-sales?date=2026-10-27"

Format error: `{"error":{"code":"INSUFFICIENT_STOCK","message":"...","sku":"...","available":3}}`.

## Keputusan desain
- `last_movement_at` dihitung dengan subquery `max(created_at)` pada `stock_movements`, tidak disimpan di baris produk, supaya tidak bisa tidak sinkron.
- Stok dikurangi dengan `UPDATE ... SET stock = stock - $qty WHERE sku = $sku AND stock >= $qty`. Kalau tidak ada baris yang berubah, transaksi di-rollback dan hasilnya 409. Tidak ada SELECT lalu UPDATE, jadi order bersamaan tidak bisa membuat stok minus. Item diurutkan menurut SKU sebelum di-update supaya dua order tidak saling deadlock.
- Seluruh order (header, item, pengurangan stok, movement) berada dalam satu transaksi.
- Cancel order mengunci baris order dengan `FOR UPDATE`, jadi dua cancel bersamaan hanya satu yang berhasil dan yang kedua dapat 409 `ALREADY_CANCELLED`.
- Hapus produk cukup `DELETE`. Foreign key dari `stock_movements` dan `sales_order_items` menggagalkan penghapusan produk yang punya riwayat, dan error itu dipetakan ke 409 `PRODUCT_HAS_HISTORY`.
- SKU dinormalisasi ke huruf besar dan di-trim, supaya `abc` dan `ABC` tidak jadi dua produk.
- Order untuk produk non-aktif ditolak dengan 400.
- Tambahan di luar kontrak: `GET /api/orders` (daftar order), `GET /api/summary` (kartu ringkasan), parameter `status=active|inactive|low` pada daftar produk, field opsional `stock` saat membuat produk (dicatat sebagai movement "initial stock"), tabel `idempotency_keys` untuk header `Idempotency-Key`.
- Hari pada laporan penjualan harian memakai zona Asia/Jakarta.

## Waktu pengerjaan
Sekitar 14 jam, waktu ini didapatkan sesuai dengan git log dari proses saya melakukan commit.

- Setup repo, skema, seed - 1 jam 
- Backend Go (API, tes, Dockerfile) - 4 jam 
- Frontend (setup, client, halaman produk dan order) - 5,5 jam 
- Kartu ringkasan, tes Vitest, Dockerfile frontend - 3,5 jam 

## Yang akan ditambah
- Autentikasi (Login/Register) dan peran pengguna.
- Daftar produk di form order masih maksimal 100 produk aktif, serta perlu pilihan produk dengan pencarian.
- Pembersihan `idempotency_keys` yang lama.
- Tes komponen React dan tes E2E.
- Migrasi dengan tool seperti golang-migrate, karena saat ini file SQL hanya dijalankan saat volume Postgres pertama kali dibuat.
