import { Link, useParams } from '@tanstack/react-router';
import { cn } from '#/lib/utils';
import type { Issue } from '../queries/issues';

type Props = {
  issues: Issue[];
};

const bucketLabel = (date: Date): string => {
  const now = new Date();
  if (date.toDateString() === now.toDateString()) return 'Today';
  const yesterday = new Date(now.getTime() - 86400000);
  if (date.toDateString() === yesterday.toDateString()) return 'Yesterday';
  return date.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'short',
    day: 'numeric',
  });
};

const IssueListColumn = ({ issues }: Props) => {
  const params = useParams({ strict: false });
  const activeId = (params as { issueId?: string }).issueId;

  const sorted = [...issues].sort(
    (a, b) => new Date(b.sentAt).getTime() - new Date(a.sentAt).getTime(),
  );

  const groups: { label: string; items: Issue[] }[] = [];
  for (const issue of sorted) {
    const label = bucketLabel(new Date(issue.sentAt));
    const last = groups.at(-1);
    if (last && last.label === label) {
      last.items.push(issue);
    } else {
      groups.push({ label, items: [issue] });
    }
  }

  return (
    <div className="py-4">
      <h1 className="px-4 pb-3 text-xl font-semibold tracking-tight">Issues</h1>

      {sorted.length === 0 ? (
        <p className="px-4 py-8 text-sm text-muted-foreground">
          No issues yet — newsletters appear here after they're first sent.
        </p>
      ) : (
        groups.map((group) => (
          <div key={group.label} className="mb-2">
            <p className="px-4 py-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground/60">
              {group.label}
            </p>
            {group.items.map((issue) => {
              const date = new Date(issue.sentAt);
              const isActive = issue.issueId === activeId;
              return (
                <Link
                  key={issue.issueId}
                  to="/issues/$issueId"
                  params={{ issueId: issue.issueId }}
                  className={cn(
                    'block border-l-2 px-4 py-3 transition-colors',
                    isActive
                      ? 'border-primary bg-accent'
                      : 'border-transparent hover:bg-accent/50',
                  )}
                >
                  <div className="flex items-baseline justify-between gap-2">
                    <span className="truncate text-xs font-semibold uppercase tracking-wide text-primary">
                      {issue.newsletterName}
                    </span>
                    <span className="shrink-0 text-xs tabular-nums text-muted-foreground">
                      {date.toLocaleTimeString('en-US', {
                        hour: '2-digit',
                        minute: '2-digit',
                        hour12: false,
                      })}
                    </span>
                  </div>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {date.toLocaleDateString('en-US', {
                      weekday: 'short',
                      day: 'numeric',
                      month: 'short',
                    })}
                  </p>
                </Link>
              );
            })}
          </div>
        ))
      )}
    </div>
  );
};

export default IssueListColumn;
