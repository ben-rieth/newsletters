import { PauseCircle } from 'lucide-react';
import { Badge } from '#/components/ui/badge';
import type { Feed } from '../queries/feeds';

export const FeedStatusBadge = ({ status }: { status: Feed['status'] }) => {
  if (status === 'active') {
    return null;
  }

  return (
    <Badge variant="outline" className="text-muted-foreground">
      <PauseCircle />
      Paused
    </Badge>
  );
};
