import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { feedPreviewOptions } from '../queries/feeds';
import type { ItemPreview } from '../queries/feeds';
import { Button } from '#/components/ui/button';
import { Badge } from '#/components/ui/badge';
import { ListPanel } from '#/components/ListPanel';
import { cn } from '#/lib/utils';
import { getErrorMessage } from '#/lib/errors';

type Props = {
  newsletterId: string;
  feedId: string;
};

const FilteredBadge = ({
  field,
  operator,
  pattern,
}: {
  field: string;
  operator: string;
  pattern: string;
}) => (
  <Badge variant="secondary" className="text-xs font-mono shrink-0">
    {field} {operator} &ldquo;{pattern}&rdquo;
  </Badge>
);

const PreviewItem = ({ item }: { item: ItemPreview }) => {
  const isFiltered = !!item.matchedFilter;

  return (
    <div
      className={cn(
        'flex items-start gap-3 px-5 py-3.5 text-sm transition-colors hover:bg-accent/40',
        isFiltered && 'opacity-50',
      )}
    >
      <div className="flex-1 min-w-0">
        <a
          href={item.url}
          target="_blank"
          rel="noopener noreferrer"
          className={cn(
            'font-medium hover:underline truncate block',
            isFiltered && 'line-through text-muted-foreground',
          )}
        >
          {item.title}
        </a>
        <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">
          {item.url}
        </p>
        {isFiltered && item.matchedFilter && (
          <div className="mt-1 flex items-center gap-2 flex-wrap">
            <span className="text-xs text-muted-foreground">filtered by</span>
            <FilteredBadge
              field={item.matchedFilter.field}
              operator={item.matchedFilter.operator}
              pattern={item.matchedFilter.pattern}
            />
          </div>
        )}
      </div>
    </div>
  );
};

export const FeedPreview = ({ newsletterId, feedId }: Props) => {
  const [hasRequested, setHasRequested] = useState(false);
  const { data, isFetching, isError, error, refetch } = useQuery(
    feedPreviewOptions(newsletterId, feedId),
  );

  const handlePreview = () => {
    setHasRequested(true);
    refetch();
  };

  const passCount = data?.filter((item) => !item.matchedFilter).length ?? 0;
  const filteredCount =
    data?.filter((item) => !!item.matchedFilter).length ?? 0;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <p className="text-sm text-muted-foreground">
            Fetch the latest items from this feed and preview which would be
            filtered out.
          </p>
          {hasRequested && !isFetching && data && (
            <p className="text-sm text-muted-foreground">
              {passCount} passing · {filteredCount} filtered
            </p>
          )}
        </div>
        <Button variant="outline" onClick={handlePreview} disabled={isFetching}>
          {isFetching ? 'Loading…' : 'Preview Filters'}
        </Button>
      </div>

      {hasRequested && isError && (
        <p className="text-sm text-destructive">{getErrorMessage(error)}</p>
      )}

      {hasRequested && !isFetching && data && data.length === 0 && (
        <p className="text-sm text-muted-foreground">
          No items found in this feed.
        </p>
      )}

      {hasRequested && !isFetching && data && data.length > 0 && (
        <ListPanel header="Item">
          {data.map((item) => (
            <PreviewItem key={item.id} item={item} />
          ))}
        </ListPanel>
      )}
    </div>
  );
};
