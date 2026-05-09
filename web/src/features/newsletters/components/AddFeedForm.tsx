import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import { Loader2, TriangleAlert } from 'lucide-react';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { feedMetadataOptions } from '../queries/feeds';

const feedUrlSchema = z.object({
  url: z
    .string()
    .min(1, 'URL is required')
    .url('Must be a valid URL')
    .startsWith('https://', 'Must be an HTTPS URL'),
  alias: z.string(),
});

export type AddFeedFormValues = {
  url: string;
  alias: string;
};

type Props = {
  onSubmit: (values: AddFeedFormValues) => Promise<void>;
  isPending?: boolean;
  submitLabel?: string;
  error?: string;
};

export const AddFeedForm = ({
  onSubmit,
  isPending = false,
  submitLabel = 'Save',
  error,
}: Props) => {
  const [fetchUrl, setFetchUrl] = useState('');

  const metadataQuery = useQuery({
    ...feedMetadataOptions(fetchUrl),
    enabled: !!fetchUrl,
  });

  const canSubmit = !!metadataQuery.data && !isPending;

  const form = useForm({
    defaultValues: { url: '', alias: '' },
    validators: { onChange: feedUrlSchema },
    onSubmit: async ({ value }) => {
      if (!metadataQuery.data) {
        return;
      }
      await onSubmit({
        url: value.url,
        alias: value.alias,
      });
    },
  });

  return (
    <form
      className="space-y-3"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <form.Field name="url">
        {(field) => (
          <FormField field={field} label="Feed URL">
            <Input
              id={field.name}
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={(e) => {
                field.handleBlur();
                const result = feedUrlSchema.shape.url.safeParse(
                  e.target.value,
                );
                if (result.success) {
                  setFetchUrl(e.target.value);
                }
              }}
              placeholder="https://example.com/feed.xml"
              aria-invalid={
                field.state.meta.isBlurred && field.state.meta.errors.length > 0
              }
            />
          </FormField>
        )}
      </form.Field>

      {metadataQuery.isFetching && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          Fetching feed info…
        </div>
      )}

      {metadataQuery.isError && !metadataQuery.isFetching && (
        <FieldError>
          Could not fetch feed info. Check the URL and try again.
        </FieldError>
      )}

      {metadataQuery.data && (
        <>
          <div className="rounded-md border bg-muted/40 p-3 space-y-1">
            <p className="text-sm font-medium">
              {metadataQuery.data.metadata.Title}
            </p>
            {metadataQuery.data.metadata.Description && (
              <p className="text-xs text-muted-foreground">
                {metadataQuery.data.metadata.Description}
              </p>
            )}
          </div>

          {metadataQuery.data.existingRecievedFeeds &&
            metadataQuery.data.existingRecievedFeeds.length > 0 && (
              <div className="flex gap-2.5 rounded-md border border-yellow-500/40 bg-yellow-500/10 p-3 text-sm text-yellow-700 dark:text-yellow-400">
                <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0" />
                <div className="space-y-1">
                  <p className="font-medium">You already receive this feed</p>
                  <ul className="list-disc pl-4 text-xs">
                    {metadataQuery.data.existingRecievedFeeds.map((f) => (
                      <li key={`${f.newsletterName}-${f.alias}`}>
                        In newsletter{' '}
                        <span className="font-medium">
                          "{f.newsletterName}"
                        </span>{' '}
                        {f.alias !== '' && `(as ${f.alias})`}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            )}

          <form.Field name="alias">
            {(field) => (
              <FormField field={field} label="Custom name (optional)">
                <Input
                  id={field.name}
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  placeholder={metadataQuery.data.metadata.Title}
                />
              </FormField>
            )}
          </form.Field>
        </>
      )}

      {error && <FieldError>{error}</FieldError>}

      <Button type="submit" disabled={!canSubmit}>
        {isPending ? `${submitLabel}…` : submitLabel}
      </Button>
    </form>
  );
};
