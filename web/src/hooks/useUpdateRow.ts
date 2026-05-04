import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateRow } from '../api';

export function useUpdateRow() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ index, columns }: { index: number; columns: Record<string, any> }) =>
      updateRow(index, columns),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rows'] });
    },
  });
}
