import * as React from 'react';
import { NumberField as NumberFieldPrimitive } from '@base-ui/react/number-field';
import { Minus, Plus } from 'lucide-react';

import { cn } from '#/lib/utils';

function NumberFieldRoot({
  className,
  ...props
}: React.ComponentProps<typeof NumberFieldPrimitive.Root>) {
  return (
    <NumberFieldPrimitive.Root
      data-slot="number-field"
      className={cn('w-full', className)}
      {...props}
    />
  );
}

function NumberFieldGroup({
  className,
  ...props
}: React.ComponentProps<typeof NumberFieldPrimitive.Group>) {
  return (
    <NumberFieldPrimitive.Group
      data-slot="number-field-group"
      className={cn(
        'flex h-9 w-full min-w-0 items-center rounded-md border border-input bg-input/20 transition-colors',
        'focus-within:border-ring focus-within:ring-2 focus-within:ring-ring/30',
        'data-[disabled]:pointer-events-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50',
        'has-[[aria-invalid=true]]:border-destructive has-[[aria-invalid=true]]:ring-2 has-[[aria-invalid=true]]:ring-destructive/20',
        'dark:bg-input/30 dark:has-[[aria-invalid=true]]:border-destructive/50 dark:has-[[aria-invalid=true]]:ring-destructive/40',
        className,
      )}
      {...props}
    />
  );
}

function NumberFieldInput({
  className,
  ...props
}: React.ComponentProps<typeof NumberFieldPrimitive.Input>) {
  return (
    <NumberFieldPrimitive.Input
      data-slot="number-field-input"
      className={cn(
        'h-full w-full min-w-0 bg-transparent px-2 text-center text-sm outline-none placeholder:text-muted-foreground md:text-xs/relaxed',
        className,
      )}
      {...props}
    />
  );
}

function NumberFieldDecrement({
  className,
  children,
  ...props
}: React.ComponentProps<typeof NumberFieldPrimitive.Decrement>) {
  return (
    <NumberFieldPrimitive.Decrement
      data-slot="number-field-decrement"
      className={cn(
        'flex h-full shrink-0 items-center pl-2 pr-1.5 text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-40',
        className,
      )}
      {...props}
    >
      {children ?? <Minus className="size-3" />}
    </NumberFieldPrimitive.Decrement>
  );
}

function NumberFieldIncrement({
  className,
  children,
  ...props
}: React.ComponentProps<typeof NumberFieldPrimitive.Increment>) {
  return (
    <NumberFieldPrimitive.Increment
      data-slot="number-field-increment"
      className={cn(
        'flex h-full shrink-0 items-center pl-1.5 pr-2 text-muted-foreground transition-colors hover:text-foreground disabled:pointer-events-none disabled:opacity-40',
        className,
      )}
      {...props}
    >
      {children ?? <Plus className="size-3" />}
    </NumberFieldPrimitive.Increment>
  );
}

type NumberFieldProps = React.ComponentProps<
  typeof NumberFieldPrimitive.Root
> & {
  id?: string;
  'aria-invalid'?: boolean | 'true' | 'false';
  inputClassName?: string;
};

function NumberField({
  id,
  className,
  inputClassName,
  'aria-invalid': ariaInvalid,
  ...props
}: NumberFieldProps) {
  return (
    <NumberFieldRoot className={className} {...props}>
      <NumberFieldGroup>
        <NumberFieldDecrement />
        <NumberFieldInput
          id={id}
          aria-invalid={ariaInvalid}
          className={inputClassName}
        />
        <NumberFieldIncrement />
      </NumberFieldGroup>
    </NumberFieldRoot>
  );
}

export {
  NumberField,
  NumberFieldRoot,
  NumberFieldGroup,
  NumberFieldInput,
  NumberFieldDecrement,
  NumberFieldIncrement,
};
