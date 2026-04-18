import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import type { components } from '#/api/schema';

type FeedFilter = components['schemas']['FeedFilter'];

const FIELD_OPTIONS = [
  { value: 'title', label: 'Title' },
  { value: 'url', label: 'URL' },
];

const OPERATOR_OPTIONS = [
  { value: 'contains', label: 'Contains' },
  { value: 'does_not_contain', label: 'Does Not Contain' },
];

const filterSchema = z.object({
  field: z.enum(['title', 'url']),
  operator: z.enum(['contains', 'does_not_contain']),
  pattern: z.string().min(1, 'Pattern is required'),
});

type Props = {
  filter?: FeedFilter;
  onSubmit: (
    values: components['schemas']['SubmittableFeedFilterFields'],
  ) => Promise<void>;
  onCancel?: () => void;
  isPending?: boolean;
  error?: string;
};

export const FeedFilterForm = ({
  filter,
  onSubmit,
  onCancel,
  isPending = false,
  error,
}: Props) => {
  const isEditing = filter !== undefined;

  const form = useForm({
    defaultValues: {
      field: (filter?.field ?? 'title') as 'title' | 'url',
      operator: (filter?.operator ?? 'contains') as
        | 'contains'
        | 'does_not_contain',
      pattern: filter?.pattern ?? '',
    },
    validators: { onChange: filterSchema },
    onSubmit: async ({ value }) => {
      await onSubmit(value);
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
      <div className="grid grid-cols-2 gap-3">
        <form.Field name="field">
          {(field) => (
            <FormField field={field} label="Field">
              <Select
                value={field.state.value}
                onValueChange={(value) =>
                  field.handleChange(value as 'title' | 'url')
                }
                items={FIELD_OPTIONS}
              >
                <SelectTrigger id={field.name} className="w-full">
                  <SelectValue placeholder="Select field" />
                </SelectTrigger>
                <SelectContent>
                  {FIELD_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          )}
        </form.Field>

        <form.Field name="operator">
          {(field) => (
            <FormField field={field} label="Operator">
              <Select
                value={field.state.value}
                onValueChange={(value) =>
                  field.handleChange(value as 'contains' | 'does_not_contain')
                }
                items={OPERATOR_OPTIONS}
              >
                <SelectTrigger id={field.name} className="w-full">
                  <SelectValue placeholder="Select operator" />
                </SelectTrigger>
                <SelectContent>
                  {OPERATOR_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
          )}
        </form.Field>
      </div>

      <form.Field name="pattern">
        {(field) => (
          <FormField field={field} label="Pattern">
            <Input
              id={field.name}
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              placeholder="e.g. example.com"
              aria-invalid={
                field.state.meta.isBlurred && field.state.meta.errors.length > 0
              }
            />
          </FormField>
        )}
      </form.Field>

      {error && <FieldError>{error}</FieldError>}

      <div className="flex gap-2">
        <Button type="submit" disabled={isPending} size="sm">
          {isPending
            ? isEditing
              ? 'Saving…'
              : 'Adding…'
            : isEditing
              ? 'Save'
              : 'Add Filter'}
        </Button>
        {onCancel && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onCancel}
            disabled={isPending}
          >
            Cancel
          </Button>
        )}
      </div>
    </form>
  );
};
