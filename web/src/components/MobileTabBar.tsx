import { Link } from '@tanstack/react-router';
import { Inbox, Newspaper, User } from 'lucide-react';
import type { LinkProps } from '@tanstack/react-router';
import { cn } from '#/lib/utils';

const TABS: { to: LinkProps['to']; label: string; icon: typeof Inbox }[] = [
  { to: '/issues', label: 'Issues', icon: Inbox },
  { to: '/newsletters', label: 'Newsletters', icon: Newspaper },
  { to: '/profile', label: 'Profile', icon: User },
];

const tab =
  'flex min-h-14 flex-1 flex-col items-center justify-center gap-1 text-[11px] text-muted-foreground transition-colors active:bg-accent/60';

const MobileTabBar = () => (
  <nav
    aria-label="Main"
    className="shrink-0 border-t border-sidebar-border bg-sidebar pb-safe-b md:hidden"
  >
    <div className="flex">
      {TABS.map(({ to, label, icon: Icon }) => (
        <Link
          key={to}
          to={to}
          className={tab}
          activeProps={{ className: cn(tab, 'text-foreground') }}
        >
          {({ isActive }) => (
            <>
              <Icon className={cn('size-5', isActive && 'text-primary')} />
              {label}
            </>
          )}
        </Link>
      ))}
    </div>
  </nav>
);

export default MobileTabBar;
