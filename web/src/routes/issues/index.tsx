import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { useMobileHeader } from '#/components/MobileHeader';
import { issuesOptions } from '#/features/issues/queries/issues';

const IssuesIndex = () => {
  useMobileHeader({ title: 'Issues' });

  const { data } = useSuspenseQuery(issuesOptions);
  const issues = data ?? [];

  // With nothing to select, the list column already carries the onboarding —
  // a second prompt here would only restate it.
  if (issues.length === 0) return null;

  const unread = issues.filter((issue) => issue.state === 'unread').length;

  return (
    <div className="flex h-full items-center px-8 lg:px-12">
      <p className="max-w-xs text-sm leading-relaxed text-muted-foreground">
        {unread > 0
          ? `${unread} ${unread === 1 ? 'issue is' : 'issues are'} waiting. Pick one from the list to start reading.`
          : 'You’re all caught up. Pick any issue from the list to read it again.'}
      </p>
    </div>
  );
};

export const Route = createFileRoute('/issues/')({
  component: IssuesIndex,
});
