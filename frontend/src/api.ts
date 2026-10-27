import type {
  DailySales,
  Movement,
  Order,
  OrderBody,
  Page,
  Product,
  ProductBody,
  Summary,
} from "./types";

export class ApiError extends Error {
  status: number;
  code: string;
  extra: Record<string, unknown>;

  constructor(status: number, code: string, message: string, extra: Record<string, unknown> = {}) {
    super(message);
    this.status = status;
    this.code = code;
    this.extra = extra;
  }
}

type Envelope = { error?: Record<string, unknown> };

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  headers: Record<string, string> = {},
): Promise<T> {
  let res: Response;
  let json: unknown = null;
  try {
    res = await fetch(path, {
      method,
      headers: body === undefined ? headers : { "Content-Type": "application/json", ...headers },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    if (res.status !== 204) json = await res.json();
  } catch {
    throw new ApiError(0, "NETWORK", "Cannot reach the server or the response was not valid JSON. Check that the backend is running.");
  }
  if (!res.ok) {
    const e = (json as Envelope | null)?.error ?? {};
    throw new ApiError(
      res.status,
      typeof e.code === "string" ? e.code : "UNKNOWN",
      typeof e.message === "string" ? e.message : `Request failed with status ${res.status}`,
      e,
    );
  }
  return json as T;
}

export const pesan = (e: unknown) => (e instanceof Error ? e.message : "Unexpected error");

export const getProducts = (params: { search: string; status: string; page: number; limit?: number }) => {
  const q = new URLSearchParams({
    search: params.search,
    status: params.status,
    page: String(params.page),
    limit: String(params.limit ?? 10),
  });
  return request<Page<Product>>("GET", `/api/products?${q}`);
};

export const saveProduct = (id: number | null, body: ProductBody) =>
  id === null
    ? request<{ data: Product }>("POST", "/api/products", body)
    : request<{ data: Product }>("PUT", `/api/products/${id}`, body);

export const deleteProduct = (id: number) => request<void>("DELETE", `/api/products/${id}`);

export const adjustStock = (id: number, body: { type: string; qty: number; note: string }) =>
  request<unknown>("POST", `/api/products/${id}/stock`, body);

export const getMovements = (id: number) =>
  request<{ data: Movement[] }>("GET", `/api/products/${id}/movements`);

export const getOrders = (page: number) => request<Page<Order>>("GET", `/api/orders?page=${page}&limit=10`);

export const getOrder = (id: number) => request<{ data: Order }>("GET", `/api/orders/${id}`);

export const createOrder = (body: OrderBody, kunci: string) =>
  request<{ data: Order }>("POST", "/api/orders", body, { "Idempotency-Key": kunci });

export const cancelOrder = (id: number) => request<{ data: Order }>("POST", `/api/orders/${id}/cancel`);

export const getSummary = () => request<{ data: Summary }>("GET", "/api/summary");

export const getDailySales = () => request<{ data: DailySales }>("GET", "/api/reports/daily-sales");
