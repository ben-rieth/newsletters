import { useMemo, useState } from 'react';
import { createFileRoute, Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { ChevronRight, Plus } from 'lucide-react';
import { cn } from '#/lib/utils';
import { Button } from '#/components/ui/button';
import { ListPanel, listRowClass } from '#/components/ListPanel';
import { EmptyState } from '#/components/EmptyState';
import { useMobileHeader, MobileHeaderAction } from '#/components/MobileHeader';
import { newslettersOptions } from '#/features/newsletters/queries/newsletters';
import { formatSchedule } from '#/features/newsletters/lib/format';
import { CreateNewsletterDialog } from '#/features/newsletters/components/CreateNewsletterDialog';

const NewslettersPage = () => {
  const { data } = useSuspenseQuery(newslettersOptions);
  const [createOpen, setCreateOpen] = useState(false);

  useMobileHeader({ title: 'Newsletters' });

  const sorted = useMemo(
    () => [...(data ?? [])].sort((a, b) => a.name.localeCompare(b.name)),
    [data],
  );

  return (
    <div className="mx-auto max-w-3xl px-4 py-6 md:px-6 md:py-8 lg:px-10">
      <MobileHeaderAction>
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
        <Button onClick={() => setCreateOpen(true)}>
          <Plus data-icon="inline-start" />
          New newsletter
        </Button>
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
        <ListPanel>
          {sorted.map((newsletter) => {
            const isActive = newsletter.status === 'active';
            return (
              <Link
                key={newsletter.id}
                to="/newsletters/$newsletterId"
                params={{ newsletterId: newsletter.id }}
                className={cn('group min-h-16', listRowClass)}
              >
                <div className="min-w-0">
                  <p className="flex items-center gap-2">
                    <span
                      className={cn(
                        'size-1.5 shrink-0 rounded-full',
                        isActive ? 'bg-primary' : 'bg-muted-foreground/40',
                      )}
                      aria-hidden="true"
                    />
                    <span className="truncate font-medium group-hover:underline">
                      {newsletter.name}
                    </span>
                  </p>
                  <p className="mt-0.5 truncate text-sm text-muted-foreground md:text-xs">
                    {isActive ? formatSchedule(newsletter) : 'Paused'}
                  </p>
                </div>
                <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
              </Link>
            );
          })}
        </ListPanel>
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
