import type { ReactNode } from 'react';
import { cn } from '#/lib/utils';

export const listRowClass =
  'flex items-center justify-between gap-3 px-1 py-4 transition-colors hover:bg-accent/30 md:gap-4 md:py-5';

type Props = {
  children: ReactNode;
  className?: string;
};

export const ListPanel = ({ children, className }: Props) => (
  <div
    className={cn('divide-y divide-border border-y border-border', className)}
  >
    {children}
  </div>
);
