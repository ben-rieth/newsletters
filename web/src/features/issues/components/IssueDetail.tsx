import { Check, Circle, CircleCheck } from 'lucide-react';
import { formatRelativeTime } from '#/utils/format';
import { Button } from '#/components/ui/button';
import { MobileHeaderAction } from '#/components/MobileHeader';
import { cn } from '#/lib/utils';
import type {
  DetailedIssue,
  IssueFeed,
  IssueItem,
} from '#/features/issues/queries/issues';
import useUpdateIssueState from '#/features/issues/queries/hooks/useUpdateIssueState';
import useUpdateIssueItemState from '#/features/issues/queries/hooks/useUpdateIssueItemState';

interface IssueDetailProps {
  issue: DetailedIssue;
}

const IssueDetail = ({ issue }: IssueDetailProps) => {
  const updateIssueState = useUpdateIssueState(issue.issueId);
  const updateItemState = useUpdateIssueItemState(issue.issueId);

  const feeds = issue.feeds ?? [];
  const feedCount = feeds.length;
  const itemCount = feeds.reduce(
    (sum, feed) => sum + (feed.items ?? []).length,
    0,
  );

  const issueRead = issue.state === 'read';
  const pendingItemId = updateItemState.isPending
    ? updateItemState.variables.itemId
    : undefined;

  const sentDate = new Date(issue.sentAt).toLocaleDateString('en-US', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  });

  const unreadFirst = (a: IssueItem, b: IssueItem) => {
    const aRead = a.state === 'read';
    const bRead = b.state === 'read';
    if (aRead !== bRead) return aRead ? 1 : -1;

    const byDate =
      new Date(b.publishDate).getTime() - new Date(a.publishDate).getTime();
    if (byDate !== 0) return byDate;

    return a.title.localeCompare(b.title);
  };

  const sortedFeeds = feeds
    .map((feed: IssueFeed, feedIdx: number) => ({
      feed,
      feedIdx,
      items: [...(feed.items ?? [])].sort(unreadFirst),
    }))
    .sort((a, b) => {
      const aRead = a.items.every((i) => i.state === 'read');
      const bRead = b.items.every((i) => i.state === 'read');
      if (aRead !== bRead) return aRead ? 1 : -1;
      return a.feed.title.localeCompare(b.feed.title);
    });

  const toggleIssueRead = () =>
    updateIssueState.mutate(issueRead ? 'unread' : 'read');

  return (
    <div className="space-y-8">
      {itemCount > 0 && (
        <MobileHeaderAction>
          <Button
            variant="ghost"
            size="icon"
            className="size-11"
            aria-label={issueRead ? 'Mark all unread' : 'Mark all read'}
            aria-pressed={issueRead}
            disabled={updateIssueState.isPending}
            focusableWhenDisabled
            onClick={toggleIssueRead}
          >
            {issueRead ? (
              <CircleCheck className="size-5 text-primary" />
            ) : (
              <Check className="size-5" />
            )}
          </Button>
        </MobileHeaderAction>
      )}

      <header className="flex items-start justify-between gap-4 border-b border-border pb-6">
        <div className="min-w-0">
          <h1 className="hidden font-serif text-3xl font-medium tracking-tight md:block md:text-4xl">
            {issue.newsletterName}
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            {sentDate} · {itemCount} {itemCount === 1 ? 'item' : 'items'} from{' '}
            {feedCount} {feedCount === 1 ? 'feed' : 'feeds'}
          </p>
        </div>

        {itemCount > 0 && (
          <Button
            variant="outline"
            size="sm"
            className="hidden shrink-0 md:inline-flex"
            disabled={updateIssueState.isPending}
            focusableWhenDisabled
            onClick={toggleIssueRead}
          >
            {issueRead ? (
              <CircleCheck data-icon="inline-start" />
            ) : (
              <Check data-icon="inline-start" />
            )}
            {issueRead ? 'Mark all unread' : 'Mark all read'}
          </Button>
        )}
      </header>

      {feeds.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No content in this issue.
        </p>
      ) : (
        <div className="space-y-8">
          {sortedFeeds.map(({ feed, feedIdx, items }) => (
            <section key={feedIdx} className="space-y-4">
              <a
                href={feed.webUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs font-semibold uppercase tracking-wide text-primary hover:underline"
              >
                {feed.title}
              </a>

              <ul className="space-y-5">
                {items.map((item: IssueItem) => {
                  const read = item.state === 'read';
                  return (
                    <li
                      key={item.itemId}
                      className="flex items-start gap-3 border-b border-border/50 pb-5 last:border-0"
                    >
                      <div className="min-w-0 flex-1">
                        <a
                          href={`/api/link/${item.token}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className={cn(
                            'block font-serif text-xl font-medium leading-snug transition-colors hover:text-primary',
                            read ? 'text-muted-foreground' : 'text-foreground',
                          )}
                          onClick={() => {
                            if (!read) {
                              updateItemState.mutate({
                                itemId: item.itemId,
                                state: 'read',
                              });
                            }
                          }}
                        >
                          {item.title}
                        </a>
                        <p className="mt-1 text-sm text-muted-foreground md:text-xs">
                          {formatRelativeTime(item.publishDate)}
                        </p>
                      </div>

                      <Button
                        variant="ghost"
                        size="icon-sm"
                        className="size-11 shrink-0 text-muted-foreground hover:text-foreground md:size-8"
                        aria-label="Read"
                        aria-pressed={read}
                        title={read ? 'Mark as unread' : 'Mark as read'}
                        disabled={pendingItemId === item.itemId}
                        focusableWhenDisabled
                        onClick={() =>
                          updateItemState.mutate({
                            itemId: item.itemId,
                            state: read ? 'unread' : 'read',
                          })
                        }
                      >
                        {read ? (
                          <CircleCheck className="text-primary" />
                        ) : (
                          <Circle />
                        )}
                      </Button>
                    </li>
                  );
                })}
              </ul>
            </section>
          ))}
        </div>
      )}
    </div>
  );
};

export default IssueDetail;
