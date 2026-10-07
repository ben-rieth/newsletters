import { CircleAlert, LoaderCircle, RotateCw, X } from 'lucide-react';
import { ListPanel, listRowClass } from '#/components/ListPanel';
import { Badge } from '#/components/ui/badge';
import { Button } from '#/components/ui/button';
import type { FeedImport } from '../queries/feedImports';
import useRetryFeedImport from '../queries/hooks/useRetryFeedImport';
import useDeleteFeedImport from '../queries/hooks/useDeleteFeedImport';

type Props = {
  newsletterId: string;
  imports: FeedImport[];
};

export const FeedImportsList = ({ newsletterId, imports }: Props) => {
  const retry = useRetryFeedImport(newsletterId);
  const remove = useDeleteFeedImport(newsletterId);

  if (imports.length === 0) {
    return null;
  }

  return (
    <div aria-live="polite">
      <ListPanel>
        {imports.map((feedImport) => {
          const failed = feedImport.state === 'failed';
          const name = feedImport.alias || feedImport.url;
          const busy =
            (retry.isPending && retry.variables === feedImport.id) ||
            (remove.isPending && remove.variables === feedImport.id);

          return (
            <div key={feedImport.id} className={listRowClass}>
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
        })}
      </ListPanel>
    </div>
  );
};
