import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Check, Circle, CircleCheck, Trash2 } from 'lucide-react';
import { formatDisplayUrl, formatRelativeTime } from '#/utils/format';
import { Button } from '#/components/ui/button';
import { MobileHeaderAction } from '#/components/MobileHeader';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '#/components/ui/alert-dialog';
import { cn } from '#/lib/utils';
import type {
  DetailedIssue,
  IssueFeed,
  IssueItem,
} from '#/features/issues/queries/issues';
import useUpdateIssueState from '#/features/issues/queries/hooks/useUpdateIssueState';
import useUpdateIssueItemState from '#/features/issues/queries/hooks/useUpdateIssueItemState';
import useDeleteIssue from '#/features/issues/queries/hooks/useDeleteIssue';
import { IssueEnd } from './IssueEnd';

interface IssueDetailProps {
  issue: DetailedIssue;
}

const unreadFirst = (a: IssueItem, b: IssueItem) => {
  const aRead = a.state === 'read';
  const bRead = b.state === 'read';
  if (aRead !== bRead) return aRead ? 1 : -1;

  const byDate =
    new Date(b.publishDate).getTime() - new Date(a.publishDate).getTime();
  if (byDate !== 0) return byDate;

  return a.title.localeCompare(b.title);
};

const sortUnreadFirst = (feeds: IssueFeed[]) =>
  feeds
    .map((feed) => ({
      feed,
      items: [...(feed.items ?? [])].sort(unreadFirst),
    }))
    .sort((a, b) => {
      const aRead = a.items.every((i) => i.state === 'read');
      const bRead = b.items.every((i) => i.state === 'read');
      if (aRead !== bRead) return aRead ? 1 : -1;
      return a.feed.title.localeCompare(b.feed.title);
    });

const rankBy = <TKey,>(keys: TKey[]) =>
  new Map(keys.map((key, index) => [key, index]));

const rankOf = <TKey,>(ranks: Map<TKey, number>, key: TKey) =>
  ranks.get(key) ?? Number.MAX_SAFE_INTEGER;

