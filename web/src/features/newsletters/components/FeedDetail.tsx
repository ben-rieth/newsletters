import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedDetailOptions } from '../queries/feeds';
import useUpdateFeed from '../queries/hooks/useUpdateFeed';
import useAddFeedFilter from '../queries/hooks/useAddFeedFilter';
import useDeleteFeed from '../queries/hooks/useDeleteFeed';
import useUpdateFeedStatus from '../queries/hooks/useUpdateFeedStatus';
import { EditFeedForm } from './EditFeedForm';
import { FeedFilterForm } from './FeedFilterForm';
import { FeedFiltersList } from './FeedFiltersList';
import { FeedPreview } from './FeedPreview';
import { FeedHealthAlert, FeedHealthBadge } from './FeedHealth';
import { FeedStatusBadge } from './FeedStatusBadge';
import { Button } from '#/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { SettingsRow, SettingsSection } from '#/components/SettingsRow';
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
import { getErrorMessage } from '#/lib/errors';

type Props = {
  newsletterId: string;
  feedId: string;
};

export const FeedDetail = ({ newsletterId, feedId }: Props) => {
  const [showAddFilter, setShowAddFilter] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const navigate = useNavigate();
  const { data: feed } = useSuspenseQuery(
    feedDetailOptions(newsletterId, feedId),
  );

  const filterCount = feed.filters?.length ?? 0;
  const isActive = feed.status === 'active';

  const updateFeed = useUpdateFeed(newsletterId, feedId, () => {
    toast.success('Feed updated!');
  });

  const addFilter = useAddFeedFilter(newsletterId, feedId, () => {
    toast.success('Filter added!');
    setShowAddFilter(false);
  });

  const updateStatus = useUpdateFeedStatus(newsletterId, (status) => {
    toast.success(status === 'active' ? 'Feed resumed!' : 'Feed paused.');
  });

  const deleteFeed = useDeleteFeed(newsletterId, () => {
    toast.success('Feed deleted!');
    navigate({
      to: '/newsletters/$newsletterId',
      params: { newsletterId },
    });
  });

  return (
    <div className="space-y-6">
      <header className="space-y-2">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-semibold tracking-tight">
            {feed.alias || feed.title}
          </h1>
          <FeedStatusBadge status={feed.status} />
          <FeedHealthBadge health={feed.health} />
        </div>
        <p className="truncate font-mono text-sm text-muted-foreground">
          {feed.url}
        </p>
        {feed.description && (
          <p className="text-sm text-muted-foreground">{feed.description}</p>
        )}
      </header>

      <FeedHealthAlert health={feed.health} />

      <Tabs defaultValue="details">
        <TabsList
          variant="line"
          className="w-full justify-start gap-6 border-b **:data-[slot=tabs-trigger]:flex-none **:data-[slot=tabs-trigger]:px-0"
        >
          <TabsTrigger value="details">Details</TabsTrigger>
          <TabsTrigger value="filters">
            Filters
            {filterCount > 0 && (
              <span className="ml-1.5 text-muted-foreground">
                {filterCount}
              </span>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="details" className="pt-6" keepMounted>
          <div className="space-y-10">
            <EditFeedForm
              feed={feed}
              onSubmit={async (alias) => {
                await updateFeed.mutateAsync({ alias });
              }}
              isPending={updateFeed.isPending}
              error={
                updateFeed.isError
                  ? getErrorMessage(updateFeed.error)
                  : undefined
              }
            />

            <SettingsSection>
              <SettingsRow
                title="Feed status"
                description={
                  isActive
                    ? 'Items from this feed are included in new issues.'
                    : 'This feed is paused. Its items are left out of new issues until you resume it.'
                }
              >
                <Button
                  variant="outline"
                  onClick={() =>
                    updateStatus.mutate({
                      feedId,
                      status: isActive ? 'inactive' : 'active',
                    })
                  }
                  disabled={updateStatus.isPending}
                >
                  {updateStatus.isPending
                    ? '…'
                    : isActive
                      ? 'Pause feed'
                      : 'Resume feed'}
                </Button>
              </SettingsRow>
            </SettingsSection>

            <SettingsSection>
              <SettingsRow
                title="Delete feed"
                description="Remove this feed from the newsletter. Past issues keep their items."
              >
                <Button
                  variant="outline"
                  className="border-destructive/40 text-destructive hover:bg-destructive/10 hover:text-destructive"
                  onClick={() => setDeleteDialogOpen(true)}
                >
                  Delete feed
                </Button>
              </SettingsRow>
            </SettingsSection>
          </div>
        </TabsContent>

        <TabsContent value="filters" className="pt-6">
          <div className="max-w-2xl space-y-6">
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">
                An article is hidden if it matches any of these filters.
              </p>
              <FeedFiltersList
                newsletterId={newsletterId}
                feedId={feedId}
                filters={feed.filters ?? []}
              />
              {showAddFilter ? (
                <div className="rounded-lg border p-4">
                  <FeedFilterForm
                    onSubmit={async (values) => {
                      await addFilter.mutateAsync(values);
                    }}
                    onCancel={() => setShowAddFilter(false)}
                    isPending={addFilter.isPending}
                    error={
                      addFilter.isError
                        ? getErrorMessage(addFilter.error)
                        : undefined
                    }
                  />
                </div>
              ) : (
                <Button
                  variant="outline"
                  onClick={() => setShowAddFilter(true)}
                >
                  Add Filter
                </Button>
              )}
            </div>

            <div className="border-t pt-6">
              <FeedPreview newsletterId={newsletterId} feedId={feedId} />
            </div>
          </div>
        </TabsContent>
      </Tabs>

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete feed?</AlertDialogTitle>
            <AlertDialogDescription>
              &ldquo;{feed.alias || feed.title}&rdquo; will be removed from this
              newsletter.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={deleteFeed.isPending}
              onClick={() => deleteFeed.mutate(feedId)}
            >
              {deleteFeed.isPending ? 'Deleting…' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};
