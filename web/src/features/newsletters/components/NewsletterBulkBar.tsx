import { useState } from 'react';
import { toast } from 'sonner';
import { cn } from '#/lib/utils';
import { Button } from '#/components/ui/button';
import { Checkbox } from '#/components/ui/checkbox';
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
import { pluralize } from '../lib/format';
import type { Newsletter } from '../queries/newsletters';
import useBulkDeleteNewsletters from '../queries/hooks/useBulkDeleteNewsletters';
import useBulkUpdateNewsletterStatus from '../queries/hooks/useBulkUpdateNewsletterStatus';
import useExportNewsletters from '../queries/hooks/useExportNewsletters';

const MAX_NAMES_IN_CONFIRM = 5;
const MAX_BULK_SELECTION = 100;

type Props = {
  newsletters: Newsletter[];
  selected: Newsletter[];
  onSelectAll: (selectAll: boolean) => void;
  onActionComplete: (action: 'status' | 'delete' | 'export') => void;
};

export const NewsletterBulkBar = ({
  newsletters,
  selected,
  onSelectAll,
  onActionComplete,
}: Props) => {
  const [confirmDeleteOpen, setConfirmDeleteOpen] = useState(false);

  const selectedIds = selected.map((newsletter) => newsletter.id);
  const hasActive = selected.some((n) => n.status === 'active');
  const hasPaused = selected.some((n) => n.status !== 'active');
  const allSelected =
    newsletters.length > 0 && selected.length === newsletters.length;

  const updateStatus = useBulkUpdateNewsletterStatus(({ ids, status }) => {
    toast.success(
      `${status === 'active' ? 'Resumed' : 'Paused'} ${pluralize(ids.length, 'newsletter')}.`,
    );
    onActionComplete('status');
  });

  const bulkDelete = useBulkDeleteNewsletters((ids) => {
    toast.success(`Deleted ${pluralize(ids.length, 'newsletter')}.`);
    setConfirmDeleteOpen(false);
    onActionComplete('delete');
  });

  const exportSelected = useExportNewsletters();

  const isPending =
    updateStatus.isPending || bulkDelete.isPending || exportSelected.isPending;
  const noneSelected = selected.length === 0;
  const tooManySelected = selected.length > MAX_BULK_SELECTION;
  const actionsDisabled = isPending || tooManySelected;

  const changeStatus = (status: 'active' | 'inactive') =>
    updateStatus.mutate({
      ids: selected
        .filter((n) => (n.status === 'active') !== (status === 'active'))
        .map((n) => n.id),
      status,
    });

  const hiddenNameCount = selected.length - MAX_NAMES_IN_CONFIRM;

  return (
    <div className="sticky top-0 z-10 flex flex-wrap items-center justify-between gap-x-4 gap-y-2 border-y border-border bg-background px-1 py-3">
      <label className="flex items-center gap-3 text-sm">
        <Checkbox
          checked={allSelected}
          indeterminate={!noneSelected && !allSelected}
          onCheckedChange={(checked) => onSelectAll(checked)}
          aria-label="Select all newsletters"
        />
        <span
          className={cn(
            'tabular-nums',
            tooManySelected ? 'text-destructive' : 'text-muted-foreground',
          )}
        >
          {noneSelected ? 'Select all' : `${selected.length} selected`}
          {tooManySelected && ` · max ${MAX_BULK_SELECTION}`}
        </span>
      </label>

      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          className="max-md:h-11"
          disabled={!hasActive || actionsDisabled}
          onClick={() => changeStatus('inactive')}
        >
          Pause
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="max-md:h-11"
          disabled={!hasPaused || actionsDisabled}
          onClick={() => changeStatus('active')}
        >
          Resume
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="max-md:h-11"
          disabled={noneSelected || actionsDisabled}
          onClick={() =>
            exportSelected.mutate(selectedIds, {
              onSuccess: () => onActionComplete('export'),
            })
          }
        >
          {exportSelected.isPending ? 'Exporting…' : 'Export'}
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="max-md:h-11 border-destructive/40 text-destructive hover:bg-destructive/10 hover:text-destructive"
          disabled={noneSelected || actionsDisabled}
          onClick={() => setConfirmDeleteOpen(true)}
        >
          Delete
        </Button>
      </div>

      <AlertDialog open={confirmDeleteOpen} onOpenChange={setConfirmDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Delete {pluralize(selected.length, 'newsletter')}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              These newsletters, their feeds, and their sent issues will be
              permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <ul className="space-y-1 text-sm">
            {selected.slice(0, MAX_NAMES_IN_CONFIRM).map((newsletter) => (
              <li key={newsletter.id} className="truncate">
                {newsletter.name}
              </li>
            ))}
            {hiddenNameCount > 0 && (
              <li className="text-muted-foreground">
                and {hiddenNameCount} more
              </li>
            )}
          </ul>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={bulkDelete.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={bulkDelete.isPending}
              onClick={() => bulkDelete.mutate(selectedIds)}
            >
              {bulkDelete.isPending ? 'Deleting…' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};
