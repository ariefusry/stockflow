import { useEffect, useState } from "react";
import { getMovements, pesan } from "../api";
import { tanggal } from "../lib/format";
import type { Movement, Product } from "../types";
import Modal from "./Modal";

export default function MovementsModal({ product, onClose }: { product: Product; onClose: () => void }) {
  const [data, setData] = useState<Movement[] | null>(null);
  const [error, setError] = useState("");
  const [coba, setCoba] = useState(0);

  useEffect(() => {
    setData(null);
    setError("");
    getMovements(product.id)
      .then((r) => setData(r.data))
      .catch((e) => setError(pesan(e)));
  }, [product.id, coba]);

  return (
    <Modal title={`History: ${product.name}`} onClose={onClose}>
      {error && (
        <div className="banner error">
          {error}
          <button className="btn small" onClick={() => setCoba((n) => n + 1)}>Retry</button>
        </div>
      )}
      {!error && data === null && <div className="skeleton block" />}
      {data && data.length === 0 && <p className="muted">No stock movements yet.</p>}
      {data && data.length > 0 && (
        <table>
          <thead>
            <tr><th>Time</th><th>Type</th><th className="num">Qty</th><th>Note</th></tr>
          </thead>
          <tbody>
            {data.map((m) => (
              <tr key={m.id}>
                <td>{tanggal(m.created_at)}</td>
                <td><span className={m.type === "in" ? "badge ok" : "badge out"}>{m.type === "in" ? "In" : "Out"}</span></td>
                <td className="num">{m.qty}</td>
                <td>{m.note ?? "-"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </Modal>
  );
}
