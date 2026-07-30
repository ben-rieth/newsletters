import { createFileRoute } from '@tanstack/react-router';
import { Mail } from 'lucide-react';
import { useMobileHeader } from '#/components/MobileHeader';

const IssuesEmptyState = () => {
  useMobileHeader({ title: 'Issues' });

  return (
    <div className="flex h-full flex-col items-center justify-center gap-3 p-6 text-center">
      <span className="flex size-12 items-center justify-center rounded-full bg-muted text-muted-foreground">
        <Mail className="size-6" />
      </span>
      <p className="text-sm text-muted-foreground">
        Select an issue to start reading.
      </p>
    </div>
  );
};

export const Route = createFileRoute('/issues/')({
  component: IssuesEmptyState,
});
