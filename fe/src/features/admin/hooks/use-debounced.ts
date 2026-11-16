import { useEffect, useState } from "react";

export function useDebounced<TValue>(value: TValue, ms = 300): TValue {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), ms);
    return () => clearTimeout(id);
  }, [value, ms]);
  return debounced;
}
