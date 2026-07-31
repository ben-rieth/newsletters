import { useState } from 'react';
import { Link, useParams } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { cn } from '#/lib/utils';
import { Button, buttonVariants } from '#/components/ui/button';
import { EmptyState } from '#/components/EmptyState';
import { ErrorState } from '#/components/ErrorState';
import { formatUpcoming } from '#/utils/format';
import { newslettersOptions } from '#/features/newsletters/queries/newsletters';
import { CreateNewsletterDialog } from '#/features/newsletters/components/CreateNewsletterDialog';
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

const IssuesEmpty = () => {
  const { data, isError, refetch } = useQuery(newslettersOptions);
  const [createOpen, setCreateOpen] = useState(false);
  const newsletters = data ?? [];

  // Without newsletters loaded we can't tell "you have none" from "the request
  // failed" — and guessing wrong tells you to recreate what you already have.
  if (isError) {
    return (
      <ErrorState
        className="px-4 md:px-4"
        title="Couldn’t check your newsletters"
        description="Your issues are fine — this is just the list of newsletters that decides what to show here."
        onRetry={() => void refetch()}
      />
    );
  }

  if (newsletters.length === 0) {
    return (
      <>
        <EmptyState
          title="Nothing to read yet"
          description="A newsletter is a group of feeds delivered on a schedule you choose. Make one, paste in a few feed URLs, and issues land here."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              Create your first newsletter
            </Button>
          }
        />
        <CreateNewsletterDialog
          open={createOpen}
          onOpenChange={setCreateOpen}
        />
      </>
    );
  }

  const next = newsletters
    .filter((n) => n.status === 'active')
    .toSorted(
      (a, b) =>
        new Date(a.nextSendTime).getTime() - new Date(b.nextSendTime).getTime(),
    )
    .at(0);

  if (!next) {
    return (
      <EmptyState
        title="Everything is paused"
        description={`You have ${newsletters.length} ${
          newsletters.length === 1 ? 'newsletter' : 'newsletters'
        }, but none are sending. Activate one to start receiving issues.`}
        action={
          <Link
            to="/newsletters"
            className={buttonVariants({ variant: 'outline' })}
          >
            Manage newsletters
          </Link>
        }
      />
    );
  }

  return (
    <EmptyState
      title="Your first issue is on the way"
      description={
        <>
          <span className="text-foreground">{next.name}</span> sends{' '}
          {formatUpcoming(next.nextSendTime)}. Until then there is nothing to
          read — that&rsquo;s the point.
        </>
      }
      action={
        <Link
          to="/newsletters"
          className={buttonVariants({ variant: 'outline' })}
        >
          Add more feeds
        </Link>
      }
    />
  );
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
      <h2 className="hidden px-4 pb-3 font-serif text-2xl font-medium tracking-tight md:block">
        Issues
      </h2>

      {sorted.length === 0 ? (
        <IssuesEmpty />
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
                        'mt-1.5 space-y-1 font-serif text-[0.9375rem] leading-snug',
                        unread ? 'text-foreground' : 'text-muted-foreground',
                      )}
                    >
                      {titles.map((title, index) => (
                        <li key={`${index}-${title}`} className="line-clamp-1">
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
