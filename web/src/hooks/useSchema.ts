import { useQuery } from '@tanstack/react-query';
import { fetchSchema } from '../api';

export function useSchema() {
  return useQuery({
    queryKey: ['schema'],
    queryFn: fetchSchema,
  });
}
