import { useMemo, useState } from 'react';
import { Link } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { Menu, Plus, Rss } from 'lucide-react';
import { cn } from '#/lib/utils';
import { Button } from '#/components/ui/button';
import {
  Sheet,
  SheetContent,
  SheetTitle,
  SheetTrigger,
} from '#/components/ui/sheet';
import { newslettersOptions } from '#/features/newsletters/queries/newsletters';
import { CreateNewsletterDialog } from '#/features/newsletters/components/CreateNewsletterDialog';
import useLogout from '#/features/auth/queries/hooks/useLogout';

const sectionLabel =
  'px-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground';

const navRow =
  'flex items-center gap-2.5 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground';

const SidebarContent = ({ onNavigate }: { onNavigate?: () => void }) => {
  const { data } = useQuery(newslettersOptions);
  const newsletters = data ?? [];
  const [createOpen, setCreateOpen] = useState(false);
  const logout = useLogout();

  const sorted = useMemo(
    () => [...newsletters].sort((a, b) => a.name.localeCompare(b.name)),
    [newsletters],
  );

  const handleLogout = () => {
    logout.mutate(undefined, {
      onSuccess: () => {
        onNavigate?.();
      },
    });
  };

  return (
    <div className="flex h-full flex-col">
      <Link
        to="/issues"
        onClick={onNavigate}
        className="flex items-center gap-2.5 px-5 py-5"
      >
        <span className="flex size-7 items-center justify-center rounded-md bg-primary text-primary-foreground">
          <Rss className="size-4" />
        </span>
        <span className="text-base font-semibold">Slowfeed</span>
      </Link>

      <div className="flex-1 overflow-y-auto px-2">
        <p className={cn(sectionLabel, 'mt-2 mb-2')}>Read</p>
        <nav aria-label="Reading" className="space-y-0.5">
          <Link
            to="/issues"
            onClick={onNavigate}
            className={navRow}
            activeProps={{
              className: cn(navRow, 'bg-accent text-foreground'),
            }}
          >
            Issues
          </Link>
        </nav>

        <p className={cn(sectionLabel, 'mt-4 mb-2')}>Newsletters</p>
        <nav aria-label="Newsletters" className="space-y-0.5">
          {sorted.map((n) => (
            <Link
              key={n.id}
              to="/newsletters/$newsletterId"
              params={{ newsletterId: n.id }}
              onClick={onNavigate}
              activeOptions={{ exact: false }}
              className={cn(navRow, 'justify-between')}
              activeProps={{
                className: cn(
                  navRow,
                  'justify-between bg-accent text-foreground',
                ),
              }}
            >
              <span className="flex items-center gap-2.5 truncate">
                <span
                  className={cn(
                    'size-1.5 shrink-0 rounded-full',
                    n.status === 'active'
                      ? 'bg-primary'
                      : 'bg-muted-foreground/40',
                  )}
                />
                <span className="truncate">{n.name}</span>
              </span>
              {n.status !== 'active' && (
                <span className="text-[10px] uppercase tracking-wide text-muted-foreground">
                  Paused
                </span>
              )}
            </Link>
          ))}
        </nav>

        <button
          type="button"
          onClick={() => setCreateOpen(true)}
          className={cn(
            navRow,
            'mt-2 w-full border border-dashed border-border',
          )}
        >
          <Plus className="size-4" />
          New newsletter
        </button>
      </div>

      <div className="border-t border-sidebar-border">
        <nav aria-label="Account" className="space-y-0.5 px-2 py-3">
          <Link to="/profile" onClick={onNavigate} className={navRow}>
            Profile
          </Link>
          <Link to="/about" onClick={onNavigate} className={navRow}>
            About
          </Link>
          <button
            type="button"
            onClick={handleLogout}
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

const AppSidebar = () => {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <>
      <aside className="hidden w-64 shrink-0 border-r border-sidebar-border bg-sidebar md:block">
        <SidebarContent />
      </aside>

      <header className="flex items-center justify-between border-b border-sidebar-border bg-sidebar px-4 py-3 md:hidden">
        <Link to="/issues" className="flex items-center gap-2 font-semibold">
          <span className="flex size-6 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <Rss className="size-3.5" />
          </span>
          Slowfeed
        </Link>
        <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
          <SheetTrigger
            render={<Button variant="ghost" size="icon" />}
            aria-label="Open menu"
          >
            <Menu />
          </SheetTrigger>
          <SheetContent
            side="left"
            className="bg-sidebar p-0 data-[side=left]:w-72"
          >
            <SheetTitle className="sr-only">Navigation</SheetTitle>
            <SidebarContent onNavigate={() => setMobileOpen(false)} />
          </SheetContent>
        </Sheet>
      </header>
    </>
  );
};

export default AppSidebar;
