import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useUpdateEmail = () => {
  return useMutation({
    mutationFn: async ({
      email,
      password,
    }: {
      email: string;
      password: string;
    }) => {
      const { error } = await client.PATCH('/user/email', {
        body: { email, password },
      });
      if (error) {
        throw error;
      }
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useUpdateEmail;
