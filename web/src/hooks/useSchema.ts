import { useQuery } from '@tanstack/react-query';
import { fetchSchema, fetchParquets } from '../api';

export function useSchema(parquetName?: string) {
  return useQuery({
    queryKey: ['schema', parquetName || 'default'],
    queryFn: () => fetchSchema(parquetName),
  });
}

export function useParquets() {
  return useQuery({
    queryKey: ['parquets'],
    queryFn: fetchParquets,
  });
}
