import { Bug } from 'lucide-react';
import useRunScheduler from '../queries/hooks/useRunScheduler';
import { Button } from '#/components/ui/button';

const DebugPanel = () => {
  const { mutate: runScheduler, isPending } = useRunScheduler();

  return (
    <div className="fixed bottom-4 right-4 z-50 rounded-lg border bg-card p-3 shadow-lg">
      <p className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
        <Bug className="size-3.5" />
        Debug
      </p>
      <Button
        size="sm"
        variant="outline"
        onClick={() => runScheduler()}
        disabled={isPending}
      >
        {isPending ? 'Running…' : 'Run Scheduler'}
      </Button>
    </div>
  );
};

export default DebugPanel;
