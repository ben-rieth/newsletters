import { PauseCircle } from 'lucide-react';
import { Badge } from '#/components/ui/badge';

export const FeedStatusBadge = ({ status }: { status: string }) => {
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
