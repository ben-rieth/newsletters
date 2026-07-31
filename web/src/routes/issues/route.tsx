import { createFileRoute, Outlet, useParams } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { cn } from '#/lib/utils';
import IssueListColumn from '#/features/issues/components/IssueListColumn';
import { issuesOptions } from '#/features/issues/queries/issues';
import { newslettersOptions } from '#/features/newsletters/queries/newsletters';

const IssuesLayout = () => {
  const { data: issues } = useSuspenseQuery(issuesOptions);
  const params: { issueId?: string } = useParams({ strict: false });
  const hasSelection = Boolean(params.issueId);

  return (
    <div className="flex h-full">
      <aside
        className={cn(
          'w-full shrink-0 overflow-y-auto border-r border-border md:w-72 lg:w-80',
          hasSelection && 'hidden md:block',
        )}
      >
        <IssueListColumn issues={issues ?? []} />
      </aside>
      <div
        className={cn(
          'flex-1 overflow-y-auto',
          !hasSelection && 'hidden md:block',
        )}
      >
        <Outlet />
      </div>
    </div>
  );
};

export const Route = createFileRoute('/issues')({
  component: IssuesLayout,
  loader: async ({ context }) => {
    await Promise.all([
      context.queryClient.ensureQueryData(issuesOptions),
      context.queryClient.ensureQueryData(newslettersOptions),
    ]);
  },
});
