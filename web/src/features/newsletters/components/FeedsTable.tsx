import { Link } from '@tanstack/react-router';
import { ExternalLink, Settings, Trash2 } from 'lucide-react';
import { Button, buttonVariants } from '#/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '#/components/ui/table';
import type { Feed } from '../queries/feeds';

type Props = {
  feeds: Feed[];
  newsletterId: string;
  onDeleteClick: (feed: Feed) => void;
  onAddClick: () => void;
};

export const FeedsTable = ({
  feeds,
  newsletterId,
  onDeleteClick,
  onAddClick,
}: Props) => {
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
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Feed
          </TableHead>
          <TableHead className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            Description
          </TableHead>
          <TableHead className="w-px" />
        </TableRow>
      </TableHeader>
      <TableBody>
        {feeds.map((feed) => (
          <TableRow key={feed.id} className="group">
            <TableCell className="py-3.5">
              <Link
                to="/newsletters/$newsletterId/feeds/$feedId"
                params={{ newsletterId, feedId: feed.id }}
                className="block"
              >
                <span className="font-medium group-hover:underline">
                  {feed.alias || feed.title}
                </span>
                <span className="mt-0.5 block truncate font-mono text-xs text-muted-foreground">
                  {feed.url}
                </span>
              </Link>
            </TableCell>
            <TableCell className="text-muted-foreground">
              {feed.description || '—'}
            </TableCell>
            <TableCell>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label="Visit feed website"
                >
                  <a
                    href={feed.htmlUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <ExternalLink className="size-3.5" />
                  </a>
                </Button>
                <Link
                  to="/newsletters/$newsletterId/feeds/$feedId"
                  params={{ newsletterId, feedId: feed.id }}
                  aria-label="Configure feed"
                  className={buttonVariants({
                    variant: 'ghost',
                    size: 'icon-sm',
                  })}
                >
                  <Settings className="size-3.5" />
                </Link>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => onDeleteClick(feed)}
                  aria-label="Delete feed"
                >
                  <Trash2 className="size-3.5" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
};
