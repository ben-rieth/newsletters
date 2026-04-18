import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { getErrorMessage } from '#/lib/errors';
import useUpdateEmail from '../queries/hooks/useUpdateEmail';

const updateEmailSchema = z.object({
  email: z.string().email('Invalid email address'),
  confirmEmail: z.string().email('Invalid email address'),
}).refine((data) => data.email === data.confirmEmail, {
  message: 'Emails do not match',
  path: ['confirmEmail'],
});

export const UpdateEmailForm = () => {
  const { mutateAsync, isPending, error } = useUpdateEmail();

  const form = useForm({
    defaultValues: { email: '', confirmEmail: '' },
    validators: { onChange: updateEmailSchema },
    onSubmit: async ({ value }) => {
      await mutateAsync(value.email);
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
      <form.Field name="email">
        {(field) => (
          <FormField field={field} label="New Email">
            <Input
              id={field.name}
              type="email"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              placeholder="new@example.com"
              aria-invalid={field.state.meta.isBlurred && field.state.meta.errors.length > 0}
            />
          </FormField>
        )}
      </form.Field>

      <form.Field name="confirmEmail">
        {(field) => (
          <FormField field={field} label="Confirm New Email">
            <Input
              id={field.name}
              type="email"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              placeholder="new@example.com"
              aria-invalid={field.state.meta.isBlurred && field.state.meta.errors.length > 0}
            />
          </FormField>
        )}
      </form.Field>

      {error && <FieldError>{getErrorMessage(error)}</FieldError>}

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Updating…' : 'Update Email'}
      </Button>
    </form>
  );
};