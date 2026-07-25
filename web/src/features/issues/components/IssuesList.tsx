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
        return (
          <Link
            key={issue.issueId}
            to="/issues/$issueId"
            params={{ issueId: issue.issueId }}
            className={cn('group', listRowClass)}
          >
            <div className="min-w-0">
              <p className="truncate font-medium group-hover:underline">
                {date.toLocaleDateString('en-US', {
                  weekday: 'long',
                  month: 'long',
                  day: 'numeric',
                  year: 'numeric',
                })}
              </p>
              <p className="mt-0.5 truncate text-xs text-muted-foreground">
                {issue.newsletterName} ·{' '}
                {date.toLocaleTimeString('en-US', {
                  hour: '2-digit',
                  minute: '2-digit',
                  hour12: false,
                })}
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
