'use client';

import { Switch as SwitchPrimitive } from '@base-ui/react/switch';

import { cn } from '#/lib/utils.ts';

function Switch({ className, ...props }: SwitchPrimitive.Root.Props) {
  return (
    <SwitchPrimitive.Root
      data-slot="switch"
      className={cn(
        'inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border border-transparent bg-input p-px transition-colors outline-none select-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 disabled:pointer-events-none disabled:opacity-50 data-checked:bg-primary',
        className,
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        data-slot="switch-thumb"
        className="block size-4 rounded-full bg-background shadow-sm transition-transform data-checked:translate-x-4"
      />
    </SwitchPrimitive.Root>
  );
}

export { Switch };
