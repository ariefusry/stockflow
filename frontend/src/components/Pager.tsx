type Props = { page: number; limit: number; total: number; onPage: (p: number) => void };

export default function Pager({ page, limit, total, onPage }: Props) {
  const jumlah = Math.max(1, Math.ceil(total / limit));
  return (
    <div className="pager">
      <span>{total} total</span>
      <button className="btn small" disabled={page <= 1} onClick={() => onPage(page - 1)}>Prev</button>
      <span>Page {page} of {jumlah}</span>
      <button className="btn small" disabled={page >= jumlah} onClick={() => onPage(page + 1)}>Next</button>
    </div>
  );
}
