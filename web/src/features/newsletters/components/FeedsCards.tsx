import { Link } from '@tanstack/react-router';
import { ExternalLink, Settings, Trash2 } from 'lucide-react';
import { Button, buttonVariants } from '#/components/ui/button';
import type { Feed } from '../queries/feeds';

type Props = {
  feeds: Feed[];
  newsletterId: string;
  onDeleteClick: (feed: Feed) => void;
  onAddClick: () => void;
};

export const FeedsCards = ({ feeds, newsletterId, onDeleteClick, onAddClick }: Props) => {
  if (feeds.length === 0) {
    return (
      <div className="py-6 text-center space-y-3 rounded-md border border-dashed">
        <p className="text-sm text-muted-foreground">No feeds added yet.</p>
        <Button size="sm" onClick={onAddClick}>
          Add your first feed
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {feeds.map((feed) => (
        <div
          key={feed.id}
          className="flex items-center justify-between rounded-md border px-4 py-3"
        >
          <div className="space-y-0.5">
            <Link
              to="/newsletters/$newsletterId/feeds/$feedId"
              params={{ newsletterId, feedId: feed.id }}
              className="text-sm font-medium hover:underline"
            >
              {feed.alias || feed.title}
            </Link>
            {feed.description && (
              <p className="text-xs text-muted-foreground">{feed.description}</p>
            )}
          </div>
          <div className="flex items-center gap-1 ml-4 shrink-0">
            <Button variant="ghost" size="icon" aria-label="Visit feed website">
              <a href={feed.url} target="_blank" rel="noopener noreferrer">
                <ExternalLink className="size-4" />
              </a>
            </Button>
            <Link
              to="/newsletters/$newsletterId/feeds/$feedId"
              params={{ newsletterId, feedId: feed.id }}
              aria-label="Configure feed"
              className={buttonVariants({ variant: 'ghost', size: 'icon' })}
            >
              <Settings className="size-4" />
            </Link>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => onDeleteClick(feed)}
              aria-label="Delete feed"
            >
              <Trash2 className="size-4" />
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
};
