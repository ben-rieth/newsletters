import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';
import { userKeys } from '../user';

const useVerifyEmailUpdate = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (code: string) => {
      const { error } = await client.POST('/user/email/verify', {
        body: { code },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.me });
      toast.success('Email updated.');
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useVerifyEmailUpdate;
