import { useState } from 'react';
import { Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  ExternalLink,
  MoreVertical,
  Pause,
  Play,
  Search,
  Settings,
  Trash2,
} from 'lucide-react';
import { cn } from '#/lib/utils';
import { ListPanel, listRowClass } from '#/components/ListPanel';
import { Button, buttonVariants } from '#/components/ui/button';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '#/components/ui/input-group';
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
import { Sheet, SheetContent, SheetTitle } from '#/components/ui/sheet';
import { feedsOptions } from '../queries/feeds';
import type { Feed } from '../queries/feeds';
import { AddFeedDialog } from './AddFeedDialog';
import { FeedHealthBadge } from './FeedHealth';
import { FeedStatusBadge } from './FeedStatusBadge';
import useDeleteFeed from '../queries/hooks/useDeleteFeed';
import useUpdateFeedStatus from '../queries/hooks/useUpdateFeedStatus';

type Props = {
  newsletterId: string;
};

const sheetActionClass =
  'flex min-h-12 w-full items-center gap-3 rounded-md px-3 text-base transition-colors active:bg-accent';

export const FeedsList = ({ newsletterId }: Props) => {
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [deletingFeed, setDeletingFeed] = useState<Feed | null>(null);
  const [actionsFeed, setActionsFeed] = useState<Feed | null>(null);
  const [search, setSearch] = useState('');

  const { data: feeds } = useSuspenseQuery(feedsOptions(newsletterId));

  const query = search.trim().toLowerCase();
  const filteredFeeds = (
    query
      ? feeds.filter((f) => {
          const name = (f.alias || f.title).toLowerCase();
          const desc = f.description.toLowerCase();
          return name.includes(query) || desc.includes(query);
        })
      : feeds
  ).toSorted((a, b) => (a.alias || a.title).localeCompare(b.alias || b.title));

  const deleteFeed = useDeleteFeed(newsletterId, () => {
    toast.success('Feed deleted!');
    setDeletingFeed(null);
  });

  const updateStatus = useUpdateFeedStatus(newsletterId, (status) => {
    toast.success(status === 'active' ? 'Feed resumed!' : 'Feed paused.');
  });

  const pendingFeedId = updateStatus.isPending
    ? updateStatus.variables.feedId
    : undefined;

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        {feeds.length > 0 && (
          <InputGroup className="sm:flex-1">
            <InputGroupAddon>
              <Search />
            </InputGroupAddon>
            <InputGroupInput
              placeholder="Search feeds..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </InputGroup>
        )}
        <Button
          className="h-11 sm:ml-auto sm:h-9"
          onClick={() => setAddDialogOpen(true)}
        >
          Add Feed
        </Button>
      </div>

      {feeds.length === 0 ? (
        <div className="space-y-3 rounded-lg border border-dashed py-12 text-center">
          <p className="text-sm text-muted-foreground">No feeds added yet.</p>
          <Button onClick={() => setAddDialogOpen(true)}>
            Add your first feed
          </Button>
        </div>
      ) : query && filteredFeeds.length === 0 ? (
        <div className="rounded-lg border border-dashed py-12 text-center">
          <p className="text-sm text-muted-foreground">
            No feeds match &ldquo;{query}&rdquo;.
          </p>
        </div>
      ) : (
        <ListPanel header="Feed">
          {filteredFeeds.map((feed) => {
            const isActive = feed.status === 'active';

            return (
              <div key={feed.id} className={cn('group', listRowClass)}>
                <Link
                  to="/newsletters/$newsletterId/feeds/$feedId"
                  params={{ newsletterId, feedId: feed.id }}
                  className="min-w-0 flex-1"
                >
                  <p className="flex items-center gap-2">
                    <span
                      className={cn(
                        'truncate font-medium group-hover:underline',
                        !isActive && 'text-muted-foreground',
                      )}
                    >
                      {feed.alias || feed.title}
                    </span>
                    <FeedStatusBadge status={feed.status} />
                    <FeedHealthBadge health={feed.health} />
                  </p>
                  <p className="mt-0.5 hidden truncate font-mono text-xs text-muted-foreground md:block">
                    {feed.url}
                  </p>
                  {feed.description && (
                    <p className="mt-0.5 truncate text-sm text-muted-foreground md:hidden">
                      {feed.description}
                    </p>
                  )}
                </Link>

                <Button
                  variant="ghost"
                  size="icon"
                  className="size-11 shrink-0 text-muted-foreground md:hidden"
                  onClick={() => setActionsFeed(feed)}
                  aria-label={`Actions for ${feed.alias || feed.title}`}
                >
                  <MoreVertical className="size-5" />
                </Button>

                <div className="hidden shrink-0 items-center gap-0.5 text-muted-foreground md:flex">
                  <a
                    href={feed.htmlUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    aria-label="Visit feed website"
                    className={buttonVariants({
                      variant: 'ghost',
                      size: 'icon-sm',
                    })}
                  >
                    <ExternalLink />
                  </a>
                  <Link
                    to="/newsletters/$newsletterId/feeds/$feedId"
                    params={{ newsletterId, feedId: feed.id }}
                    aria-label="Configure feed"
                    className={buttonVariants({
                      variant: 'ghost',
                      size: 'icon-sm',
                    })}
                  >
                    <Settings />
                  </Link>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() =>
                      updateStatus.mutate({
                        feedId: feed.id,
                        status: isActive ? 'inactive' : 'active',
                      })
                    }
                    disabled={pendingFeedId === feed.id}
                    aria-label={isActive ? 'Pause feed' : 'Resume feed'}
                    title={isActive ? 'Pause feed' : 'Resume feed'}
                  >
                    {isActive ? <Pause /> : <Play />}
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => setDeletingFeed(feed)}
                    aria-label="Delete feed"
                  >
                    <Trash2 />
                  </Button>
                </div>
              </div>
            );
          })}
        </ListPanel>
      )}

      <Sheet
        open={!!actionsFeed}
        onOpenChange={(open) => {
          if (!open) setActionsFeed(null);
        }}
      >
        <SheetContent
          side="bottom"
          showCloseButton={false}
          className="rounded-t-xl p-2 pb-safe-b text-sm md:hidden"
        >
          {actionsFeed && (
            <>
              <SheetTitle className="px-3 py-3 text-muted-foreground">
                {actionsFeed.alias || actionsFeed.title}
              </SheetTitle>
              <a
                href={actionsFeed.htmlUrl}
                target="_blank"
                rel="noopener noreferrer"
                className={cn(sheetActionClass)}
                onClick={() => setActionsFeed(null)}
              >
                <ExternalLink className="size-5" />
                Visit website
              </a>
              <Link
                to="/newsletters/$newsletterId/feeds/$feedId"
                params={{ newsletterId, feedId: actionsFeed.id }}
                className={cn(sheetActionClass)}
              >
                <Settings className="size-5" />
                Configure feed
              </Link>
              <button
                type="button"
                className={cn(sheetActionClass)}
                onClick={() => {
                  updateStatus.mutate({
                    feedId: actionsFeed.id,
                    status:
                      actionsFeed.status === 'active' ? 'inactive' : 'active',
                  });
                  setActionsFeed(null);
                }}
              >
                {actionsFeed.status === 'active' ? (
                  <Pause className="size-5" />
                ) : (
                  <Play className="size-5" />
                )}
                {actionsFeed.status === 'active' ? 'Pause feed' : 'Resume feed'}
              </button>
              <button
                type="button"
                className={cn(sheetActionClass, 'text-destructive')}
                onClick={() => {
                  setDeletingFeed(actionsFeed);
                  setActionsFeed(null);
                }}
              >
                <Trash2 className="size-5" />
                Delete feed
              </button>
            </>
          )}
        </SheetContent>
      </Sheet>

      <AddFeedDialog
        newsletterId={newsletterId}
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
      />

      <AlertDialog
        open={!!deletingFeed}
        onOpenChange={(open) => {
          if (!open) {
            setDeletingFeed(null);
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete feed?</AlertDialogTitle>
            <AlertDialogDescription>
              &ldquo;{deletingFeed?.alias || deletingFeed?.title}&rdquo; will be
              permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (deletingFeed) {
                  deleteFeed.mutate(deletingFeed.id);
                }
              }}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};
