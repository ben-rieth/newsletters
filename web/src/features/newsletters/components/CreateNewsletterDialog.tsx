import { toast } from 'sonner';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { NewsletterForm } from './NewsletterForm';
import type { NewsletterFormValues } from './NewsletterForm';
import useCreateNewsletter from '../queries/hooks/useCreateNewsletter';

const DEFAULT_VALUES: NewsletterFormValues = {
  name: '',
  frequency: 'daily',
  sendHour: 9,
  sendMinute: 0,
  sendDay: undefined,
  sendTimezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
};

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export const CreateNewsletterDialog = ({ open, onOpenChange }: Props) => {
  const newsletterCreate = useCreateNewsletter(() => {
    toast.success('Newsletter created!');
    onOpenChange(false);
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New Newsletter</DialogTitle>
        </DialogHeader>

        <NewsletterForm
          key={open ? 'open' : 'closed'}
          defaultValues={DEFAULT_VALUES}
          onSubmit={async (values) => newsletterCreate.mutate(values)}
          isPending={newsletterCreate.isPending}
          submitLabel="Create"
          error={newsletterCreate.error?.message}
        />
      </DialogContent>
    </Dialog>
  );
};
