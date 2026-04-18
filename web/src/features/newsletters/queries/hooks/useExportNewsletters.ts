import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { getErrorMessage } from '#/lib/errors';
import { fetchAndDownload } from '#/utils/download';

const useExportNewsletters = () => {
  return useMutation({
    mutationFn: async () => {
      await fetchAndDownload('/export', 'newsletters-export.json');
    },
    onError: (error) => {
      toast.error(getErrorMessage(error));
    },
  });
};

export default useExportNewsletters;
