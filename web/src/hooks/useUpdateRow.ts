import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateRow } from '../api';

export function useUpdateRow(parquetName?: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ index, columns }: { index: number; columns: Record<string, any> }) =>
      updateRow(index, columns, parquetName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rows', parquetName || 'default'] });
    },
  });
}
