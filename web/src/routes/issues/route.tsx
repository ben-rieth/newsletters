import { createFileRoute, Outlet } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import IssueListColumn from '#/features/issues/components/IssueListColumn';
import { issuesOptions } from '#/features/issues/queries/issues';

const IssuesLayout = () => {
  const { data: issues } = useSuspenseQuery(issuesOptions);

  return (
    <div className="flex h-full">
      <aside className="w-72 shrink-0 overflow-y-auto border-r border-border lg:w-80">
        <IssueListColumn issues={issues ?? []} />
      </aside>
      <div className="flex-1 overflow-y-auto">
        <Outlet />
      </div>
    </div>
  );
};

export const Route = createFileRoute('/issues')({
  component: IssuesLayout,
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(issuesOptions);
  },
});
