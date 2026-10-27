import { useState } from "react";
import Summary from "./components/Summary";
import OrdersPage from "./pages/OrdersPage";
import ProductsPage from "./pages/ProductsPage";

export default function App() {
  const [tab, setTab] = useState<"products" | "orders">("products");
  const [ver, setVer] = useState(0);
  const refresh = () => setVer((v) => v + 1);

  return (
    <div className="app">
      <header className="top">
        <h1>StockFlow</h1>
        <nav className="tabs">
          <button className={tab === "products" ? "tab active" : "tab"} onClick={() => setTab("products")}>Products</button>
          <button className={tab === "orders" ? "tab active" : "tab"} onClick={() => setTab("orders")}>Orders</button>
        </nav>
      </header>
      <Summary ver={ver} />
      {tab === "products" ? <ProductsPage ver={ver} refresh={refresh} /> : <OrdersPage ver={ver} refresh={refresh} />}
    </div>
  );
}
