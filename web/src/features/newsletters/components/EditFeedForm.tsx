import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import { ExternalLink } from 'lucide-react';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { Label } from '#/components/ui/label';
const editFeedSchema = z.object({
  alias: z.string(),
});

type FeedLike = {
  alias: string;
  title: string;
  url: string;
  description?: string | null;
};

type Props = {
  feed: FeedLike;
  onSubmit: (alias: string) => Promise<void>;
  isPending?: boolean;
  error?: string;
};

export const EditFeedForm = ({
  feed,
  onSubmit,
  isPending = false,
  error,
}: Props) => {
  const form = useForm({
    defaultValues: { alias: feed.alias ?? '' },
    validators: { onChange: editFeedSchema },
    onSubmit: async ({ value }) => {
      await onSubmit(value.alias ?? '');
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
      <div className="space-y-1.5">
        <Label>Feed URL</Label>
        <div className="flex items-start gap-1.5">
          <p className="text-sm text-muted-foreground break-all">{feed.url}</p>
          <a
            href={feed.url}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="Open feed URL"
            className="mt-0.5 shrink-0 text-muted-foreground hover:text-foreground transition-colors"
          >
            <ExternalLink className="size-3.5" />
          </a>
        </div>
      </div>

      <div className="rounded-md border bg-muted/40 p-3 space-y-1">
        <p className="text-sm font-medium">{feed.title}</p>
        {feed.description && (
          <p className="text-xs text-muted-foreground">{feed.description}</p>
        )}
      </div>

      <form.Field name="alias">
        {(field) => (
          <FormField field={field} label="Custom name (optional)">
            <Input
              id={field.name}
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              placeholder={feed.title}
            />
          </FormField>
        )}
      </form.Field>

      {error && <FieldError>{error}</FieldError>}

      <Button type="submit" disabled={isPending}>
        {isPending ? 'Saving…' : 'Save Changes'}
      </Button>
    </form>
  );
};
