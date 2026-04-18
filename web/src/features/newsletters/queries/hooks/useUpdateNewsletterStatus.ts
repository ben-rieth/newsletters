import client from '#/api/client';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { newslettersKeys } from '../newsletters';

const useUpdateNewsletterStatus = (onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ newsletterId, status }: { newsletterId: string; status: 'active' | 'inactive' }) => {
      const { error } = await client.PATCH('/newsletter/{newsletterId}/status', {
        params: { path: { newsletterId } },
        body: { status },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: (_data, { newsletterId }) => {
      queryClient.invalidateQueries({ queryKey: newslettersKeys.detail(newsletterId) });
      queryClient.invalidateQueries({ queryKey: newslettersKeys.all });
      onSuccess?.();
    },
  });
};

export default useUpdateNewsletterStatus;
