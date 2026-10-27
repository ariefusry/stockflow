import { useEffect, useState } from "react";

export function useDebounce<T>(nilai: T, ms = 300) {
  const [hasil, setHasil] = useState(nilai);
  useEffect(() => {
    const t = setTimeout(() => setHasil(nilai), ms);
    return () => clearTimeout(t);
  }, [nilai, ms]);
  return hasil;
}
