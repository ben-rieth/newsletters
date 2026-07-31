import { useState } from 'react';
import { Pencil, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '#/components/ui/button';
import { ListPanel, listRowClass } from '#/components/ListPanel';
import { EmptyState } from '#/components/EmptyState';
import type { components } from '#/api/schema';
import useDeleteFeedFilter from '../queries/hooks/useDeleteFeedFilter';
import useUpdateFeedFilter from '../queries/hooks/useUpdateFeedFilter';
import { FeedFilterForm } from './FeedFilterForm';
import { getErrorMessage } from '#/lib/errors';

type FeedFilter = components['schemas']['FeedFilter'];

const FIELD_LABELS: Record<string, string> = {
  name: 'Name',
  url: 'URL',
};

const OPERATOR_LABELS: Record<string, string> = {
  contains: 'contains',
  does_not_contain: 'does not contain',
};

type Props = {
  newsletterId: string;
  feedId: string;
  filters: FeedFilter[];
};

export const FeedFiltersList = ({ newsletterId, feedId, filters }: Props) => {
  const [editingId, setEditingId] = useState<string | null>(null);

  const deleteFilter = useDeleteFeedFilter(newsletterId, feedId, () => {
    toast.success('Filter deleted!');
  });

  const updateFilter = useUpdateFeedFilter(newsletterId, feedId, () => {
    toast.success('Filter updated!');
    setEditingId(null);
  });

  if (filters.length === 0) {
    return (
      <EmptyState
        className="px-0 md:px-0"
        title="No filters"
        description="Every item from this feed is included. Add a filter to drop items whose name or URL matches a phrase — useful for keeping a noisy source in its lane."
      />
    );
  }

  return (
    <ListPanel header="Filter">
      {filters.map((filter) =>
        editingId === filter.id ? (
          <div key={filter.id} className="px-5 py-4">
            <FeedFilterForm
              filter={filter}
              onSubmit={async (values) => {
                await updateFilter.mutateAsync({
                  filterId: filter.id,
                  body: values,
                });
              }}
              onCancel={() => setEditingId(null)}
              isPending={updateFilter.isPending}
              error={
                updateFilter.isError
                  ? getErrorMessage(updateFilter.error)
                  : undefined
              }
            />
          </div>
        ) : (
          <div key={filter.id} className={listRowClass}>
            <div className="flex flex-wrap items-center gap-2 text-sm">
              <span className="font-medium">
                {FIELD_LABELS[filter.field] ?? filter.field}
              </span>
              <span className="text-muted-foreground">
                {OPERATOR_LABELS[filter.operator] ?? filter.operator}
              </span>
              <span className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
                {filter.pattern}
              </span>
            </div>
            <div className="flex shrink-0 items-center gap-0.5 text-muted-foreground">
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label="Edit filter"
                disabled={deleteFilter.isPending}
                onClick={() => setEditingId(filter.id)}
              >
                <Pencil />
              </Button>
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label="Delete filter"
                disabled={deleteFilter.isPending}
                onClick={() => deleteFilter.mutate(filter.id)}
              >
                <Trash2 />
              </Button>
            </div>
          </div>
        ),
      )}
    </ListPanel>
  );
};
