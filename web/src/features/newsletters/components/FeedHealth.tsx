import { CircleAlert, PauseCircle, TriangleAlert } from 'lucide-react';
import { cn } from '#/lib/utils';
import { Badge } from '#/components/ui/badge';
import { formatRelativeTime } from '#/utils/format';
import type { FeedHealth } from '../queries/feeds';

const describeLastFailure = (health: FeedHealth) => {
  if (!health.lastFailureAt) {
    return null;
  }

  const message = health.lastFailureMessage || 'Could not be retrieved';
  return `${message} (${formatRelativeTime(health.lastFailureAt)})`;
};

export const FeedHealthBadge = ({ health }: { health: FeedHealth }) => {
  if (health.status === 'ok') {
    return null;
  }

  const paused = health.status === 'disabled';

  return (
    <Badge
      variant={paused ? 'destructive' : 'outline'}
      title={describeLastFailure(health) ?? undefined}
    >
      {paused ? <PauseCircle /> : <CircleAlert />}
      {paused ? 'Paused' : 'Failing'}
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
          : 'border-yellow-500/40 bg-yellow-500/10 text-yellow-700 dark:text-yellow-400',
      )}
    >
      <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0" />
      <div className="space-y-1">
        <p className="font-medium">
          {paused
            ? 'We stopped fetching this feed'
            : 'This feed failed to update'}
        </p>
        <ul className="space-y-0.5 text-xs">
          {lastFailure && <li>Last error: {lastFailure}</li>}
          <li>
            Last successful update {formatRelativeTime(health.lastSuccessAt)}
          </li>
          {paused && health.disabledUntil && (
            <li>
              Fetching resumes {formatRelativeTime(health.disabledUntil)}. Fix
              the URL or remove the feed if it is gone for good.
            </li>
          )}
        </ul>
      </div>
    </div>
  );
};
