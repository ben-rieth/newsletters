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
import type { LucideIcon } from 'lucide-react';
import { cn } from '#/lib/utils';
import { ListPanel, listRowClass } from '#/components/ListPanel';
import { EmptyState } from '#/components/EmptyState';
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
  'flex min-h-12 w-full items-center gap-3 rounded-md px-3 text-base transition-colors active:bg-accent disabled:opacity-50';

type FeedAction = {
  key: string;
  label: string;
  icon: LucideIcon;
  destructive?: boolean;
  disabled?: boolean;
} & (
  | { kind: 'external'; href: string }
  | { kind: 'configure'; feedId: string }
  | { kind: 'action'; onSelect: () => void }
);

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

  const buildFeedActions = (feed: Feed): FeedAction[] => {
    const isActive = feed.status === 'active';

    return [
      {
        kind: 'external',
        key: 'visit',
        label: 'Visit website',
        icon: ExternalLink,
        href: feed.htmlUrl,
      },
      {
        kind: 'configure',
        key: 'configure',
        label: 'Configure feed',
        icon: Settings,
        feedId: feed.id,
      },
      {
        kind: 'action',
        key: 'status',
        label: isActive ? 'Pause feed' : 'Resume feed',
        icon: isActive ? Pause : Play,
        disabled: updateStatus.isPending,
        onSelect: () =>
          updateStatus.mutate({
            feedId: feed.id,
            status: isActive ? 'inactive' : 'active',
          }),
      },
      {
        kind: 'action',
        key: 'delete',
        label: 'Delete feed',
        icon: Trash2,
        destructive: true,
        onSelect: () => setDeletingFeed(feed),
      },
    ];
  };

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
        <EmptyState
          className="px-0 md:px-0"
          title="No feeds yet"
          description="Paste the RSS or Atom URL of any site, blog, YouTube channel, or podcast. Slowfeed checks each one resolves before adding it."
          action={
            <Button onClick={() => setAddDialogOpen(true)}>
              Add your first feed
            </Button>
          }
        />
      ) : query && filteredFeeds.length === 0 ? (
        <EmptyState
          className="px-0 md:px-0"
          title="No matches"
          description={
            <>
              None of your {feeds.length} feeds match &ldquo;{query}&rdquo;.
            </>
          }
          action={
            <Button variant="outline" onClick={() => setSearch('')}>
              Clear search
            </Button>
          }
        />
      ) : (
        <ListPanel>
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
                  {buildFeedActions(feed).map((action) => {
                    const Icon = action.icon;
                    const iconButtonClass = buttonVariants({
                      variant: 'ghost',
                      size: 'icon-sm',
                    });

                    if (action.kind === 'external') {
                      return (
                        <a
                          key={action.key}
                          href={action.href}
                          target="_blank"
                          rel="noopener noreferrer"
                          aria-label={action.label}
                          title={action.label}
                          className={iconButtonClass}
                        >
                          <Icon />
                        </a>
                      );
                    }

                    if (action.kind === 'configure') {
                      return (
                        <Link
                          key={action.key}
                          to="/newsletters/$newsletterId/feeds/$feedId"
                          params={{ newsletterId, feedId: action.feedId }}
                          aria-label={action.label}
                          title={action.label}
                          className={iconButtonClass}
                        >
                          <Icon />
                        </Link>
                      );
                    }

                    return (
                      <Button
                        key={action.key}
                        variant="ghost"
                        size="icon-sm"
                        onClick={action.onSelect}
                        disabled={action.disabled}
                        aria-label={action.label}
                        title={action.label}
                      >
                        <Icon />
                      </Button>
                    );
                  })}
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
              {buildFeedActions(actionsFeed).map((action) => {
                const Icon = action.icon;
                const className = cn(
                  sheetActionClass,
                  action.destructive && 'text-destructive',
                );
                const content = (
                  <>
                    <Icon className="size-5" />
                    {action.label}
                  </>
                );

                if (action.kind === 'external') {
                  return (
                    <a
                      key={action.key}
                      href={action.href}
                      target="_blank"
                      rel="noopener noreferrer"
                      className={className}
                      onClick={() => setActionsFeed(null)}
                    >
                      {content}
                    </a>
                  );
                }

                if (action.kind === 'configure') {
                  return (
                    <Link
                      key={action.key}
                      to="/newsletters/$newsletterId/feeds/$feedId"
                      params={{ newsletterId, feedId: action.feedId }}
                      className={className}
                      onClick={() => setActionsFeed(null)}
                    >
                      {content}
                    </Link>
                  );
                }

                return (
                  <button
                    key={action.key}
                    type="button"
                    className={className}
                    disabled={action.disabled}
                    onClick={() => {
                      action.onSelect();
                      setActionsFeed(null);
                    }}
                  >
                    {content}
                  </button>
                );
              })}
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
