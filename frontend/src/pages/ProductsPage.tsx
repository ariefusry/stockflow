import { useEffect, useState } from "react";
import { deleteProduct, getProducts, pesan } from "../api";
import MovementsModal from "../components/MovementsModal";
import Pager from "../components/Pager";
import ProductForm from "../components/ProductForm";
import StockModal from "../components/StockModal";
import { useDebounce } from "../hooks/useDebounce";
import { ago, rupiah } from "../lib/format";
import type { Page, Product } from "../types";

type Props = { ver: number; refresh: () => void };

const filter = [
  { nilai: "", label: "All" },
  { nilai: "active", label: "Active" },
  { nilai: "inactive", label: "Inactive" },
  { nilai: "low", label: "Low stock" },
];

function badge(p: Product) {
  if (!p.is_active) return <span className="badge off">Inactive</span>;
  if (p.stock <= 10) return <span className="badge warn">Low stock</span>;
  return <span className="badge ok">Active</span>;
}

export default function ProductsPage({ ver, refresh }: Props) {
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");
  const [page, setPage] = useState(1);
  const [hasil, setHasil] = useState<Page<Product> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [coba, setCoba] = useState(0);
  const [aksiError, setAksiError] = useState("");
  const [stokProduct, setStokProduct] = useState<Product | null>(null);
  const [riwayat, setRiwayat] = useState<Product | null>(null);
  const [form, setForm] = useState<Product | "baru" | null>(null);
  const cari = useDebounce(search, 300);

  useEffect(() => {
    let batal = false;
    setLoading(true);
    setError("");
    getProducts({ search: cari, status, page })
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
  }, [cari, status, page, ver, coba]);

  const hapus = (p: Product) => {
    if (!window.confirm(`Delete ${p.name}?`)) return;
    setAksiError("");
    deleteProduct(p.id)
      .then(refresh)
      .catch((e) => setAksiError(pesan(e)));
  };

  const data = hasil?.data ?? [];

  return (
    <section>
      <div className="toolbar">
        <input
          className="search"
          placeholder="Search SKU or name"
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            setPage(1);
          }}
        />
        <div className="tabs">
          {filter.map((f) => (
            <button
              key={f.nilai}
              className={status === f.nilai ? "tab active" : "tab"}
              onClick={() => {
                setStatus(f.nilai);
                setPage(1);
              }}
            >
              {f.label}
            </button>
          ))}
        </div>
        <button className="btn primary" onClick={() => setForm("baru")}>New product</button>
      </div>

      {aksiError && <div className="banner error">{aksiError}</div>}

      {error ? (
        <div className="banner error">
          {error}
          <button className="btn small" onClick={() => setCoba((n) => n + 1)}>Retry</button>
        </div>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>SKU</th>
                <th>Name</th>
                <th className="num">Price</th>
                <th className="num">Stock</th>
                <th>Status</th>
                <th>Last movement</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {loading &&
                Array.from({ length: 6 }, (_, i) => (
                  <tr key={i}>
                    <td colSpan={7}><div className="skeleton row" /></td>
                  </tr>
                ))}
              {!loading && data.length === 0 && (
                <tr>
                  <td colSpan={7} className="empty">No products match your search or filter.</td>
                </tr>
              )}
              {!loading &&
                data.map((p) => (
                  <tr key={p.id}>
                    <td className="mono">{p.sku}</td>
                    <td>{p.name}</td>
                    <td className="num">{rupiah(p.price)}</td>
                    <td className="num">{p.stock}</td>
                    <td>{badge(p)}</td>
                    <td>{ago(p.last_movement_at)}</td>
                    <td className="actions">
                      <button className="btn small" onClick={() => setStokProduct(p)}>Adjust stock</button>
                      <button className="btn small" onClick={() => setRiwayat(p)}>History</button>
                      <button className="btn small" onClick={() => setForm(p)}>Edit</button>
                      <button className="btn small danger" onClick={() => hapus(p)}>Delete</button>
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      )}

      {hasil && !error && <Pager page={hasil.page} limit={hasil.limit} total={hasil.total} onPage={setPage} />}

      {stokProduct && <StockModal product={stokProduct} onClose={() => setStokProduct(null)} onDone={refresh} />}
      {riwayat && <MovementsModal product={riwayat} onClose={() => setRiwayat(null)} />}
      {form && <ProductForm product={form === "baru" ? null : form} onClose={() => setForm(null)} onDone={refresh} />}
    </section>
  );
}
