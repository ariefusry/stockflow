import { describe, expect, it } from "vitest";
import { validateOrder, validateProduct, validateStock } from "./validate";

describe("validateStock", () => {
  it("menolak kosong, nol, negatif, dan desimal", () => {
    for (const s of ["", "0", "-3", "2.5", "abc"]) {
      expect(validateStock(s)).not.toBe("");
    }
  });

  it("menerima bilangan bulat positif", () => {
    expect(validateStock("12")).toBe("");
  });
});

describe("validateProduct", () => {
  const ok = { sku: "BRS-PRM-5KG", name: "Beras Premium", price: "74500", stock: "" };

  it("menerima data lengkap", () => {
    expect(validateProduct(ok)).toBe("");
  });

  it("menolak sku, nama, atau harga yang tidak valid", () => {
    expect(validateProduct({ ...ok, sku: " " })).toMatch(/SKU/);
    expect(validateProduct({ ...ok, name: "" })).toMatch(/Name/);
    expect(validateProduct({ ...ok, price: "-1" })).toMatch(/Price/);
    expect(validateProduct({ ...ok, price: "12.5" })).toMatch(/Price/);
  });
});

describe("validateOrder", () => {
  const baris = [{ sku: "MIE-IND-GRG", qty: "10" }];

  it("menerima order yang benar", () => {
    expect(validateOrder("Warung Bu Siti", baris)).toBe("");
  });

  it("menolak customer kosong dan item kosong", () => {
    expect(validateOrder("  ", baris)).toMatch(/Customer/);
    expect(validateOrder("Warung Bu Siti", [])).toMatch(/at least one/);
  });

  it("menolak baris tanpa produk, qty salah, atau sku ganda", () => {
    expect(validateOrder("Toko A", [{ sku: "", qty: "1" }])).toMatch(/Item 1/);
    expect(validateOrder("Toko A", [{ sku: "X", qty: "0" }])).toMatch(/Item 1/);
    expect(validateOrder("Toko A", [{ sku: "X", qty: "1" }, { sku: "X", qty: "2" }])).toMatch(/more than once/);
  });
});
