import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useSuspenseQuery, useQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import type { Newsletter } from '../queries/newsletters';
import { feedsOptions } from '../queries/feeds';
import { formatSchedule } from '../lib/format';
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
import IssuesList from '#/features/issues/components/IssuesList';
import { issuesOptions } from '#/features/issues/queries/issues';
import { Button } from '#/components/ui/button';
import { Badge } from '#/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
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

  const { data: feeds } = useSuspenseQuery(feedsOptions(newsletter.id));
  const { data: issuesData } = useQuery(issuesOptions);
  const newsletterIssues = (issuesData ?? []).filter(
    (i) => i.newsletterName === newsletter.name,
  );

  const isActive = newsletter.status === 'active';

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
      isActive ? 'Newsletter deactivated.' : 'Newsletter activated!',
    );
  });

  const cancelOneOff = useCancelOneOffSend(() => {
    toast.success('One-off send cancelled.');
  });

  return (
    <div className="space-y-6">
      <header className="space-y-2">
        <div className="flex items-center gap-3">
          <h2 className="text-2xl font-semibold tracking-tight">
            {newsletter.name}
          </h2>
          <Badge
            variant="outline"
            className={
              isActive
                ? 'border-primary/30 bg-primary/10 text-primary'
                : 'text-muted-foreground'
            }
          >
            {isActive ? 'Active' : 'Paused'}
          </Badge>
        </div>
        <p className="text-sm text-muted-foreground">
          {feeds.length} {feeds.length === 1 ? 'feed' : 'feeds'} ·{' '}
          {formatSchedule(newsletter)}
        </p>
      </header>

      <Tabs defaultValue="feeds">
        <TabsList variant="line" className="w-full justify-start border-b">
          <TabsTrigger value="feeds">Feeds</TabsTrigger>
          <TabsTrigger value="schedule">Schedule</TabsTrigger>
          <TabsTrigger value="history">History</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>

        <TabsContent value="feeds" className="pt-6">
          <div className="space-y-4">
            <FeedsList newsletterId={newsletter.id} />
          </div>
        </TabsContent>

        <TabsContent value="schedule" className="pt-6">
          <div className="max-w-lg space-y-6">
            <div className="rounded-lg border p-4">
              <p className="text-sm text-muted-foreground">Regular schedule</p>
              <p className="mt-1 font-medium">{formatSchedule(newsletter)}</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Next send {formatDateTime(newsletter.nextSendTime)}
              </p>
            </div>

            <div className="space-y-3">
              <h3 className="text-sm font-semibold">One-off send</h3>
              {newsletter.oneOffSendTime ? (
                <div className="space-y-2 rounded-lg border p-4">
                  <p className="text-sm">
                    Scheduled for{' '}
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
                      {cancelOneOff.isPending
                        ? 'Cancelling...'
                        : 'Cancel one-off'}
                    </Button>
                  </div>
                </div>
              ) : (
                <ScheduleSendDialog
                  newsletterId={newsletter.id}
                  maxSendTime={newsletter.nextSendTime}
                />
              )}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="history" className="pt-6">
          <IssuesList issues={newsletterIssues} />
        </TabsContent>

        <TabsContent value="settings" className="pt-6">
          <div className="max-w-sm space-y-8">
            <div className="space-y-4">
              <NewsletterForm
                defaultValues={defaultValues}
                onSubmit={async (values) => newsletterUpdate.mutate(values)}
                submitLabel="Save Changes"
              />
            </div>

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
                    status: isActive ? 'inactive' : 'active',
                  })
                }
                disabled={updateStatus.isPending}
              >
                {updateStatus.isPending
                  ? '...'
                  : isActive
                    ? 'Deactivate'
                    : 'Activate'}
              </Button>
            </div>

            {import.meta.env.DEV && (
              <NewsletterDebugActions newsletterId={newsletter.id} />
            )}

            <section className="space-y-3 rounded-lg border border-destructive/30 bg-destructive/5 p-4">
              <div>
                <h3 className="text-sm font-semibold text-destructive">
                  Danger Zone
                </h3>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  Permanently deletes this newsletter and all of its feeds.
                </p>
              </div>
              <Button
                variant="destructive"
                onClick={() => setDeleteDialogOpen(true)}
              >
                Delete Newsletter
              </Button>
            </section>
          </div>
        </TabsContent>
      </Tabs>

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
