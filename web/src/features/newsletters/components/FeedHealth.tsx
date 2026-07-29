import { CircleAlert, PauseCircle, TriangleAlert } from 'lucide-react';
import { cn } from '#/lib/utils';
import { Badge } from '#/components/ui/badge';
import { formatRelativeTime } from '#/utils/format';
import type { FeedHealth } from '../queries/feeds';

const warningClasses =
  'border-yellow-500/40 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400';

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

  const paused = health.status === 'disabled';
  const lastFailure = describeLastFailure(health);

  return (
    <Badge
      variant={paused ? 'destructive' : 'outline'}
      className={cn(!paused && warningClasses)}
      title={lastFailure ?? undefined}
    >
      {paused ? <PauseCircle /> : <CircleAlert />}
      {paused ? 'Fetch paused' : 'Failing'}
      {lastFailure && <span className="sr-only">: {lastFailure}</span>}
    </Badge>
  );
};

export const FeedHealthAlert = ({ health }: { health: FeedHealth }) => {
  if (health.status === 'ok') {
    return null;
  }

  const paused = health.status === 'disabled';
  const lastFailure = describeLastFailure(health);

  return (
    <div
      className={cn(
        'flex gap-2.5 rounded-md border p-3 text-sm',
        paused
          ? 'border-destructive/40 bg-destructive/10 text-destructive'
          : warningClasses,
      )}
    >
      <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0" />
      <div className="space-y-1">
        <p className="font-medium">
          {paused
            ? 'We stopped fetching this feed'
            : 'This feed failed to update'}
        </p>
        <ul className="list-disc space-y-0.5 pl-4 text-xs">
          {lastFailure && <li>Last error: {lastFailure}</li>}
          <li>
            Last successful update {formatRelativeTime(health.lastSuccessAt)}
          </li>
          {paused && health.disabledUntil && (
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
