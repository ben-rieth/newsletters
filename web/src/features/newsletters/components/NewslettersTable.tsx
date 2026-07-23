import { Link } from '@tanstack/react-router';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '#/components/ui/table';
import { Button } from '#/components/ui/button';
import { Badge } from '#/components/ui/badge';
import { formatRelativeTime, formatShortDate } from '#/utils/format';
import { formatSchedule } from '../lib/format';
import type { Newsletter } from '../queries/newsletters';

type Props = {
  newsletters: Newsletter[];
  onCreateClick?: () => void;
};

export const NewslettersTable = ({ newsletters, onCreateClick }: Props) => {
  return (
    <div className="overflow-x-auto -mx-4 px-4 sm:mx-0 sm:px-0">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Frequency</TableHead>
            <TableHead>Schedule</TableHead>
            <TableHead>Next Send</TableHead>
            <TableHead>Last Sent</TableHead>
            <TableHead>Created</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {newsletters.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="py-10 text-center">
                <p className="text-muted-foreground mb-3">
                  No newsletters yet.
                </p>
                {onCreateClick && (
                  <Button size="sm" onClick={onCreateClick}>
                    Create your first newsletter
                  </Button>
                )}
              </TableCell>
            </TableRow>
          ) : (
            newsletters.map((n) => (
              <TableRow key={n.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-2">
                    <Link
                      to="/newsletters/$newsletterId"
                      params={{ newsletterId: n.id }}
                      className="hover:underline"
                    >
                      {n.name}
                    </Link>
                    {n.status === 'inactive' && (
                      <Badge variant="secondary">Inactive</Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell className="capitalize">{n.frequency}</TableCell>
                <TableCell>{formatSchedule(n)}</TableCell>
                <TableCell>
                  <div className="flex items-center gap-1.5">
                    {formatRelativeTime(n.nextSendTime)}
                    {n.oneOffSendTime && (
                      <Badge variant="secondary">one-off</Badge>
                    )}
                  </div>
                </TableCell>
                <TableCell>
                  {n.lastSentAt ? formatRelativeTime(n.lastSentAt) : 'Never'}
                </TableCell>
                <TableCell>{formatShortDate(n.createdAt)}</TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
};
