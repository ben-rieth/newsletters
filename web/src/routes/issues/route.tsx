import {
  createFileRoute,
  Link,
  Outlet,
  useParams,
} from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { ChevronLeft } from 'lucide-react';
import { cn } from '#/lib/utils';
import IssueListColumn from '#/features/issues/components/IssueListColumn';
import { issuesOptions } from '#/features/issues/queries/issues';

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
        {hasSelection && (
          <Link
            to="/issues"
            className="flex items-center gap-1 px-6 pt-6 text-sm text-muted-foreground hover:text-foreground md:hidden"
          >
            <ChevronLeft className="size-4" />
            All issues
          </Link>
        )}
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
