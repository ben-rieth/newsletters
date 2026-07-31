import { Link, useParams } from '@tanstack/react-router';
import { cn } from '#/lib/utils';
import type { Issue } from '../queries/issues';

type Props = {
  issues: Issue[];
};

const bucketLabel = (date: Date): string => {
  const now = new Date();
  if (date.toDateString() === now.toDateString()) return 'Today';
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  if (date.toDateString() === yesterday.toDateString()) return 'Yesterday';
  return date.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'short',
    day: 'numeric',
  });
};

const summarize = (
  itemCount: number,
  unreadCount: number,
  remaining: number,
): string => {
  if (itemCount === 0) return 'No items';

  const parts = [];
  if (remaining > 0) parts.push(`+${remaining} more`);
  parts.push(unreadCount > 0 ? `${unreadCount} unread` : 'All read');
  return parts.join(' · ');
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
      <h2 className="hidden px-4 pb-3 text-xl font-semibold tracking-tight md:block">
        Issues
      </h2>

      {sorted.length === 0 ? (
        <p className="px-4 py-8 text-sm text-muted-foreground">
          No issues yet — newsletters appear here after they're first sent.
        </p>
      ) : (
        groups.map((group) => (
          <div key={group.label} className="mb-2">
            <p className="px-4 py-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              {group.label}
            </p>
            {group.items.map((issue) => {
              const date = new Date(issue.sentAt);
              const isActive = issue.issueId === activeId;
              const unread = issue.state === 'unread';
              const titles = issue.previewTitles ?? [];
              const remaining = issue.itemCount - titles.length;
              return (
                <Link
                  key={issue.issueId}
                  to="/issues/$issueId"
                  params={{ issueId: issue.issueId }}
                  className={cn(
                    'block border-l-2 px-4 py-3.5 transition-colors active:bg-accent/60 md:py-3',
                    isActive
                      ? 'border-primary bg-accent'
                      : 'border-transparent hover:bg-accent/50',
                  )}
                >
                  <div className="flex items-baseline justify-between gap-2">
                    <span
                      className={cn(
                        'flex min-w-0 items-baseline gap-1.5 text-xs font-semibold uppercase tracking-wide',
                        unread ? 'text-primary' : 'text-muted-foreground',
                      )}
                    >
                      {unread && (
                        <>
                          <span
                            className="size-1.5 shrink-0 self-center rounded-full bg-primary"
                            aria-hidden="true"
                          />
                          <span className="sr-only">Unread</span>
                        </>
                      )}
                      <span className="truncate">{issue.newsletterName}</span>
                    </span>
                    <span className="shrink-0 text-xs tabular-nums text-muted-foreground">
                      {date.toLocaleTimeString('en-US', {
                        hour: '2-digit',
                        minute: '2-digit',
                        hour12: false,
                      })}
                    </span>
                  </div>

                  {titles.length > 0 && (
                    <ul
                      className={cn(
                        'mt-1.5 space-y-1 text-sm leading-snug',
                        unread ? 'text-foreground' : 'text-muted-foreground',
                      )}
                    >
                      {titles.map((title) => (
                        <li key={title} className="line-clamp-1">
                          {title}
                        </li>
                      ))}
                    </ul>
                  )}

                  <p className="mt-1.5 text-xs text-muted-foreground">
                    {summarize(issue.itemCount, issue.unreadCount, remaining)}
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
