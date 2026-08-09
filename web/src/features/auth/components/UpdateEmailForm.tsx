import { useState } from 'react';
import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { getErrorMessage } from '#/lib/errors';
import useUpdateEmail from '../queries/hooks/useUpdateEmail';
import useVerifyEmailUpdate from '../queries/hooks/useVerifyEmailUpdate';
import useResendEmailUpdateVerification from '../queries/hooks/useResendEmailUpdateVerification';
import { VerificationCodeInput } from './VerificationCodeInput';

const updateEmailSchema = z
  .object({
    email: z.string().email('Invalid email address'),
    confirmEmail: z.string().email('Invalid email address'),
    password: z.string().min(1, 'Current password is required'),
  })
  .refine((data) => data.email === data.confirmEmail, {
    message: 'Emails do not match',
    path: ['confirmEmail'],
  });

export const UpdateEmailForm = () => {
  const [isVerifying, setIsVerifying] = useState(false);

  const {
    mutateAsync: updateEmail,
    isPending: isUpdating,
    error: updateError,
  } = useUpdateEmail();
  const {
    mutate: verifyCode,
    isPending: isVerifyingCode,
    error: verifyError,
  } = useVerifyEmailUpdate();
  const {
    mutate: resend,
    isPending: isResending,
    error: resendError,
  } = useResendEmailUpdateVerification();

  const form = useForm({
    defaultValues: { email: '', confirmEmail: '', password: '' },
    validators: { onChange: updateEmailSchema },
    onSubmit: async ({ value }) => {
      await updateEmail(
        { email: value.email, password: value.password },
        { onSuccess: () => setIsVerifying(true) },
      );
    },
  });

  if (isVerifying) {
    return (
      <div className="space-y-4">
        <p className="text-sm text-muted-foreground">
          We sent an 8-digit code to your new email address.
        </p>
        <VerificationCodeInput
          onVerify={(code) =>
            verifyCode(code, {
              onSuccess: () => {
                setIsVerifying(false);
                form.reset();
              },
            })
          }
          isPending={isVerifyingCode}
          error={verifyError ?? resendError}
          onResend={(onSuccess) => resend(undefined, { onSuccess })}
          isResending={isResending}
        />
        <Button
          variant="ghost"
          className="w-full"
          onClick={() => setIsVerifying(false)}
        >
          Cancel
        </Button>
      </div>
    );
  }

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
              aria-invalid={
                field.state.meta.isBlurred && field.state.meta.errors.length > 0
              }
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
              aria-invalid={
                field.state.meta.isBlurred && field.state.meta.errors.length > 0
              }
            />
          </FormField>
        )}
      </form.Field>

      <form.Field name="password">
        {(field) => (
          <FormField field={field} label="Current Password">
            <Input
              id={field.name}
              type="password"
              autoComplete="current-password"
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

      {updateError && <FieldError>{getErrorMessage(updateError)}</FieldError>}

      <Button type="submit" className="w-full" disabled={isUpdating}>
        {isUpdating ? 'Updating…' : 'Update Email'}
      </Button>
    </form>
  );
};
