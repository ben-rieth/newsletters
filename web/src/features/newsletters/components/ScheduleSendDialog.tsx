import { useState } from 'react';
import { toast } from 'sonner';
import useScheduleOneOffSend from '../queries/hooks/useScheduleOneOffSend';
import { Button } from '#/components/ui/button';
import { Label } from '#/components/ui/label';
import { DateTimePicker } from '#/components/DateTimePicker';
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

type Props = {
  newsletterId: string;
  nextSendTime: string;
};

const formatDateTime = (date: Date) =>
  date.toLocaleString('en-US', { dateStyle: 'medium', timeStyle: 'short' });

export const ScheduleSendDialog = ({ newsletterId, nextSendTime }: Props) => {
  const [open, setOpen] = useState(false);
  const [sendTime, setSendTime] = useState<Date | undefined>(undefined);

  const nextSend = new Date(nextSendTime);

  const scheduleSend = useScheduleOneOffSend(() => {
    toast.success('One-off send scheduled!');
    setOpen(false);
    setSendTime(undefined);
  });

  const handleSubmit = () => {
    if (!sendTime) {
      toast.error('Pick a date to send.');
      return;
    }
    if (sendTime.getTime() <= Date.now()) {
      toast.error('Send time must be in the future.');
      return;
    }
    if (sendTime > nextSend) {
      toast.error('Send time must be before the next scheduled send.');
      return;
    }
    scheduleSend.mutate({ newsletterId, sendTime: sendTime.toISOString() });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>Schedule Send</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Schedule a one-off send</DialogTitle>
          <DialogDescription>
            Send this newsletter once at a time before its next scheduled send (
            {formatDateTime(nextSend)}). This does not change the regular
            schedule.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-1.5">
          <Label htmlFor="one-off-send-date">Send date &amp; time</Label>
          <DateTimePicker
            id="one-off-send-date"
            value={sendTime}
            onChange={setSendTime}
            minDate={new Date()}
            maxDate={nextSend}
          />
        </div>

        {sendTime && (
          <p className="text-xs text-muted-foreground">
            Will send on{' '}
            <span className="font-medium text-foreground">
              {formatDateTime(sendTime)}
            </span>
            .
          </p>
        )}

        <DialogFooter>
          <DialogClose render={<Button variant="outline" />}>
            Cancel
          </DialogClose>
          <Button onClick={handleSubmit} disabled={scheduleSend.isPending}>
            {scheduleSend.isPending ? 'Scheduling...' : 'Schedule'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
