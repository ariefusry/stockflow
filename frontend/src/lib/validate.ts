const bulatPositif = (s: string) => /^\d+$/.test(s.trim()) && Number(s) > 0;

export function validateStock(qty: string) {
  return bulatPositif(qty) ? "" : "Quantity must be a whole number greater than 0";
}

export function validateProduct(f: { sku: string; name: string; price: string; stock: string }) {
  if (!f.sku.trim()) return "SKU is required";
  if (!f.name.trim()) return "Name is required";
  if (!/^\d+$/.test(f.price.trim())) return "Price must be a whole number of 0 or more";
  if (f.stock.trim() && !/^\d+$/.test(f.stock.trim())) return "Initial stock must be a whole number of 0 or more";
  return "";
}

export function validateOrder(customer: string, rows: { sku: string; qty: string }[]) {
  if (!customer.trim()) return "Customer name is required";
  if (rows.length === 0) return "Add at least one item";
  const dipakai = new Set<string>();
  for (const [i, r] of rows.entries()) {
    if (!r.sku) return `Item ${i + 1}: choose a product`;
    if (!bulatPositif(r.qty)) return `Item ${i + 1}: quantity must be a whole number greater than 0`;
    if (dipakai.has(r.sku)) return `${r.sku} is listed more than once`;
    dipakai.add(r.sku);
  }
  return "";
}
