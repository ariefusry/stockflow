import { useEffect, useState } from "react";
import { getDailySales, getSummary, pesan } from "../api";
import { rupiah } from "../lib/format";
import type { DailySales, Summary as Ringkasan } from "../types";

export default function Summary({ ver }: { ver: number }) {
  const [ringkas, setRingkas] = useState<Ringkasan | null>(null);
  const [harian, setHarian] = useState<DailySales | null>(null);
  const [error, setError] = useState("");
  const [coba, setCoba] = useState(0);

  useEffect(() => {
    let batal = false;
    Promise.all([getSummary(), getDailySales()])
      .then(([s, d]) => {
        if (batal) return;
        setRingkas(s.data);
        setHarian(d.data);
        setError("");
      })
      .catch((e) => {
        if (!batal) setError(pesan(e));
      });
    return () => {
      batal = true;
    };
  }, [ver, coba]);

  if (error) {
    return (
      <div className="banner error">
        Summary failed to load: {error}
        <button className="btn small" onClick={() => setCoba((n) => n + 1)}>Retry</button>
      </div>
    );
  }

  const kartu = [
    { label: "Total products", nilai: ringkas && String(ringkas.total_products) },
    { label: "Active products", nilai: ringkas && String(ringkas.active_products) },
    { label: "Stock value", nilai: ringkas && rupiah(ringkas.stock_value) },
    { label: "Low stock (10 or fewer)", nilai: ringkas && String(ringkas.low_stock), warn: (ringkas?.low_stock ?? 0) > 0 },
    { label: "Today's sales", nilai: harian && rupiah(harian.revenue), sub: harian && `${harian.orders} orders, ${harian.items_sold} items` },
  ];

  return (
    <div className="cards">
      {kartu.map((k) => (
        <div className={k.warn ? "card warn" : "card"} key={k.label}>
          <span className="card-label">{k.label}</span>
          {k.nilai === null ? <span className="skeleton card-skel" /> : <strong>{k.nilai}</strong>}
          {k.sub && <span className="card-sub">{k.sub}</span>}
        </div>
      ))}
    </div>
  );
}
