import client from '#/api/client';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { newslettersKeys } from '../newsletters';
import type { components } from '#/api/schema';
import { getErrorMessage } from '#/lib/errors';

const useUpdateNewsletter = (newsletterId: string, onSuccess?: () => void) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (
      formValues: components['schemas']['SubmittableNewsletterFields'],
    ) => {
      const { error } = await client.PUT('/newsletters/{newsletterId}', {
        params: { path: { newsletterId } },
        body: formValues,
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: newslettersKeys.all });
      onSuccess?.();
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useUpdateNewsletter;
