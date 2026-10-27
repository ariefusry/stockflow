export type Product = {
  id: number;
  sku: string;
  name: string;
  price: number;
  stock: number;
  is_active: boolean;
  created_at: string;
  last_movement_at: string | null;
};

export type Movement = {
  id: number;
  product_id: number;
  type: "in" | "out";
  qty: number;
  note: string | null;
  created_at: string;
};

export type OrderItem = {
  sku: string;
  name: string;
  qty: number;
  unit_price: number;
};

export type Order = {
  id: number;
  customer: string;
  status: "CREATED" | "CANCELLED";
  total: number;
  created_at: string;
  cancelled_at: string | null;
  items?: OrderItem[];
};

export type Page<T> = {
  data: T[];
  page: number;
  limit: number;
  total: number;
};

export type Summary = {
  total_products: number;
  active_products: number;
  stock_value: number;
  low_stock: number;
};

export type DailySales = {
  date: string;
  orders: number;
  items_sold: number;
  revenue: number;
};

export type ProductBody = {
  sku: string;
  name: string;
  price: number;
  stock?: number;
  is_active: boolean;
};

export type OrderBody = {
  customer: string;
  items: { sku: string; qty: number }[];
};
