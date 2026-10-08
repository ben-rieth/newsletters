import { useEffect, useMemo, useState } from 'react';
import { createFileRoute, Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { ListChecks, Plus, X } from 'lucide-react';
import { cn } from '#/lib/utils';
import { Button } from '#/components/ui/button';
import { Checkbox } from '#/components/ui/checkbox';
import { ListPanel, listRowClass } from '#/components/ListPanel';
import { EmptyState } from '#/components/EmptyState';
import { useMobileHeader, MobileHeaderAction } from '#/components/MobileHeader';
import { newslettersOptions } from '#/features/newsletters/queries/newsletters';
import { formatSchedule } from '#/features/newsletters/lib/format';
import { CreateNewsletterDialog } from '#/features/newsletters/components/CreateNewsletterDialog';
import { NewsletterBulkBar } from '#/features/newsletters/components/NewsletterBulkBar';
import type { Newsletter } from '#/features/newsletters/queries/newsletters';

const NewsletterSummary = ({ newsletter }: { newsletter: Newsletter }) => (
  <div className="min-w-0">
    <p className="truncate font-medium group-hover:underline">
      {newsletter.name}
    </p>
    <p className="mt-0.5 truncate text-sm text-muted-foreground md:text-xs">
      {newsletter.status === 'active' ? formatSchedule(newsletter) : 'Paused'}
    </p>
  </div>
);

const NewslettersPage = () => {
  const { data } = useSuspenseQuery(newslettersOptions);
  const [createOpen, setCreateOpen] = useState(false);
  const [selecting, setSelecting] = useState(false);
  const [selectedIds, setSelectedIds] = useState<ReadonlySet<string>>(
    () => new Set(),
  );

  useMobileHeader({ title: 'Newsletters' });

  const sorted = useMemo(
    () => [...(data ?? [])].sort((a, b) => a.name.localeCompare(b.name)),
    [data],
  );

  const selected = sorted.filter((newsletter) =>
    selectedIds.has(newsletter.id),
  );

  const exitSelecting = () => {
    setSelecting(false);
    setSelectedIds(new Set());
  };

  const toggleSelected = (id: string, isSelected: boolean) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (isSelected) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  };

  const toggleSelecting = () =>
    selecting ? exitSelecting() : setSelecting(true);

  useEffect(() => {
    if (!selecting) {
      return;
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        exitSelecting();
      }
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [selecting]);

  return (
    <div className="mx-auto max-w-3xl px-4 py-6 md:px-6 md:py-8 lg:px-10">
      <MobileHeaderAction>
        {sorted.length > 0 && (
          <Button
            variant="ghost"
            size="icon"
            className="size-11"
            aria-label={selecting ? 'Done selecting' : 'Select newsletters'}
            onClick={toggleSelecting}
          >
            {selecting ? (
              <X className="size-5" />
            ) : (
              <ListChecks className="size-5" />
            )}
          </Button>
        )}
        <Button
          variant="ghost"
          size="icon"
          className="size-11"
          aria-label="New newsletter"
          onClick={() => setCreateOpen(true)}
        >
          <Plus className="size-5" />
        </Button>
      </MobileHeaderAction>

      <header className="mb-6 hidden items-center justify-between gap-4 md:flex">
        <h1 className="font-serif text-3xl font-medium tracking-tight">
          Newsletters
        </h1>
        <div className="flex items-center gap-2">
          {sorted.length > 0 && (
            <Button variant="ghost" onClick={toggleSelecting}>
              {selecting ? 'Done' : 'Select'}
            </Button>
          )}
          <Button onClick={() => setCreateOpen(true)}>
            <Plus data-icon="inline-start" />
            New newsletter
          </Button>
        </div>
      </header>

      {sorted.length === 0 ? (
        <EmptyState
          className="px-0 md:px-0"
          title="No newsletters yet"
          description="A newsletter is a group of feeds delivered on a schedule you pick — daily, weekly, or a specific day. Most people keep two or three, split by topic."
          action={
            <Button onClick={() => setCreateOpen(true)}>
              Create your first newsletter
            </Button>
          }
        />
      ) : (
        <>
          {selecting && (
            <NewsletterBulkBar
              newsletters={sorted}
              selected={selected}
              onSelectAll={(selectAll) =>
                setSelectedIds(
                  new Set(selectAll ? sorted.map((n) => n.id) : []),
                )
              }
              onActionComplete={(action) =>
                action === 'delete'
                  ? exitSelecting()
                  : setSelectedIds(new Set())
              }
            />
          )}
          <ListPanel className={cn(selecting && 'border-t-0')}>
            {sorted.map((newsletter) =>
              selecting ? (
                <label
                  key={newsletter.id}
                  className={cn(
                    'min-h-16 cursor-pointer',
                    listRowClass,
                    'justify-start',
                  )}
                >
                  <Checkbox
                    checked={selectedIds.has(newsletter.id)}
                    onCheckedChange={(checked) =>
                      toggleSelected(newsletter.id, checked)
                    }
                  />
                  <NewsletterSummary newsletter={newsletter} />
                </label>
              ) : (
                <Link
                  key={newsletter.id}
                  to="/newsletters/$newsletterId"
                  params={{ newsletterId: newsletter.id }}
                  className={cn('group min-h-16', listRowClass)}
                >
                  <NewsletterSummary newsletter={newsletter} />
                </Link>
              ),
            )}
          </ListPanel>
        </>
      )}

      <CreateNewsletterDialog open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  );
};

export const Route = createFileRoute('/newsletters/')({
  component: NewslettersPage,
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(newslettersOptions);
  },
});
