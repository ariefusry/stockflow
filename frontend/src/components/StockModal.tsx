import { useState } from "react";
import { ApiError, adjustStock, pesan } from "../api";
import { validateStock } from "../lib/validate";
import type { Product } from "../types";
import Modal from "./Modal";

type Props = { product: Product; onClose: () => void; onDone: () => void };

export default function StockModal({ product, onClose, onDone }: Props) {
  const [type, setType] = useState("in");
  const [qty, setQty] = useState("");
  const [note, setNote] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  const kirim = (e: React.FormEvent) => {
    e.preventDefault();
    const salah = validateStock(qty);
    if (salah) {
      setError(salah);
      return;
    }
    setSaving(true);
    setError("");
    adjustStock(product.id, { type, qty: Number(qty), note })
      .then(() => {
        onDone();
        onClose();
      })
      .catch((err) => {
        if (err instanceof ApiError && err.code === "INSUFFICIENT_STOCK") {
          setError(`Insufficient stock: only ${String(err.extra.available)} available`);
        } else {
          setError(pesan(err));
        }
        setSaving(false);
      });
  };

  return (
    <Modal title={`Adjust stock: ${product.name}`} onClose={onClose}>
      <p className="muted">{product.sku} - current stock {product.stock}</p>
      <form onSubmit={kirim} className="form">
        <label>
          Direction
          <select value={type} onChange={(e) => setType(e.target.value)}>
            <option value="in">Stock in</option>
            <option value="out">Stock out</option>
          </select>
        </label>
        <label>
          Quantity
          <input value={qty} onChange={(e) => setQty(e.target.value)} inputMode="numeric" autoFocus />
        </label>
        <label>
          Note
          <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="Optional, e.g. stock opname" />
        </label>
        {error && <div className="banner error">{error}</div>}
        <button className="btn primary" disabled={saving}>
          {saving ? "Saving..." : "Save"}
        </button>
      </form>
    </Modal>
  );
}
