import { Bug } from 'lucide-react';
import { toast } from 'sonner';
import useForceSendNewsletter from '../queries/hooks/useForceSendNewsletter';
import { Button } from '#/components/ui/button';

type Props = {
  newsletterId: string;
};

/**
 * Dev-only debugging actions for a single newsletter. Render this behind an
 * `import.meta.env.DEV` guard — the backing endpoints only exist in dev.
 */
export const NewsletterDebugActions = ({ newsletterId }: Props) => {
  const forceSend = useForceSendNewsletter(() => {
    toast.success('Newsletter queued to send!');
  });

  return (
    <div className="rounded-md border border-dashed p-4 space-y-3">
      <p className="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
        <Bug className="size-3.5" />
        Debug (dev only)
      </p>
      <Button
        variant="outline"
        onClick={() => forceSend.mutate(newsletterId)}
        disabled={forceSend.isPending}
      >
        {forceSend.isPending ? 'Sending...' : 'Send Now'}
      </Button>
    </div>
  );
};
