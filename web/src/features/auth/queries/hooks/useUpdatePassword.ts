import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useUpdatePassword = () => {
  return useMutation({
    mutationFn: async ({
      currentPassword,
      newPassword,
    }: {
      currentPassword: string;
      newPassword: string;
    }) => {
      const { error } = await client.PATCH('/user/password', {
        body: { currentPassword, newPassword },
      });
      if (error) {
        throw error;
      }
    },
    onSuccess: () => {
      toast.success('Password updated.');
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useUpdatePassword;
