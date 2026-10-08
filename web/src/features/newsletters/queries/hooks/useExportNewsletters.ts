import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { getErrorMessage } from '#/lib/errors';
import { fetchAndDownload } from '#/utils/download';

const useExportNewsletters = () => {
  return useMutation({
    mutationFn: async (ids?: string[]) => {
      const toastId = toast.loading(
        ids?.length === 1
          ? 'Exporting newsletter...'
          : 'Exporting newsletters...',
      );
      try {
        const query = ids?.length
          ? `?${new URLSearchParams({ ids: ids.join(',') })}`
          : '';
        await fetchAndDownload(`/export${query}`, 'newsletters-export.json');
        toast.success('Export complete!', { id: toastId });
      } catch (error) {
        toast.error(getErrorMessage(error), { id: toastId });
        throw error;
      }
    },
  });
};

export default useExportNewsletters;
