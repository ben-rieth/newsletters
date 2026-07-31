import { useMemo, useState } from 'react';
import { Link } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { ChevronLeft, Plus } from 'lucide-react';
import { cn } from '#/lib/utils';
import { buttonVariants } from '#/components/ui/button';
import { useMobileHeaderStore } from '#/components/MobileHeader';
import { newslettersOptions } from '#/features/newsletters/queries/newsletters';
import { CreateNewsletterDialog } from '#/features/newsletters/components/CreateNewsletterDialog';
import useLogout from '#/features/auth/queries/hooks/useLogout';

const sectionLabel =
  'px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground';

const navRow =
  'flex min-h-11 items-center gap-2.5 border-l-2 border-transparent px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent/40 hover:text-foreground md:min-h-0';

const navRowActive = 'border-primary text-foreground';

const SidebarContent = () => {
  const { data, isPending, isError, refetch } = useQuery(newslettersOptions);
  const newsletters = data ?? [];
  const [createOpen, setCreateOpen] = useState(false);
  const logout = useLogout();

  const sorted = useMemo(
    () => [...newsletters].sort((a, b) => a.name.localeCompare(b.name)),
    [newsletters],
  );

  return (
    <div className="flex h-full flex-col">
      <Link to="/issues" className="flex items-center gap-2 px-5 py-5">
        <span className="h-4 w-1 rounded-full bg-primary" />
        <span className="text-base font-semibold tracking-tight">Slowfeed</span>
      </Link>

      <div className="flex-1 overflow-y-auto px-2">
        <nav aria-label="Reading" className="mt-2">
          <Link
            to="/issues"
            className={cn(navRow, 'font-medium text-foreground')}
            activeProps={{
              className: cn(navRow, 'font-medium', navRowActive),
            }}
          >
            Issues
          </Link>
        </nav>

        <p className={cn(sectionLabel, 'mt-6 mb-2')}>Newsletters</p>
        <nav aria-label="Newsletters">
          {isPending && (
            <p className="px-3 py-2 text-sm text-muted-foreground">Loading…</p>
          )}

          {isError && (
            <div className="px-3 py-2">
              <p className="text-sm text-muted-foreground">
                Couldn’t load newsletters.
              </p>
              <button
                type="button"
                onClick={() => void refetch()}
                className="mt-1 text-sm text-foreground underline underline-offset-4"
              >
                Retry
              </button>
            </div>
          )}

          {sorted.map((n) => (
            <Link
              key={n.id}
              to="/newsletters/$newsletterId"
              params={{ newsletterId: n.id }}
              activeOptions={{ exact: false }}
              className={cn(navRow, 'justify-between')}
              activeProps={{
                className: cn(navRow, 'justify-between', navRowActive),
              }}
            >
              <span className="truncate">{n.name}</span>
              {n.status !== 'active' && (
                <span className="text-xs uppercase tracking-wide text-muted-foreground md:text-[10px]">
                  Paused
                </span>
              )}
            </Link>
          ))}
        </nav>

        <button
          type="button"
          onClick={() => setCreateOpen(true)}
          className={cn(navRow, 'mt-1 w-full')}
        >
          <Plus className="size-4" />
          New newsletter
        </button>
      </div>

      <div className="border-t border-sidebar-border pb-safe-b">
        <nav aria-label="Account" className="space-y-0.5 px-2 py-3">
          <Link to="/profile" className={navRow}>
            Profile
          </Link>
          <Link to="/about" className={navRow}>
            About
          </Link>
          <button
            type="button"
            onClick={() => logout.mutate()}
            disabled={logout.isPending}
            className={cn(navRow, 'w-full')}
          >
            {logout.isPending ? 'Logging out…' : 'Log out'}
          </button>
        </nav>
      </div>

      <CreateNewsletterDialog open={createOpen} onOpenChange={setCreateOpen} />
    </div>
  );
};

const MobileTopBar = () => {
  const { config, registerActionSlot } = useMobileHeaderStore();

  return (
    <header className="shrink-0 border-b border-sidebar-border bg-sidebar pt-safe-t md:hidden">
      <div className="flex h-14 items-center gap-1 pr-2 pl-1">
        {config.back && (
          <Link
            {...config.back}
            aria-label="Back"
            className={cn(
              buttonVariants({ variant: 'ghost', size: 'icon' }),
              'size-11',
            )}
          >
            <ChevronLeft className="size-5" />
          </Link>
        )}

        {config.title ? (
          <h1
            className={cn(
              'min-w-0 flex-1 truncate text-base font-semibold',
              !config.back && 'pl-3',
            )}
          >
            {config.title}
          </h1>
        ) : (
          <Link
            to="/issues"
            className="flex min-w-0 flex-1 items-center gap-2 pl-3 font-semibold tracking-tight"
          >
            <span className="h-4 w-1 shrink-0 rounded-full bg-primary" />
            Slowfeed
          </Link>
        )}

        <div
          ref={registerActionSlot}
          className="flex shrink-0 items-center gap-1"
        />
      </div>
    </header>
  );
};

const AppSidebar = () => (
  <>
    <aside className="hidden w-64 shrink-0 border-r border-sidebar-border bg-sidebar md:block">
      <SidebarContent />
    </aside>

    <MobileTopBar />
  </>
);

export default AppSidebar;
