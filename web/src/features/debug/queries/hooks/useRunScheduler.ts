import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

const useRunScheduler = () => {
  return useMutation({
    mutationFn: async () => {
      const { error } = await client.POST('/scheduler/run', {});
      if (error) {
        throw error;
      }
    },
    onSuccess: () => {
      toast.success('Scheduler ran successfully');
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useRunScheduler;
