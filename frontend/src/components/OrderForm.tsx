import { useEffect, useRef, useState } from "react";
import { ApiError, createOrder, getProducts, pesan } from "../api";
import { rupiah } from "../lib/format";
import { validateOrder } from "../lib/validate";
import type { Order, Product } from "../types";
import Modal from "./Modal";

type Props = { onClose: () => void; onCreated: (o: Order) => void };
type Baris = { sku: string; qty: string };

export default function OrderForm({ onClose, onCreated }: Props) {
  const [produk, setProduk] = useState<Product[]>([]);
  const [muatError, setMuatError] = useState("");
  const [customer, setCustomer] = useState("");
  const [baris, setBaris] = useState<Baris[]>([{ sku: "", qty: "1" }]);
  const [error, setError] = useState("");
  const [gagal, setGagal] = useState<{ sku: string; available: number } | null>(null);
  const [saving, setSaving] = useState(false);
  const kunci = useRef(crypto.randomUUID());

  useEffect(() => {
    getProducts({ search: "", status: "active", page: 1, limit: 100 })
      .then((r) => setProduk(r.data))
      .catch((e) => setMuatError(pesan(e)));
  }, []);

  useEffect(() => {
    kunci.current = crypto.randomUUID();
  }, [customer, baris]);

  const ubah = (i: number, isi: Partial<Baris>) => {
    setBaris((b) => b.map((x, n) => (n === i ? { ...x, ...isi } : x)));
    setGagal(null);
  };

  const kirim = (e: React.FormEvent) => {
    e.preventDefault();
    const salah = validateOrder(customer, baris);
    if (salah) {
      setError(salah);
      return;
    }
    setSaving(true);
    setError("");
    setGagal(null);
    createOrder(
      { customer, items: baris.map((b) => ({ sku: b.sku, qty: Number(b.qty) })) },
      kunci.current,
    )
      .then((r) => onCreated(r.data))
      .catch((err) => {
        if (err instanceof ApiError && err.code === "INSUFFICIENT_STOCK") {
          const sku = String(err.extra.sku);
          const available = Number(err.extra.available);
          setGagal({ sku, available });
          setError(`${sku} has only ${available} in stock. The order was not saved.`);
        } else {
          setError(pesan(err));
        }
        setSaving(false);
      });
  };

  return (
    <Modal title="New order" onClose={onClose}>
      <form onSubmit={kirim} className="form">
        <label>
          Customer
          <input value={customer} onChange={(e) => setCustomer(e.target.value)} placeholder="e.g. Toko Sumber Rejeki Bekasi" autoFocus />
        </label>

        {muatError && <div className="banner error">Products failed to load: {muatError}</div>}

        {baris.map((b, i) => (
          <div className={gagal && gagal.sku === b.sku ? "item-row bad" : "item-row"} key={i}>
            <select value={b.sku} onChange={(e) => ubah(i, { sku: e.target.value })}>
              <option value="">Choose product</option>
              {produk.map((p) => (
                <option key={p.id} value={p.sku}>
                  {p.sku} - {p.name} ({rupiah(p.price)}, stock {p.stock})
                </option>
              ))}
            </select>
            <input value={b.qty} onChange={(e) => ubah(i, { qty: e.target.value })} inputMode="numeric" aria-label="Quantity" />
            <button type="button" className="btn small" onClick={() => setBaris((x) => x.filter((_, n) => n !== i))}>
              Remove
            </button>
            {gagal && gagal.sku === b.sku && <span className="item-note">Available: {gagal.available}</span>}
          </div>
        ))}

        <button type="button" className="btn small" onClick={() => setBaris((x) => [...x, { sku: "", qty: "1" }])}>
          Add item
        </button>

        {error && <div className="banner error">{error}</div>}
        <button className="btn primary" disabled={saving}>{saving ? "Saving..." : "Create order"}</button>
      </form>
    </Modal>
  );
}
