import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { getErrorMessage } from '#/lib/errors';
import { fetchAndDownload } from '#/utils/download';

const useExportNewsletter = () => {
  return useMutation({
    mutationFn: async (id: string) => {
      const toastId = toast.loading('Exporting newsletter...');
      try {
        await fetchAndDownload(`/export/${id}`, 'newsletter-export.json');
        toast.success('Export complete!', { id: toastId });
      } catch (error) {
        toast.error(getErrorMessage(error), { id: toastId });
        throw error;
      }
    },
    onError: () => {},
  });
};

export default useExportNewsletter;
