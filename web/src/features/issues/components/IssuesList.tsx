import { Link } from '@tanstack/react-router';
import { ChevronRight } from 'lucide-react';
import { ListPanel, listRowClass } from '#/components/ListPanel';
import { cn } from '#/lib/utils';
import type { Issue } from '#/features/issues/queries/issues';

interface IssuesListProps {
  issues: Issue[];
}

const IssuesList = ({ issues }: IssuesListProps) => {
  if (issues.length === 0) {
    return (
      <div className="rounded-lg border border-dashed py-12 text-center">
        <p className="text-sm text-muted-foreground">
          No issues yet — sent issues will appear here.
        </p>
      </div>
    );
  }

  const sorted = [...issues].sort(
    (a, b) => new Date(b.sentAt).getTime() - new Date(a.sentAt).getTime(),
  );

  return (
    <ListPanel header="Issue">
      {sorted.map((issue) => {
        const date = new Date(issue.sentAt);
        const unread = issue.state === 'unread';
        return (
          <Link
            key={issue.issueId}
            to="/issues/$issueId"
            params={{ issueId: issue.issueId }}
            className={cn('group', listRowClass)}
          >
            <div className="min-w-0">
              <p
                className={cn(
                  'flex min-w-0 items-center gap-2 group-hover:underline',
                  unread ? 'font-semibold' : 'font-medium',
                )}
              >
                {unread && (
                  <>
                    <span
                      className="size-1.5 shrink-0 rounded-full bg-primary"
                      aria-hidden="true"
                    />
                    <span className="sr-only">Unread</span>
                  </>
                )}
                <span className="truncate">
                  {date.toLocaleDateString('en-US', {
                    weekday: 'long',
                    month: 'long',
                    day: 'numeric',
                    year: 'numeric',
                  })}
                </span>
              </p>
              <p className="mt-0.5 truncate text-xs text-muted-foreground">
                {[
                  date.toLocaleTimeString('en-US', {
                    hour: '2-digit',
                    minute: '2-digit',
                    hour12: false,
                  }),
                  `${issue.itemCount} ${issue.itemCount === 1 ? 'item' : 'items'}`,
                  ...(issue.unreadCount > 0
                    ? [`${issue.unreadCount} unread`]
                    : []),
                ].join(' · ')}
              </p>
            </div>
            <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
          </Link>
        );
      })}
    </ListPanel>
  );
};

export default IssuesList;
