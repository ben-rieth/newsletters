import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
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

export type { NewslettersExport };

export const parseExportFile = async (
  file: File,
): Promise<NewslettersExport> => {
  let parsed: unknown;
  try {
    parsed = JSON.parse(await file.text());
  } catch {
    throw new Error("That file isn't valid JSON.");
  }
  const newsletters = (parsed as Partial<NewslettersExport> | null)
    ?.newsletters;
  if (!Array.isArray(newsletters)) {
    throw new Error("That file isn't a Slowfeed export.");
  }
  if (newsletters.length === 0) {
    throw new Error('That export has no newsletters in it.');
  }
  return parsed as NewslettersExport;
};

const describeImportError = (error: unknown) => {
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
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (body: NewslettersExport) => {
      const { data, error, response } = await client.POST('/import', {
        body,
      });
      if (error || !response.ok) {
        throw error ?? new Error(response.statusText || 'Import failed');
      }
      return data;
    },
    onSuccess: async (result) => {
      const [onlyId, ...rest] = result.newsletterIds ?? [];
      toast.success(describeImport(result), {
        action:
          onlyId && rest.length === 0
            ? {
                label: 'View',
                onClick: () =>
                  navigate({
                    to: '/newsletters/$newsletterId',
                    params: { newsletterId: onlyId },
                  }),
              }
            : undefined,
      });
      await queryClient.invalidateQueries({ queryKey: newslettersKeys.all });
    },
    onError: (error) => {
      toast.error(describeImportError(error));
    },
  });
};

export default useImportNewsletters;