const IssueDetail = ({ issue }: IssueDetailProps) => {
  const navigate = useNavigate();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const updateIssueState = useUpdateIssueState(issue.issueId);
  const updateItemState = useUpdateIssueItemState(issue.issueId);
  const deleteIssue = useDeleteIssue(issue.issueId, () => {
    navigate({ to: '/issues' });
  });

  const feeds = issue.feeds ?? [];
  const feedCount = feeds.length;
  const itemCount = feeds.reduce(
    (sum, feed) => sum + (feed.items ?? []).length,
    0,
  );

  const issueRead = issue.state === 'read';
  const allRead =
    itemCount > 0 &&
    feeds.every((feed) => (feed.items ?? []).every((i) => i.state === 'read'));
  const pendingItemId = updateItemState.isPending
    ? updateItemState.variables.itemId
    : undefined;

  const sentDate = new Date(issue.sentAt).toLocaleDateString('en-US', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  });

  // Frozen at load: re-sorting as items are read would move them out from
  // under the cursor, and a misclicked row would vanish before it could be undone.
  const [rank] = useState(() => {
    const sorted = sortUnreadFirst(feeds);
    return {
      feed: rankBy(sorted.map(({ feed }) => feed.feedId)),
      item: rankBy(
        sorted.flatMap(({ items }) => items.map((item) => item.itemId)),
      ),
    };
  });

  const sortedFeeds = feeds
    .map((feed) => ({
      feed,
      items: [...(feed.items ?? [])].sort(
        (a, b) => rankOf(rank.item, a.itemId) - rankOf(rank.item, b.itemId),
      ),
    }))
    .sort(
      (a, b) =>
        rankOf(rank.feed, a.feed.feedId) - rankOf(rank.feed, b.feed.feedId),
    );

  const toggleIssueRead = () =>
    updateIssueState.mutate(issueRead ? 'unread' : 'read');

  return (
    <div className="space-y-8">
      <MobileHeaderAction>
        <>
          {itemCount > 0 && (
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
                <CircleCheck className="size-5" />
              ) : (
                <Check className="size-5" />
              )}
            </Button>
          )}

          <Button
            variant="ghost"
            size="icon"
            className="size-11 text-muted-foreground hover:text-destructive"
            aria-label="Delete issue"
            disabled={deleteIssue.isPending}
            focusableWhenDisabled
            onClick={() => setDeleteDialogOpen(true)}
          >
            <Trash2 className="size-5" />
          </Button>
        </>
      </MobileHeaderAction>

      <header className="flex items-start justify-between gap-4 border-b border-border pb-6">
        <div className="min-w-0">
          <h1 className="hidden font-serif text-3xl font-medium tracking-tight md:block md:text-4xl">
            {issue.newsletterName}
          </h1>
          <p className="text-sm text-muted-foreground md:mt-2">
            {sentDate} · {itemCount} {itemCount === 1 ? 'item' : 'items'} from{' '}
            {feedCount} {feedCount === 1 ? 'feed' : 'feeds'}
          </p>
        </div>

        <div className="hidden shrink-0 items-center gap-2 md:flex">
          {itemCount > 0 && (
            <Button
              variant="outline"
              size="sm"
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

          <Button
            variant="ghost"
            size="sm"
            className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
            disabled={deleteIssue.isPending}
            focusableWhenDisabled
            onClick={() => setDeleteDialogOpen(true)}
          >
            <Trash2 data-icon="inline-start" />
            Delete
          </Button>
        </div>
      </header>

      {feeds.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No content in this issue.
        </p>
      ) : (
        <div className="space-y-8">
          {sortedFeeds.map(({ feed, items }) => (
            <section
              key={feed.feedId}
              className="space-y-4 border-t border-border pt-6 first:border-0 first:pt-0"
            >
              <h2 className="text-sm font-semibold">
                <a
                  href={feed.webUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:underline"
                >
                  {feed.title}
                </a>
              </h2>

              <ul className="space-y-5">
                {items.map((item: IssueItem) => {
                  const read = item.state === 'read';
                  return (
                    <li
                      key={item.itemId}
                      className="group relative flex items-start gap-3 border-b border-border/50 pb-5 last:border-0"
                    >
                      <span
                        aria-hidden="true"
                        className={cn(
                          'absolute top-px -left-3 h-5 w-0.5 rounded-full bg-primary transition-opacity duration-300 ease-out',
                          read && 'opacity-0',
                        )}
                      />
                      <div className="min-w-0 flex-1">
                        <a
                          href={`/api/link/${item.token}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className={cn(
                            'block font-serif text-xl font-medium leading-snug transition-colors duration-300 ease-out hover:underline',
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
                          {!read && <span className="sr-only">Unread: </span>}
                          {item.title}
                        </a>
                        <p className="mt-1 flex min-w-0 gap-1.5 text-sm text-muted-foreground md:text-xs">
                          <span className="shrink-0">
                            {formatRelativeTime(item.publishDate)}
                          </span>
                          <span aria-hidden="true">·</span>
                          <span
                            className="truncate text-muted-foreground/70"
                            title={item.url}
                          >
                            {formatDisplayUrl(item.url)}
                          </span>
                        </p>
                      </div>

                      <Button
                        variant="ghost"
                        size="icon-sm"
                        className="size-11 shrink-0 text-muted-foreground transition-opacity hover:text-foreground md:size-8 md:opacity-0 md:group-focus-within:opacity-100 md:group-hover:opacity-100"
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
                        {read ? <CircleCheck /> : <Circle />}
                      </Button>
                    </li>
                  );
                })}
              </ul>
            </section>
          ))}
          {allRead && <IssueEnd issueId={issue.issueId} />}
        </div>
      )}

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete this issue?</AlertDialogTitle>
            <AlertDialogDescription>
              The {sentDate} issue of &ldquo;{issue.newsletterName}&rdquo; and
              all {itemCount} of its {itemCount === 1 ? 'item' : 'items'} will
              be permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={deleteIssue.isPending}
              onClick={() => deleteIssue.mutate()}
            >
              {deleteIssue.isPending ? 'Deleting…' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};

export default IssueDetail;
