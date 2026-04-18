import { useState } from 'react';
import { Pencil, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '#/components/ui/button';
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
      <p className="text-sm text-muted-foreground">No filters added yet.</p>
    );
  }

  return (
    <div className="space-y-1">
      {filters.map((filter, index) => (
        <div key={filter.id}>
          {index > 0 && (
            <div className="flex items-center gap-2 py-1">
              <div className="flex-1 border-t" />
              <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                OR
              </span>
              <div className="flex-1 border-t" />
            </div>
          )}
          {editingId === filter.id ? (
            <div className="rounded-md border px-3 py-3">
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
            <div className="flex items-center justify-between gap-2 rounded-md border px-3 py-2 text-sm">
              <div className="flex items-center gap-2">
                <span className="font-medium">
                  {FIELD_LABELS[filter.field] ?? filter.field}
                </span>
                <span className="text-muted-foreground">
                  {OPERATOR_LABELS[filter.operator] ?? filter.operator}
                </span>
                <span className="font-mono text-xs bg-muted rounded px-1.5 py-0.5">
                  {filter.pattern}
                </span>
              </div>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label="Edit filter"
                  disabled={deleteFilter.isPending}
                  onClick={() => setEditingId(filter.id)}
                >
                  <Pencil className="size-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label="Delete filter"
                  disabled={deleteFilter.isPending}
                  onClick={() => deleteFilter.mutate(filter.id)}
                >
                  <Trash2 className="size-3.5" />
                </Button>
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
};
