import { toast } from 'sonner';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { AddFeedForm } from './AddFeedForm';
import useAddFeed from '../queries/hooks/useAddFeed';

type Props = {
  newsletterId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export const AddFeedDialog = ({ newsletterId, open, onOpenChange }: Props) => {
  const addFeed = useAddFeed(newsletterId, () => {
    toast.success('Feed added!');
    onOpenChange(false);
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add Feed</DialogTitle>
        </DialogHeader>

        <AddFeedForm
          key={open ? 'open' : 'closed'}
          onSubmit={async (values) => addFeed.mutate(values)}
          isPending={addFeed.isPending}
          submitLabel="Add Feed"
          error={addFeed.error?.message}
        />
      </DialogContent>
    </Dialog>
  );
};
