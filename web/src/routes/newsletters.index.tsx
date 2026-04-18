import { useState } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { MoreHorizontal, Download } from 'lucide-react';
import { Button } from '#/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '#/components/ui/dropdown-menu';
import { H2 } from '#/components/ui/typography';
import { ViewToggle, type View } from '#/components/ViewToggle';
import { CreateNewsletterDialog } from '#/features/newsletters/components/CreateNewsletterDialog';
import { NewslettersTable } from '#/features/newsletters/components/NewslettersTable';
import { NewslettersCards } from '#/features/newsletters/components/NewslettersCards';
import { newslettersOptions } from '#/features/newsletters/queries/newsletters';
import useExportNewsletters from '#/features/newsletters/queries/hooks/useExportNewsletters';

const NewslettersPage = () => {
  const { data: newsletters } = useSuspenseQuery(newslettersOptions);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [view, setView] = useState<View>('table');
  const exportAll = useExportNewsletters();

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <H2>Newsletters</H2>
        <div className="flex items-center gap-2">
          <ViewToggle value={view} onChange={setView} />
          <Button onClick={() => setDialogOpen(true)}>New Newsletter</Button>
          <DropdownMenu>
            <DropdownMenuTrigger
              render={<Button variant="outline" size="icon" />}
            >
              <MoreHorizontal className="h-4 w-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={() => exportAll.mutate()}
                disabled={exportAll.isPending}
              >
                <Download className="mr-2 h-4 w-4" />
                {exportAll.isPending ? 'Exporting...' : 'Export All'}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {view === 'table' ? (
        <NewslettersTable
          newsletters={newsletters ?? []}
          onCreateClick={() => setDialogOpen(true)}
        />
      ) : (
        <NewslettersCards
          newsletters={newsletters ?? []}
          onCreateClick={() => setDialogOpen(true)}
        />
      )}

      <CreateNewsletterDialog open={dialogOpen} onOpenChange={setDialogOpen} />
    </div>
  );
};

export const Route = createFileRoute('/newsletters/')({
  component: NewslettersPage,
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(newslettersOptions);
  },
});
