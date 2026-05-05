import { useQuery } from '@tanstack/react-query';
import { fetchRows } from '../api';

export function useRows(params: {
  page: number;
  size: number;
  sort?: string;
  order?: string;
  search?: string;
  search_col?: string;
  subdirs?: string[];
  parquet?: string;
}) {
  return useQuery({
    queryKey: ['rows', params.parquet || 'default', params],
    queryFn: () => fetchRows(params),
    enabled: !!params.size,
  });
}
