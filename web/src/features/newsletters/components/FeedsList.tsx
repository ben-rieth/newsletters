import { useState } from 'react';
import { useSuspenseQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Button } from '#/components/ui/button';
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
import { ViewToggle, type View } from '#/components/ViewToggle';
import { feedsOptions, type Feed } from '../queries/feeds';
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

  const { data: feeds } = useSuspenseQuery(feedsOptions(newsletterId));

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

      {view === 'cards' ? (
        <FeedsCards
          feeds={feeds ?? []}
          newsletterId={newsletterId}
          onDeleteClick={setDeletingFeed}
          onAddClick={() => setAddDialogOpen(true)}
        />
      ) : (
        <FeedsTable
          feeds={feeds ?? []}
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
