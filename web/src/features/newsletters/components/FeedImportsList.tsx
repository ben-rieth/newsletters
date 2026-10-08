import { useState } from 'react';
import { CircleAlert, LoaderCircle, RotateCw, X } from 'lucide-react';
import { ListPanel, listRowClass } from '#/components/ListPanel';
import { Badge } from '#/components/ui/badge';
import { Button } from '#/components/ui/button';
import { pluralize } from '../lib/format';
import type { FeedImport } from '../queries/feedImports';
import useRetryFeedImport from '../queries/hooks/useRetryFeedImport';
import useDeleteFeedImport from '../queries/hooks/useDeleteFeedImport';

type Props = {
  newsletterId: string;
  imports: FeedImport[];
};

type RowProps = {
  newsletterId: string;
  feedImport: FeedImport;
};

const FeedImportRow = ({ newsletterId, feedImport }: RowProps) => {
  const retry = useRetryFeedImport(newsletterId);
  const remove = useDeleteFeedImport(newsletterId);
  const failed = feedImport.state === 'failed';
  const name = feedImport.alias || feedImport.url;
  const busy = retry.isPending || remove.isPending;

  return (
    <div className={listRowClass}>
      <div className="min-w-0 flex-1">
        <p className="flex items-center gap-2">
          <span className="truncate font-medium text-muted-foreground">
            {name}
          </span>
          {failed ? (
            <Badge variant="destructive">
              <CircleAlert />
              Couldn&rsquo;t add
            </Badge>
          ) : (
            <Badge variant="outline" className="text-muted-foreground">
              <LoaderCircle className="motion-safe:animate-spin" />
              Importing…
            </Badge>
          )}
        </p>
        {failed ? (
          <p
            className="mt-0.5 line-clamp-2 text-xs break-words text-destructive"
            title={feedImport.error}
          >
            {feedImport.error}
          </p>
        ) : (
          feedImport.alias && (
            <p className="mt-0.5 truncate font-mono text-xs text-muted-foreground">
              {feedImport.url}
            </p>
          )
        )}
      </div>

      <div className="flex shrink-0 items-center gap-0.5 text-muted-foreground">
        {failed && (
          <Button
            variant="ghost"
            size="icon-sm"
            className="max-md:size-11"
            onClick={() => retry.mutate(feedImport.id)}
            disabled={busy}
            aria-label={`Retry adding ${name}`}
            title="Retry"
          >
            <RotateCw />
          </Button>
        )}
        <Button
          variant="ghost"
          size="icon-sm"
          className="max-md:size-11"
          onClick={() => remove.mutate(feedImport.id)}
          disabled={busy}
          aria-label={`Remove ${name}`}
          title="Remove"
        >
          <X />
        </Button>
      </div>
    </div>
  );
};

export const FeedImportsList = ({ newsletterId, imports }: Props) => {
  if (imports.length === 0) {
    return null;
  }

  return (
    <ListPanel>
      {imports.map((feedImport) => (
        <FeedImportRow
          key={feedImport.id}
          newsletterId={newsletterId}
          feedImport={feedImport}
        />
      ))}
    </ListPanel>
  );
};

// Live regions only announce changes to content that was already mounted, so
// this stays rendered even when there is nothing to say.
export const FeedImportsStatus = ({ imports }: { imports: FeedImport[] }) => {
  const pending = imports.filter((i) => i.state === 'pending').length;
  const failed = imports.filter((i) => i.state === 'failed').length;
  const [sawPending, setSawPending] = useState(pending > 0);

  if (pending > 0 && !sawPending) {
    setSawPending(true);
  }

  const messages = [
    pending > 0 && `Importing ${pluralize(pending, 'feed')}.`,
    failed > 0 && `${pluralize(failed, 'feed')} couldn\u2019t be added.`,
    sawPending && pending === 0 && failed === 0 && 'Import finished.',
  ].filter(Boolean);

  return (
    <p role="status" className="sr-only">
      {messages.join(' ')}
    </p>
  );
};
