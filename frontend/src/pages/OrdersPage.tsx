import { useEffect, useState } from "react";
import { getOrders, pesan } from "../api";
import OrderDetail from "../components/OrderDetail";
import OrderForm from "../components/OrderForm";
import Pager from "../components/Pager";
import { rupiah, tanggal } from "../lib/format";
import type { Order, Page } from "../types";

type Props = { ver: number; refresh: () => void };

export default function OrdersPage({ ver, refresh }: Props) {
  const [page, setPage] = useState(1);
  const [hasil, setHasil] = useState<Page<Order> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [coba, setCoba] = useState(0);
  const [buat, setBuat] = useState(false);
  const [pilih, setPilih] = useState<number | null>(null);

  useEffect(() => {
    let batal = false;
    setLoading(true);
    setError("");
    getOrders(page)
      .then((r) => {
        if (!batal) setHasil(r);
      })
      .catch((e) => {
        if (!batal) setError(pesan(e));
      })
      .finally(() => {
        if (!batal) setLoading(false);
      });
    return () => {
      batal = true;
    };
  }, [page, ver, coba]);

  const data = hasil?.data ?? [];

  return (
    <section>
      <div className="toolbar">
        <h2>Sales orders</h2>
        <button className="btn primary" onClick={() => setBuat(true)}>New order</button>
      </div>

      {error ? (
        <div className="banner error">
          {error}
          <button className="btn small" onClick={() => setCoba((n) => n + 1)}>Retry</button>
        </div>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr><th>No.</th><th>Customer</th><th>Status</th><th className="num">Total</th><th>Created</th><th></th></tr>
            </thead>
            <tbody>
              {loading &&
                Array.from({ length: 5 }, (_, i) => (
                  <tr key={i}>
                    <td colSpan={6}><div className="skeleton row" /></td>
                  </tr>
                ))}
              {!loading && data.length === 0 && (
                <tr>
                  <td colSpan={6} className="empty">No orders yet. Create the first one.</td>
                </tr>
              )}
              {!loading &&
                data.map((o) => (
                  <tr key={o.id}>
                    <td className="mono">#{o.id}</td>
                    <td>{o.customer}</td>
                    <td><span className={o.status === "CANCELLED" ? "badge off" : "badge ok"}>{o.status}</span></td>
                    <td className="num">{rupiah(o.total)}</td>
                    <td>{tanggal(o.created_at)}</td>
                    <td className="actions"><button className="btn small" onClick={() => setPilih(o.id)}>Detail</button></td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      )}

      {hasil && !error && <Pager page={hasil.page} limit={hasil.limit} total={hasil.total} onPage={setPage} />}

      {buat && (
        <OrderForm
          onClose={() => setBuat(false)}
          onCreated={(o) => {
            setBuat(false);
            setPilih(o.id);
            refresh();
          }}
        />
      )}
      {pilih !== null && <OrderDetail id={pilih} onClose={() => setPilih(null)} onChanged={refresh} />}
    </section>
  );
}
