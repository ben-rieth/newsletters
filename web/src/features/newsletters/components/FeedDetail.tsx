import { useState } from 'react';
import { useSuspenseQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { feedDetailOptions } from '../queries/feeds';
import useUpdateFeed from '../queries/hooks/useUpdateFeed';
import useAddFeedFilter from '../queries/hooks/useAddFeedFilter';
import { EditFeedForm } from './EditFeedForm';
import { FeedFilterForm } from './FeedFilterForm';
import { FeedFiltersList } from './FeedFiltersList';
import { FeedPreview } from './FeedPreview';
import { Button } from '#/components/ui/button';
import { getErrorMessage } from '#/lib/errors';

type Props = {
  newsletterId: string;
  feedId: string;
};

export const FeedDetail = ({ newsletterId, feedId }: Props) => {
  const [showAddFilter, setShowAddFilter] = useState(false);
  const { data: feed } = useSuspenseQuery(feedDetailOptions(newsletterId, feedId));

  const updateFeed = useUpdateFeed(newsletterId, feedId, () => {
    toast.success('Feed updated!');
  });

  const addFilter = useAddFeedFilter(newsletterId, feedId, () => {
    toast.success('Filter added!');
    setShowAddFilter(false);
  });

  return (
    <div className="space-y-8">
      <section className="space-y-3">
        <h3 className="text-base font-semibold">Feed Info</h3>
        <EditFeedForm
          feed={feed}
          onSubmit={async (alias) => {
            await updateFeed.mutateAsync({ alias });
          }}
          isPending={updateFeed.isPending}
          error={updateFeed.isError ? getErrorMessage(updateFeed.error) : undefined}
        />
      </section>

      <section className="space-y-3">
        <h3 className="text-base font-semibold">Filters</h3>
        <p className="text-sm text-muted-foreground">
          Filters hide articles from this feed when the condition matches.
        </p>
        <FeedFiltersList newsletterId={newsletterId} feedId={feedId} filters={feed.filters ?? []} />
        {showAddFilter ? (
          <div className="pt-2 border-t">
            <FeedFilterForm
              onSubmit={async (values) => {
                await addFilter.mutateAsync(values);
              }}
              onCancel={() => setShowAddFilter(false)}
              isPending={addFilter.isPending}
              error={addFilter.isError ? getErrorMessage(addFilter.error) : undefined}
            />
          </div>
        ) : (
          <Button size="sm" variant="outline" onClick={() => setShowAddFilter(true)}>
            Add Filter
          </Button>
        )}
        <div className="pt-4 border-t">
          <FeedPreview newsletterId={newsletterId} feedId={feedId} />
        </div>
      </section>
    </div>
  );
};
