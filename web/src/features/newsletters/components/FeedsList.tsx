import { useState } from 'react';
import { Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ExternalLink, Search, Settings, Trash2 } from 'lucide-react';
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
import { feedsOptions } from '../queries/feeds';
import type { Feed } from '../queries/feeds';
import { AddFeedDialog } from './AddFeedDialog';
import { FeedHealthBadge } from './FeedHealth';
import useDeleteFeed from '../queries/hooks/useDeleteFeed';

type Props = {
  newsletterId: string;
};

export const FeedsList = ({ newsletterId }: Props) => {
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [deletingFeed, setDeletingFeed] = useState<Feed | null>(null);
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

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        {feeds.length > 0 && (
          <InputGroup className="flex-1">
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
        <Button className="ml-auto" onClick={() => setAddDialogOpen(true)}>
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
          {filteredFeeds.map((feed) => (
            <div key={feed.id} className={cn('group', listRowClass)}>
              <Link
                to="/newsletters/$newsletterId/feeds/$feedId"
                params={{ newsletterId, feedId: feed.id }}
                className="min-w-0 flex-1"
              >
                <p className="flex items-center gap-2">
                  <span className="truncate font-medium group-hover:underline">
                    {feed.alias || feed.title}
                  </span>
                  <FeedHealthBadge health={feed.health} />
                </p>
                <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">
                  {feed.url}
                </p>
              </Link>

              <div className="flex shrink-0 items-center gap-0.5 text-muted-foreground">
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
                  onClick={() => setDeletingFeed(feed)}
                  aria-label="Delete feed"
                >
                  <Trash2 />
                </Button>
              </div>
            </div>
          ))}
        </ListPanel>
      )}

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
