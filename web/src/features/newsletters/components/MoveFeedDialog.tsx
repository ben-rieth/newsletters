import { useState } from 'react';
import { Button } from '#/components/ui/button';
import { Label } from '#/components/ui/label';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '#/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import type { Newsletter } from '../queries/newsletters';

type Props = {
  feedName: string;
  newsletters: Newsletter[];
  isPending: boolean;
  onMove: (targetNewsletterId: string) => void;
};

export const MoveFeedDialog = ({
  feedName,
  newsletters,
  isPending,
  onMove,
}: Props) => {
  const [open, setOpen] = useState(false);
  const [targetId, setTargetId] = useState<string | null>(null);

  const options = newsletters.map((n) => ({ value: n.id, label: n.name }));

  const handleOpenChange = (next: boolean) => {
    setOpen(next);
    if (!next) {
      setTargetId(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={<Button variant="outline" />}>
        Move feed…
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Move feed</DialogTitle>
          <DialogDescription>
            Move &ldquo;{feedName}&rdquo; to another newsletter. Its alias,
            filters and status move with it.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-2">
          <Label htmlFor="move-feed-target">Newsletter</Label>
          <Select value={targetId} onValueChange={setTargetId} items={options}>
            <SelectTrigger id="move-feed-target" className="w-full">
              <SelectValue placeholder="Select newsletter" />
            </SelectTrigger>
            <SelectContent>
              {options.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <DialogFooter>
          <DialogClose render={<Button variant="outline" />}>
            Cancel
          </DialogClose>
          <Button
            onClick={() => targetId && onMove(targetId)}
            disabled={!targetId || isPending}
          >
            {isPending ? 'Moving…' : 'Move'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
