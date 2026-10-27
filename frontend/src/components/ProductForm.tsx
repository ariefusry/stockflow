import { useState } from "react";
import { pesan, saveProduct } from "../api";
import { validateProduct } from "../lib/validate";
import type { Product } from "../types";
import Modal from "./Modal";

type Props = { product: Product | null; onClose: () => void; onDone: () => void };

export default function ProductForm({ product, onClose, onDone }: Props) {
  const [sku, setSku] = useState(product?.sku ?? "");
  const [name, setName] = useState(product?.name ?? "");
  const [price, setPrice] = useState(product ? String(product.price) : "");
  const [stock, setStock] = useState("");
  const [aktif, setAktif] = useState(product?.is_active ?? true);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  const kirim = (e: React.FormEvent) => {
    e.preventDefault();
    const salah = validateProduct({ sku, name, price, stock });
    if (salah) {
      setError(salah);
      return;
    }
    setSaving(true);
    setError("");
    saveProduct(product?.id ?? null, {
      sku,
      name,
      price: Number(price),
      stock: product || !stock.trim() ? undefined : Number(stock),
      is_active: aktif,
    })
      .then(() => {
        onDone();
        onClose();
      })
      .catch((err) => {
        setError(pesan(err));
        setSaving(false);
      });
  };

  return (
    <Modal title={product ? `Edit ${product.sku}` : "New product"} onClose={onClose}>
      <form onSubmit={kirim} className="form">
        <label>
          SKU
          <input value={sku} onChange={(e) => setSku(e.target.value)} autoFocus />
        </label>
        <label>
          Name
          <input value={name} onChange={(e) => setName(e.target.value)} />
        </label>
        <label>
          Price (IDR)
          <input value={price} onChange={(e) => setPrice(e.target.value)} inputMode="numeric" />
        </label>
        {!product && (
          <label>
            Initial stock
            <input value={stock} onChange={(e) => setStock(e.target.value)} inputMode="numeric" placeholder="0" />
          </label>
        )}
        <label className="check">
          <input type="checkbox" checked={aktif} onChange={(e) => setAktif(e.target.checked)} />
          Active
        </label>
        {error && <div className="banner error">{error}</div>}
        <button className="btn primary" disabled={saving}>{saving ? "Saving..." : "Save"}</button>
      </form>
    </Modal>
  );
}
