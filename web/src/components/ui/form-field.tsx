import type { ReactNode } from 'react';
import type { AnyFieldApi } from '@tanstack/react-form';
import { Field, FieldError, FieldLabel } from '#/components/ui/field';

interface FormFieldProps {
  field: AnyFieldApi;
  label: string;
  children: ReactNode;
}

const FormField = ({ field, label, children }: FormFieldProps) => {
  const showError =
    field.state.meta.isBlurred || field.form.state.submissionAttempts > 0;
  const hasError = showError && field.state.meta.errors.length > 0;

  return (
    <Field data-invalid={hasError}>
      <FieldLabel htmlFor={field.name}>{label}</FieldLabel>
      {children}
      {showError && (
        <FieldError>{field.state.meta.errors[0]?.message}</FieldError>
      )}
    </Field>
  );
};

export { FormField };
