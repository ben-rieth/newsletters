import { useState } from 'react';
import { useForm } from '@tanstack/react-form';
import { z } from 'zod';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '#/components/ui/alert-dialog';
import { Button } from '#/components/ui/button';
import { FieldError } from '#/components/ui/field';
import { FormField } from '#/components/ui/form-field';
import { Input } from '#/components/ui/input';
import { getErrorMessage } from '#/lib/errors';
import useDeleteAccount from '../queries/hooks/useDeleteAccount';

const deleteAccountSchema = z.object({
  password: z.string().min(1, 'Password is required'),
});

export const DeleteAccountForm = () => {
  const { mutateAsync, isPending, error } = useDeleteAccount();
  const [open, setOpen] = useState(false);

  const form = useForm({
    defaultValues: { password: '' },
    validators: { onChange: deleteAccountSchema },
    onSubmit: async ({ value }) => {
      await mutateAsync({ password: value.password });
      setOpen(false);
    },
  });

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger render={<Button variant="destructive" />}>
        Delete Account
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete account?</AlertDialogTitle>
          <AlertDialogDescription>
            This permanently deletes your account and all newsletters. Enter
            your password to confirm.
          </AlertDialogDescription>
        </AlertDialogHeader>

        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            form.handleSubmit();
          }}
        >
          <form.Field name="password">
            {(field) => (
              <FormField field={field} label="Password">
                <Input
                  id={field.name}
                  type="password"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  aria-invalid={
                    field.state.meta.isBlurred &&
                    field.state.meta.errors.length > 0
                  }
                />
              </FormField>
            )}
          </form.Field>

          {error && <FieldError>{getErrorMessage(error)}</FieldError>}

          <AlertDialogFooter>
            <AlertDialogCancel type="button" onClick={() => form.reset()}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              type="submit"
              variant="destructive"
              disabled={isPending}
            >
              {isPending ? 'Deleting…' : 'Delete Account'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </form>
      </AlertDialogContent>
    </AlertDialog>
  );
};
