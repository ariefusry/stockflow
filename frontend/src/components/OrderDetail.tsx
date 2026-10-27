import { useEffect, useState } from "react";
import { cancelOrder, getOrder, pesan } from "../api";
import { rupiah, tanggal } from "../lib/format";
import type { Order } from "../types";
import Modal from "./Modal";

type Props = { id: number; onClose: () => void; onChanged: () => void };

export default function OrderDetail({ id, onClose, onChanged }: Props) {
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState("");
  const [coba, setCoba] = useState(0);
  const [batal, setBatal] = useState(false);

  useEffect(() => {
    getOrder(id)
      .then((r) => {
        setOrder(r.data);
        setError("");
      })
      .catch((e) => setError(pesan(e)));
  }, [id, coba]);

  const batalkan = () => {
    if (!window.confirm(`Cancel order #${id}? Stock will be returned to the products.`)) return;
    setBatal(true);
    setError("");
    cancelOrder(id)
      .then((r) => {
        setOrder(r.data);
        onChanged();
      })
      .catch((e) => setError(pesan(e)))
      .finally(() => setBatal(false));
  };

  return (
    <Modal title={`Order #${id}`} onClose={onClose}>
      {error && (
        <div className="banner error">
          {error}
          {!order && <button className="btn small" onClick={() => setCoba((n) => n + 1)}>Retry</button>}
        </div>
      )}
      {!order && !error && <div className="skeleton block" />}
      {order && (
        <>
          <dl className="meta">
            <dt>Customer</dt>
            <dd>{order.customer}</dd>
            <dt>Status</dt>
            <dd><span className={order.status === "CANCELLED" ? "badge off" : "badge ok"}>{order.status}</span></dd>
            <dt>Created</dt>
            <dd>{tanggal(order.created_at)}</dd>
            {order.cancelled_at && (
              <>
                <dt>Cancelled</dt>
                <dd>{tanggal(order.cancelled_at)}</dd>
              </>
            )}
          </dl>
          <table>
            <thead>
              <tr><th>SKU</th><th>Product</th><th className="num">Qty</th><th className="num">Unit price</th><th className="num">Subtotal</th></tr>
            </thead>
            <tbody>
              {(order.items ?? []).map((it) => (
                <tr key={it.sku}>
                  <td className="mono">{it.sku}</td>
                  <td>{it.name}</td>
                  <td className="num">{it.qty}</td>
                  <td className="num">{rupiah(it.unit_price)}</td>
                  <td className="num">{rupiah(it.unit_price * it.qty)}</td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr><td colSpan={4}>Total</td><td className="num">{rupiah(order.total)}</td></tr>
            </tfoot>
          </table>
          {order.status === "CREATED" && (
            <button className="btn danger" disabled={batal} onClick={batalkan}>
              {batal ? "Cancelling..." : "Cancel order"}
            </button>
          )}
        </>
      )}
    </Modal>
  );
}
