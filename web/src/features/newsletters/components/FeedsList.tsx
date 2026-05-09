import { useState } from 'react';
import { useSuspenseQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Search } from 'lucide-react';
import { Button } from '#/components/ui/button';
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
import { ViewToggle } from '#/components/ViewToggle';
import type { View } from '#/components/ViewToggle';
import { feedsOptions } from '../queries/feeds';
import type { Feed } from '../queries/feeds';
import { AddFeedDialog } from './AddFeedDialog';
import { FeedsCards } from './FeedsCards';
import { FeedsTable } from './FeedsTable';
import useDeleteFeed from '../queries/hooks/useDeleteFeed';

type Props = {
  newsletterId: string;
};

export const FeedsList = ({ newsletterId }: Props) => {
  const [view, setView] = useState<View>('cards');
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [deletingFeed, setDeletingFeed] = useState<Feed | null>(null);
  const [search, setSearch] = useState('');

  const { data: feeds } = useSuspenseQuery(feedsOptions(newsletterId));

  const query = search.trim().toLowerCase();
  const filteredFeeds = query
    ? feeds.filter((f) => {
        const name = (f.alias || f.title).toLowerCase();
        const desc = f.description.toLowerCase();
        return name.includes(query) || desc.includes(query);
      })
    : feeds;

  const deleteFeed = useDeleteFeed(newsletterId, () => {
    toast.success('Feed deleted!');
    setDeletingFeed(null);
  });

  return (
    <>
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Feeds</h3>
        <div className="flex items-center gap-2">
          <ViewToggle value={view} onChange={setView} />
          <Button size="sm" onClick={() => setAddDialogOpen(true)}>
            Add Feed
          </Button>
        </div>
      </div>

      {feeds.length > 0 && (
        <InputGroup>
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

      {query && filteredFeeds.length === 0 ? (
        <div className="py-6 text-center rounded-md border border-dashed">
          <p className="text-sm text-muted-foreground">
            No feeds match &ldquo;{query}&rdquo;.
          </p>
        </div>
      ) : view === 'cards' ? (
        <FeedsCards
          feeds={filteredFeeds}
          newsletterId={newsletterId}
          onDeleteClick={setDeletingFeed}
          onAddClick={() => setAddDialogOpen(true)}
        />
      ) : (
        <FeedsTable
          feeds={filteredFeeds}
          newsletterId={newsletterId}
          onDeleteClick={setDeletingFeed}
          onAddClick={() => setAddDialogOpen(true)}
        />
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
    </>
  );
};
