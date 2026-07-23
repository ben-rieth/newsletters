import { Link } from '@tanstack/react-router';
import { Card, CardContent, CardHeader, CardTitle } from '#/components/ui/card';
import { Badge } from '#/components/ui/badge';
import { Button } from '#/components/ui/button';
import { formatRelativeTime } from '#/utils/format';
import { formatSchedule } from '../lib/format';
import type { Newsletter } from '../queries/newsletters';

type Props = {
  newsletters: Newsletter[];
  onCreateClick?: () => void;
};

export const NewslettersCards = ({ newsletters, onCreateClick }: Props) => {
  if (newsletters.length === 0) {
    return (
      <div className="py-12 text-center space-y-3 rounded-md border border-dashed">
        <p className="text-sm text-muted-foreground">No newsletters yet.</p>
        {onCreateClick && (
          <Button size="sm" onClick={onCreateClick}>
            Create your first newsletter
          </Button>
        )}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {newsletters.map((n) => (
        <Link
          key={n.id}
          to="/newsletters/$newsletterId"
          params={{ newsletterId: n.id }}
        >
          <Card className="h-full hover:bg-muted/50 transition-colors cursor-pointer">
            <CardHeader className="pb-2">
              <div className="flex items-start justify-between gap-2">
                <CardTitle className="text-base font-semibold leading-snug">
                  {n.name}
                </CardTitle>
                <div className="flex items-center gap-1.5 shrink-0">
                  {n.status === 'inactive' && (
                    <Badge variant="secondary">Inactive</Badge>
                  )}
                  <span className="text-xs capitalize rounded-full bg-primary/10 text-primary px-2 py-0.5 font-medium">
                    {n.frequency}
                  </span>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-1 text-sm text-muted-foreground">
              <p>{formatSchedule(n)}</p>
              <p className="flex items-center gap-1.5">
                Next: {formatRelativeTime(n.nextSendTime)}
                {n.oneOffSendTime && <Badge variant="secondary">one-off</Badge>}
              </p>
              <p>
                Last sent:{' '}
                {n.lastSentAt ? formatRelativeTime(n.lastSentAt) : 'Never'}
              </p>
            </CardContent>
          </Card>
        </Link>
      ))}
    </div>
  );
};
