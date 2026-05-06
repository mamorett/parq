import { useMutation, useQueryClient } from '@tanstack/react-query';
import { deleteRow } from '../api';

export function useDeleteRow(parquetName?: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ index }: { index: number }) =>
      deleteRow(index, parquetName),
    onSuccess: async () => {
      await queryClient.invalidateQueries();
    },
  });
}
