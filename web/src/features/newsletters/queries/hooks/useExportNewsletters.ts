import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { getErrorMessage } from '#/lib/errors';
import { fetchAndDownload } from '#/utils/download';

const useExportNewsletters = () => {
  return useMutation({
    mutationFn: async () => {
      const toastId = toast.loading('Exporting newsletters...');
      try {
        await fetchAndDownload('/export', 'newsletters-export.json');
        toast.success('Export complete!', { id: toastId });
      } catch (error) {
        toast.error(getErrorMessage(error), { id: toastId });
        throw error;
      }
    },
  });
};

export default useExportNewsletters;
