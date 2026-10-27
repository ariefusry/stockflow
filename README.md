# StockFlow
Mini ERP for inventory and sales: Go + PostgreSQL on the backend, React + TypeScript (Vite) on the frontend.

## Running
All services at once (Postgres with migrations and seed data, backend, frontend):

    docker compose up --build

- Frontend: http://localhost:5173
- Backend: http://localhost:8080
- Postgres: localhost:5433 (user, password, and database: stockflow)

Dev mode (database only in Docker):

    docker compose up -d db
    cd backend && go run ./cmd/server
    cd frontend && npm install && npm run dev

Vite proxies `/api` to `localhost:8080`, so CORS is not needed. In compose, nginx proxies `/api` to the backend service.

Reset data: `docker compose down -v`, then `up` again.

## Tests
    cd backend
    TEST_DATABASE_URL=postgres://stockflow:stockflow@localhost:5433/stockflow go test -race ./...

Each test creates its own Postgres schema and drops it afterward, so the seed data is never touched. `-race` requires cgo (gcc).

    cd frontend
    npm test

## Pages
- Products: table with search (debounced), filters All/Active/Inactive/Low stock, paging, stock adjustment, movement history, add/edit/delete products.
- Orders: order list, create order, order detail with a cancel button.
- Summary cards at the top: total products, active products, stock value, low stock (10 or fewer), sales today.

## curl examples
    curl "localhost:8080/api/products?search=beras&page=1&limit=5"

    curl -X POST localhost:8080/api/products -d '{"sku":"SNK-CHT-70","name":"Chitato Sapi Panggang 68 g","price":11500,"stock":30}'

    curl -X POST localhost:8080/api/products/3/stock -d '{"type":"in","qty":20,"note":"kiriman distributor"}'

    curl -X POST localhost:8080/api/orders \
      -H "Idempotency-Key: pesanan-001" \
      -d '{"customer":"Toko Berkah Malang","items":[{"sku":"MIE-IND-GRG","qty":24},{"sku":"TEH-SSR-25","qty":6}]}'

    curl -X POST localhost:8080/api/orders/1/cancel

    curl "localhost:8080/api/reports/daily-sales?date=2026-10-27"

Error format: `{"error":{"code":"INSUFFICIENT_STOCK","message":"...","sku":"...","available":3}}`.

## Design decisions
- `last_movement_at` is computed with a subquery on `max(created_at)` from `stock_movements`, not stored on the product row, so it cannot go out of sync.
- Stock is reduced with `UPDATE ... SET stock = stock - $qty WHERE sku = $sku AND stock >= $qty`. If no row changes, the transaction is rolled back and the response is 409. There is no SELECT followed by UPDATE, so concurrent orders cannot drive stock negative. Items are sorted by SKU before updating so two orders do not deadlock.
- The entire order (header, items, stock reduction, movements) runs in a single transaction.
- Cancelling an order locks the order row with `FOR UPDATE`, so when two cancels run at the same time only one succeeds and the second gets 409 `ALREADY_CANCELLED`.
- Deleting a product is a plain `DELETE`. Foreign keys from `stock_movements` and `sales_order_items` block deletion of products with history, and that error is mapped to 409 `PRODUCT_HAS_HISTORY`.
- SKUs are normalized to uppercase and trimmed, so `abc` and `ABC` do not become two products.
- Orders for inactive products are rejected with 400.
- Extras beyond the contract: `GET /api/orders` (order list), `GET /api/summary` (summary cards), the `status=active|inactive|low` parameter on the product list, an optional `stock` field when creating a product (recorded as an "initial stock" movement), and an `idempotency_keys` table for the `Idempotency-Key` header.
- The day used in the daily sales report follows the Asia/Jakarta time zone.

## Time spent
About 14 hours, based on the git log of my commits.

- Repo setup, schema, seed: 1 hour
- Go backend (API, tests, Dockerfile): 4 hours
- Frontend (setup, client, products and orders pages): 5.5 hours
- Summary cards, Vitest tests, frontend Dockerfile: 3.5 hours

These figures include break time.

## Planned additions
- Authentication (login/register) and user roles.
- The product list in the order form is currently limited to 100 active products; it needs product selection with search.
- Cleanup of old rows in `idempotency_keys`.
- React component tests and E2E tests.
- Migrations with a tool such as golang-migrate, since the SQL files currently only run when the Postgres volume is first created.
