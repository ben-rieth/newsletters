import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { getErrorMessage } from '#/lib/errors';
import useUpdatePassword from '../queries/hooks/useUpdatePassword';

const updatePasswordSchema = z
  .object({
    currentPassword: z.string().min(1, 'Current password is required'),
    newPassword: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string().min(1, 'Please confirm your new password'),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

export const UpdatePasswordForm = () => {
  const { mutateAsync, isPending, error } = useUpdatePassword();

  const form = useForm({
    defaultValues: {
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    },
    validators: { onChange: updatePasswordSchema },
    onSubmit: async ({ value }) => {
      await mutateAsync({
        currentPassword: value.currentPassword,
        newPassword: value.newPassword,
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
      <form.Field name="currentPassword">
        {(field) => (
          <FormField field={field} label="Current Password">
            <Input
              id={field.name}
              type="password"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              aria-invalid={
                field.state.meta.isBlurred && field.state.meta.errors.length > 0
              }
            />
          </FormField>
        )}
      </form.Field>

      <form.Field name="newPassword">
        {(field) => (
          <FormField field={field} label="New Password">
            <Input
              id={field.name}
              type="password"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              aria-invalid={
                field.state.meta.isBlurred && field.state.meta.errors.length > 0
              }
            />
          </FormField>
        )}
      </form.Field>

      <form.Field name="confirmPassword">
        {(field) => (
          <FormField field={field} label="Confirm New Password">
            <Input
              id={field.name}
              type="password"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              aria-invalid={
                field.state.meta.isBlurred && field.state.meta.errors.length > 0
              }
            />
          </FormField>
        )}
      </form.Field>

      {error && <FieldError>{getErrorMessage(error)}</FieldError>}

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Updating…' : 'Update Password'}
      </Button>
    </form>
  );
};
