import { describe, expect, it } from "vitest";
import { ago, rupiah } from "./format";

describe("rupiah", () => {
  it("memakai titik sebagai pemisah ribuan", () => {
    expect(rupiah(1234567)).toBe("Rp 1.234.567");
    expect(rupiah(0)).toBe("Rp 0");
  });
});

describe("ago", () => {
  const sekarang = new Date("2026-10-27T10:00:00Z").getTime();

  it("tanpa data menghasilkan tanda strip", () => {
    expect(ago(null, sekarang)).toBe("-");
  });

  it("menghitung menit, jam, dan hari", () => {
    expect(ago("2026-10-27T09:59:30Z", sekarang)).toBe("just now");
    expect(ago("2026-10-27T09:45:00Z", sekarang)).toBe("15m ago");
    expect(ago("2026-10-27T08:00:00Z", sekarang)).toBe("2h ago");
    expect(ago("2026-10-24T10:00:00Z", sekarang)).toBe("3d ago");
  });
});
