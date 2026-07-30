import type { ReactNode } from 'react';
import { cn } from '#/lib/utils';

export const listRowClass =
  'flex items-center justify-between gap-3 px-4 py-3 transition-colors hover:bg-accent/40 md:gap-4 md:px-5 md:py-4';

type Props = {
  header?: ReactNode;
  children: ReactNode;
  className?: string;
};

export const ListPanel = ({ header, children, className }: Props) => (
  <div className={cn('overflow-hidden rounded-lg border', className)}>
    {header != null && (
      <div className="border-b bg-muted/20 px-4 py-2.5 md:px-5 text-xs font-medium uppercase tracking-wider text-muted-foreground">
        {header}
      </div>
    )}
    <div className="divide-y divide-border">{children}</div>
  </div>
);
