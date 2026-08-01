import type { ReactNode } from 'react';
import { cn } from '#/lib/utils';
import { Button } from '#/components/ui/button';
import { getErrorMessage } from '#/lib/errors';

type Props = {
  title?: string;
  description?: ReactNode;
  /** Raw failure, tucked behind a disclosure so backend phrasing stays out of the copy. */
  error?: unknown;
  onRetry?: () => void;
  action?: ReactNode;
  className?: string;
};

export const ErrorState = ({
  title = 'That didn’t load',
  description = 'The request didn’t come back. It is usually a connection blip — trying again is worth a shot.',
  error,
  onRetry,
  action,
  className,
}: Props) => {
  const detail = error == null ? null : getErrorMessage(error);

  return (
    <div className={cn('max-w-sm px-4 py-10 md:px-5', className)}>
      <h3 className="font-serif text-xl font-medium">{title}</h3>
      <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
        {description}
      </p>

      {(onRetry != null || action != null) && (
        <div className="mt-4 flex flex-wrap items-center gap-2">
          {onRetry != null && (
            <Button variant="outline" onClick={onRetry}>
              Try again
            </Button>
          )}
          {action}
        </div>
      )}

      {detail != null && (
        <details className="mt-6 text-xs text-muted-foreground">
          <summary className="cursor-pointer transition-colors select-none hover:text-foreground">
            Technical detail
          </summary>
          <p className="mt-2 font-mono break-words">{detail}</p>
        </details>
      )}
    </div>
  );
};
