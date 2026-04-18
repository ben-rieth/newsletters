import { useForm } from '@tanstack/react-form';
import { Link } from '@tanstack/react-router';
import { z } from 'zod';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { getErrorMessage } from '#/lib/errors';
import useAuth from '../queries/hooks/useAuth';

const signUpSchema = z
  .object({
    email: z.string().email('Invalid email address'),
    password: z
      .string()
      .min(8, 'Password must be at least 8 characters')
      .max(80, 'Password must be at most 80 characters'),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

export const SignUpForm = () => {
  const { mutateAsync, isPending, error } = useAuth('/auth/sign-up');

  const form = useForm({
    defaultValues: { email: '', password: '', confirmPassword: '' },
    validators: { onChange: signUpSchema },
    onSubmit: async ({ value }) => {
      await mutateAsync({ email: value.email, password: value.password });
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
          <FormField field={field} label="Email">
            <Input
              id={field.name}
              type="email"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              placeholder="you@example.com"
              aria-invalid={field.state.meta.isBlurred && field.state.meta.errors.length > 0}
            />
          </FormField>
        )}
      </form.Field>

      <form.Field name="password">
        {(field) => (
          <FormField field={field} label="Password">
            <Input
              id={field.name}
              type="password"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              aria-invalid={field.state.meta.isBlurred && field.state.meta.errors.length > 0}
            />
          </FormField>
        )}
      </form.Field>

      <form.Field name="confirmPassword">
        {(field) => (
          <FormField field={field} label="Confirm password">
            <Input
              id={field.name}
              type="password"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              aria-invalid={field.state.meta.isBlurred && field.state.meta.errors.length > 0}
            />
          </FormField>
        )}
      </form.Field>

      {error && <FieldError>{getErrorMessage(error)}</FieldError>}

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Creating account…' : 'Create account'}
      </Button>

      <p className="text-center text-sm text-muted-foreground">
        Already have an account?{' '}
        <Link to="/sign-in" className="underline underline-offset-4">
          Sign in
        </Link>
      </p>
    </form>
  );
};