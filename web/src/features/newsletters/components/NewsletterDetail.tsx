import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import type { Newsletter } from '../queries/newsletters';
import { NewsletterForm } from './NewsletterForm';
import type { NewsletterFormValues } from './NewsletterForm';
import useUpdateNewsletter from '../queries/hooks/useUpdateNewsletter';
import useDeleteNewsletter from '../queries/hooks/useDeleteNewsletter';
import useExportNewsletter from '../queries/hooks/useExportNewsletter';
import useUpdateNewsletterStatus from '../queries/hooks/useUpdateNewsletterStatus';
import useCancelOneOffSend from '../queries/hooks/useCancelOneOffSend';
import { FeedsList } from './FeedsList';
import { ScheduleSendDialog } from './ScheduleSendDialog';
import { NewsletterDebugActions } from '#/features/debug/components/NewsletterDebugActions';
import { Button } from '#/components/ui/button';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '#/components/ui/alert-dialog';

type Props = {
  newsletter: Newsletter;
};

const formatDateTime = (dateStr: string) =>
  new Date(dateStr).toLocaleString('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  });

export const NewsletterDetail = ({ newsletter }: Props) => {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const navigate = useNavigate();

  const defaultValues: NewsletterFormValues = {
    name: newsletter.name,
    frequency: newsletter.frequency as NewsletterFormValues['frequency'],
    sendHour: newsletter.sendHour,
    sendMinute: newsletter.sendMinute,
    sendDay: newsletter.sendDay,
    sendTimezone: newsletter.sendTimezone,
  };

  const newsletterUpdate = useUpdateNewsletter(newsletter.id, () => {
    toast.success('Newsletter updated!');
  });

  const newsletterDelete = useDeleteNewsletter(() => {
    toast.success('Newsletter deleted!');
    navigate({ to: '/newsletters' });
  });

  const exportNewsletter = useExportNewsletter();

  const updateStatus = useUpdateNewsletterStatus(() => {
    toast.success(
      newsletter.status === 'active'
        ? 'Newsletter deactivated.'
        : 'Newsletter activated!',
    );
  });

  const cancelOneOff = useCancelOneOffSend(() => {
    toast.success('One-off send cancelled.');
  });

  return (
    <div className="space-y-8">
      <section className="space-y-4">
        <FeedsList newsletterId={newsletter.id} />
      </section>
      <section className="space-y-4">
        <h3 className="text-lg font-semibold">Settings</h3>
        <div className="max-w-sm">
          <NewsletterForm
            defaultValues={defaultValues}
            onSubmit={async (values) => newsletterUpdate.mutate(values)}
            submitLabel="Save Changes"
          />
        </div>
      </section>
      <section className="space-y-4">
        <h3 className="text-lg font-semibold">Actions</h3>
        {newsletter.oneOffSendTime ? (
          <div className="max-w-sm space-y-2 rounded-md border p-3">
            <p className="text-sm">
              <span className="font-medium">One-off send</span> scheduled for{' '}
              <span className="font-medium">
                {formatDateTime(newsletter.oneOffSendTime)}
              </span>
            </p>
            <div className="flex items-center gap-2">
              <ScheduleSendDialog
                newsletterId={newsletter.id}
                maxSendTime={
                  newsletter.regularSendTime ?? newsletter.nextSendTime
                }
                initialSendTime={newsletter.oneOffSendTime}
              />
              <Button
                variant="outline"
                onClick={() => cancelOneOff.mutate(newsletter.id)}
                disabled={cancelOneOff.isPending}
              >
                {cancelOneOff.isPending ? 'Cancelling...' : 'Cancel one-off'}
              </Button>
            </div>
          </div>
        ) : (
          <ScheduleSendDialog
            newsletterId={newsletter.id}
            maxSendTime={newsletter.nextSendTime}
          />
        )}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => exportNewsletter.mutate(newsletter.id)}
            disabled={exportNewsletter.isPending}
          >
            {exportNewsletter.isPending ? 'Exporting...' : 'Export'}
          </Button>
          <Button
            variant="outline"
            onClick={() =>
              updateStatus.mutate({
                newsletterId: newsletter.id,
                status: newsletter.status === 'active' ? 'inactive' : 'active',
              })
            }
            disabled={updateStatus.isPending}
          >
            {updateStatus.isPending
              ? '...'
              : newsletter.status === 'active'
                ? 'Deactivate'
                : 'Activate'}
          </Button>
        </div>
      </section>
      {import.meta.env.DEV && (
        <NewsletterDebugActions newsletterId={newsletter.id} />
      )}
      <section className="rounded-md border border-destructive/30 bg-destructive/5 p-4 space-y-3">
        <div>
          <h3 className="text-sm font-semibold text-destructive">
            Danger Zone
          </h3>
          <p className="text-xs text-muted-foreground mt-0.5">
            Permanently deletes this newsletter and all of its feeds.
          </p>
        </div>
        <Button variant="destructive" onClick={() => setDeleteDialogOpen(true)}>
          Delete Newsletter
        </Button>
      </section>

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete newsletter?</AlertDialogTitle>
            <AlertDialogDescription>
              &ldquo;{newsletter.name}&rdquo; and all of its feeds will be
              permanently deleted.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => newsletterDelete.mutate(newsletter.id)}
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
};
