import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { SettingsRow, SettingsSection } from '#/components/SettingsRow';
import FeedLinkDisplay from './FeedLinkDisplay';

const editFeedSchema = z.object({
  alias: z.string(),
});

type FeedLike = {
  alias: string;
  title: string;
  url: string;
  htmlUrl: string;
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
    defaultValues: { alias: feed.alias },
    validators: { onChange: editFeedSchema },
    onSubmit: async ({ value }) => {
      await onSubmit(value.alias);
    },
  });

  return (
    <form
      className="space-y-6"
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <SettingsSection>
        <SettingsRow
          title="Feed source"
          description="Where this feed's articles are pulled from."
        >
          <div className="space-y-4">
            <FeedLinkDisplay label="Feed URL" url={feed.url} />
            <FeedLinkDisplay label="Web URL" url={feed.htmlUrl} />
          </div>
        </SettingsRow>

        <SettingsRow
          title="Display name"
          description="Shown in place of the feed's title in your newsletter."
        >
          <div className="space-y-4">
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
              {isPending ? 'Saving…' : 'Save changes'}
            </Button>
          </div>
        </SettingsRow>
      </SettingsSection>
    </form>
  );
};
