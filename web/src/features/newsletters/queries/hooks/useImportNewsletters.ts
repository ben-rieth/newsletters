import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { pluralize } from '../../lib/format';
import { newslettersKeys } from '../newsletters';
import type { components } from '#/api/schema';
import client from '#/api/client';
import { getErrorMessage } from '#/lib/errors';

type NewslettersExport = components['schemas']['NewslettersExport'];
type ImportResult = components['schemas']['ImportResult'];

const describeImport = (result: ImportResult) => {
  const newsletterIds = result.newsletterIds ?? [];
  const pendingFeeds = result.pendingFeeds;
  const parts = [`Imported ${pluralize(newsletterIds.length, 'newsletter')}.`];
  if (pendingFeeds > 0) {
    parts.push(
      `${pluralize(pendingFeeds, 'feed')} will be added in the background.`,
    );
  }
  return parts.join(' ');
};

const describeImportError = (error: unknown) => {
  if (error instanceof SyntaxError) {
    return "That file isn't valid JSON.";
  }
  if (typeof error === 'object' && error !== null && 'errors' in error) {
    const firstProblem = (error as components['schemas']['ErrorModel'])
      .errors?.[0];
    if (firstProblem?.message) {
      const location = (firstProblem.location ?? '').replace(/^body\./, '');
      return `Invalid export file: ${location} ${firstProblem.message}`.trim();
    }
  }
  return getErrorMessage(error);
};

const useImportNewsletters = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (file: File) => {
      const body = JSON.parse(await file.text()) as NewslettersExport;
      const { data, error, response } = await client.POST('/import', {
        body,
      });
      if (error || !response.ok) {
        throw error ?? new Error(response.statusText || 'Import failed');
      }
      return data;
    },
    onSuccess: async (result) => {
      toast.success(describeImport(result));
      await queryClient.invalidateQueries({ queryKey: newslettersKeys.all });
    },
    onError: (error) => {
      toast.error(describeImportError(error));
    },
  });
};

export default useImportNewsletters;
