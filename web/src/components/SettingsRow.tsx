import type { ReactNode } from 'react';
import { cn } from '#/lib/utils';

type SettingsRowProps = {
  title: string;
  description?: ReactNode;
  children: ReactNode;
  className?: string;
};

export const SettingsRow = ({
  title,
  description,
  children,
  className,
}: SettingsRowProps) => (
  <div
    className={cn(
      'grid gap-x-8 gap-y-4 py-6 md:grid-cols-[minmax(0,18rem)_minmax(0,1fr)]',
      className,
    )}
  >
    <div>
      <h3 className="text-sm font-medium">{title}</h3>
      {description != null && (
        <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      )}
    </div>
    <div className="max-w-md">{children}</div>
  </div>
);

export const SettingsSection = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => (
  <div
    className={cn('divide-y divide-border border-y border-border', className)}
  >
    {children}
  </div>
);
