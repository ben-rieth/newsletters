import { useRef, useState } from 'react';
import type { ComponentProps } from 'react';
import { useQuery } from '@tanstack/react-query';
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
import { getErrorMessage } from '#/lib/errors';
import { pluralize } from '../lib/format';
import { newslettersOptions } from '../queries/newsletters';
import useImportNewsletters, {
  parseExportFile,
} from '../queries/hooks/useImportNewsletters';
import type { NewslettersExport } from '../queries/hooks/useImportNewsletters';

const MAX_NAMES_IN_CONFIRM = 8;

const normalizeName = (name: string) => name.trim().toLocaleLowerCase();

type Props = {
  variant?: ComponentProps<typeof Button>['variant'];
  onImported?: () => void;
};

export const ImportNewslettersButton = ({
  variant = 'outline',
  onImported,
}: Props) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [pendingExport, setPendingExport] = useState<NewslettersExport | null>(
    null,
  );
  const { data: existing } = useQuery(newslettersOptions);
  const importNewsletters = useImportNewsletters();

  const incoming = pendingExport?.newsletters ?? [];
  const existingNames = new Set(
    (existing ?? []).map((newsletter) => normalizeName(newsletter.name)),
  );
  const isDuplicate = (name: string) => existingNames.has(normalizeName(name));
  const duplicateCount = incoming.filter((n) => isDuplicate(n.name)).length;
  const hiddenCount = incoming.length - MAX_NAMES_IN_CONFIRM;

  const closeConfirm = () => setPendingExport(null);

  return (
    <>
      <input
        ref={inputRef}
        type="file"
        accept="application/json,.json"
        className="hidden"
        aria-hidden="true"
        tabIndex={-1}
        onChange={async (event) => {
          const file = event.target.files?.[0];
          event.target.value = '';
          if (!file) {
            return;
          }
          try {
            setPendingExport(await parseExportFile(file));
          } catch (error) {
            toast.error(getErrorMessage(error));
          }
        }}
      />
      <Button
        variant={variant}
        onClick={() => inputRef.current?.click()}
        disabled={importNewsletters.isPending}
      >
        {importNewsletters.isPending ? 'Importing…' : 'Import from file'}
      </Button>

      <AlertDialog
        open={pendingExport !== null}
        onOpenChange={(open) =>
          !open && !importNewsletters.isPending && closeConfirm()
        }
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              Import {pluralize(incoming.length, 'newsletter')}?
            </AlertDialogTitle>
            <AlertDialogDescription>
              Each one is created as a new newsletter. Feeds not already in
              Slowfeed are added in the background.
              {duplicateCount > 0 &&
                ` ${duplicateCount === 1 ? 'One shares a name' : `${duplicateCount} share names`} with a newsletter you already have and will be created as a duplicate.`}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <ul className="space-y-1 text-sm">
            {incoming.slice(0, MAX_NAMES_IN_CONFIRM).map((newsletter, i) => (
              <li key={i} className="flex items-baseline gap-2">
                <span className="truncate">{newsletter.name}</span>
                <span className="shrink-0 text-xs text-muted-foreground">
                  {pluralize(newsletter.feeds?.length ?? 0, 'feed')}
                  {isDuplicate(newsletter.name) && (
                    <span className="text-foreground"> · already exists</span>
                  )}
                </span>
              </li>
            ))}
            {hiddenCount > 0 && (
              <li className="text-muted-foreground">and {hiddenCount} more</li>
            )}
          </ul>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={importNewsletters.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={importNewsletters.isPending}
              onClick={() => {
                if (!pendingExport) {
                  return;
                }
                importNewsletters.mutate(pendingExport, {
                  onSuccess: () => {
                    closeConfirm();
                    onImported?.();
                  },
                });
              }}
            >
              {importNewsletters.isPending
                ? 'Importing…'
                : `Import ${pluralize(incoming.length, 'newsletter')}`}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
};
