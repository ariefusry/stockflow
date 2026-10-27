const angka = new Intl.NumberFormat("id-ID");

export const rupiah = (n: number) => "Rp " + angka.format(n);

export function ago(iso: string | null, sekarang = Date.now()) {
  if (!iso) return "-";
  const detik = Math.floor((sekarang - new Date(iso).getTime()) / 1000);
  if (detik < 60) return "just now";
  const menit = Math.floor(detik / 60);
  if (menit < 60) return `${menit}m ago`;
  const jam = Math.floor(menit / 60);
  if (jam < 24) return `${jam}h ago`;
  return `${Math.floor(jam / 24)}d ago`;
}

export const tanggal = (iso: string) =>
  new Date(iso).toLocaleString("id-ID", { dateStyle: "medium", timeStyle: "short" });
