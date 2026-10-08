import { Ban, CircleAlert, TriangleAlert } from 'lucide-react';
import { cn } from '#/lib/utils';
import { Badge } from '#/components/ui/badge';
import { formatRelativeTime } from '#/utils/format';
import type { FeedHealth } from '../queries/feeds';

const warningClasses = 'border-warning/40 bg-warning/10 text-warning';

const describeLastFailure = (health: FeedHealth) => {
  if (!health.lastFailureAt || !health.lastFailureMessage) {
    return null;
  }

  return `${health.lastFailureMessage} (${formatRelativeTime(health.lastFailureAt)})`;
};

export const FeedHealthBadge = ({ health }: { health: FeedHealth }) => {
  if (health.status === 'ok') {
    return null;
  }

  const stopped = health.status === 'disabled';
  const lastFailure = describeLastFailure(health);

  return (
    <Badge
      variant={stopped ? 'destructive' : 'outline'}
      className={cn(!stopped && warningClasses)}
      title={lastFailure ?? undefined}
    >
      {stopped ? <Ban /> : <CircleAlert />}
      {stopped ? 'Fetching stopped' : 'Failing'}
      {lastFailure && <span className="sr-only">: {lastFailure}</span>}
    </Badge>
  );
};

export const FeedHealthAlert = ({ health }: { health: FeedHealth }) => {
  if (health.status === 'ok') {
    return null;
  }

  const stopped = health.status === 'disabled';
  const lastFailure = describeLastFailure(health);

  return (
    <div
      className={cn(
        'flex gap-2.5 rounded-md border p-3 text-sm',
        stopped
          ? 'border-destructive/40 bg-destructive/10 text-destructive'
          : warningClasses,
      )}
    >
      <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0" />
      <div className="space-y-1">
        <p className="font-medium">
          {stopped
            ? 'We stopped fetching this feed'
            : 'This feed failed to update'}
        </p>
        <ul className="list-disc space-y-0.5 pl-4 text-xs">
          {lastFailure && <li>Last error: {lastFailure}</li>}
          <li>
            Last successful update {formatRelativeTime(health.lastSuccessAt)}
          </li>
          {stopped && health.disabledUntil && (
            <li>
              Fetching resumes {formatRelativeTime(health.disabledUntil)}.
              Remove this feed and re-add it with a working URL, or delete it if
              it is gone for good.
            </li>
          )}
        </ul>
      </div>
    </div>
  );
};
