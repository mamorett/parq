import { useSearchParams } from 'react-router-dom';
import { useMemo } from 'react';

export function useUrlState() {
  const [searchParams, setSearchParams] = useSearchParams();

  const state = useMemo(() => {
    return {
      page: parseInt(searchParams.get('page') || '1', 10),
      size: parseInt(searchParams.get('size') || '0', 10), // 0 means use default from schema
      sort: searchParams.get('sort') || undefined,
      order: (searchParams.get('order') as 'asc' | 'desc') || undefined,
      search: searchParams.get('search') || '',
      search_col: searchParams.get('search_col') || undefined,
      subdirs: searchParams.getAll('subdir'),
    };
  }, [searchParams]);

  const updateState = (updates: Partial<typeof state>) => {
    const nextParams = new URLSearchParams(searchParams);
    Object.entries(updates).forEach(([key, value]) => {
      if (value === undefined || value === '' || value === null) {
        nextParams.delete(key);
      } else if (Array.isArray(value)) {
        nextParams.delete(key);
        value.forEach(v => nextParams.append(key, v));
      } else {
        nextParams.set(key, value.toString());
      }
    });
    setSearchParams(nextParams);
  };

  return { state, updateState };
}
