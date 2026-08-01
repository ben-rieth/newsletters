import type { ReactNode } from 'react';
import { cn } from '#/lib/utils';

type Props = {
  title: string;
  description?: ReactNode;
  action?: ReactNode;
  className?: string;
};

export const EmptyState = ({
  title,
  description,
  action,
  className,
}: Props) => (
  <div className={cn('max-w-sm px-4 py-10 md:px-5', className)}>
    <h3 className="font-serif text-xl font-medium">{title}</h3>
    {description != null && (
      <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
        {description}
      </p>
    )}
    {action != null && <div className="mt-4">{action}</div>}
  </div>
);
