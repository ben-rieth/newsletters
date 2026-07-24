import { useState } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { MoreHorizontal, Download, Plus } from 'lucide-react';
import { Button } from '#/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '#/components/ui/dropdown-menu';
import { H2 } from '#/components/ui/typography';
import { ViewToggle } from '#/components/ViewToggle';
import type { View } from '#/components/ViewToggle';
import { CreateNewsletterDialog } from '#/features/newsletters/components/CreateNewsletterDialog';
import { NewslettersTable } from '#/features/newsletters/components/NewslettersTable';
import { NewslettersCards } from '#/features/newsletters/components/NewslettersCards';
import { newslettersOptions } from '#/features/newsletters/queries/newsletters';
import useExportNewsletters from '#/features/newsletters/queries/hooks/useExportNewsletters';

const NewslettersPage = () => {
  const { data: newsletters } = useSuspenseQuery(newslettersOptions);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [view, setView] = useState<View>('cards');
  const exportAll = useExportNewsletters();

  return (
    <div className="mx-auto max-w-5xl space-y-6 px-6 py-8 lg:px-10">
      <div className="flex items-center justify-between">
        <H2>Newsletters</H2>
        <div className="flex items-center gap-2">
          <div className="hidden sm:flex">
            <ViewToggle value={view} onChange={setView} />
          </div>
          <Button
            className="hidden sm:flex"
            onClick={() => setDialogOpen(true)}
          >
            New Newsletter
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger
              render={<Button variant="outline" size="icon" />}
            >
              <MoreHorizontal className="h-4 w-4" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-40">
              <DropdownMenuItem
                className="sm:hidden"
                onClick={() => setDialogOpen(true)}
              >
                <Plus className="mr-2 h-4 w-4" />
                New Newsletter
              </DropdownMenuItem>
              <DropdownMenuSeparator className="sm:hidden" />
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
