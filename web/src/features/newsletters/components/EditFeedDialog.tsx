import { toast } from 'sonner';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { EditFeedForm } from './EditFeedForm';
import useUpdateFeed from '../queries/hooks/useUpdateFeed';
import type { Feed } from '../queries/feeds';

type Props = {
  newsletterId: string;
  feed: Feed;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export const EditFeedDialog = ({
  newsletterId,
  feed,
  open,
  onOpenChange,
}: Props) => {
  const updateFeed = useUpdateFeed(newsletterId, feed.url, () => {
    toast.success('Feed updated!');
    onOpenChange(false);
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Feed</DialogTitle>
        </DialogHeader>

        <EditFeedForm
          key={open ? 'open' : 'closed'}
          feed={feed}
          onSubmit={async (alias) => updateFeed.mutate({ alias })}
          isPending={updateFeed.isPending}
          error={updateFeed.error?.message}
        />
      </DialogContent>
    </Dialog>
  );
};
