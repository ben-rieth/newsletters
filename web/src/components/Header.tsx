import { Link } from '@tanstack/react-router';
import { Menu } from 'lucide-react';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';
import { H1 } from './ui/typography';
import { Button } from './ui/button';
import { Sheet, SheetContent, SheetTrigger, SheetClose } from './ui/sheet';

const navLinkClass =
  'text-muted-foreground hover:text-foreground transition-colors';

const NavLinks = ({ isSignedIn }: { isSignedIn: boolean }) => (
  <>
    {isSignedIn && (
      <>
        <Link to="/newsletters" className={navLinkClass}>
          Newsletters
        </Link>
        <Link to="/issues" className={navLinkClass}>
          Issues
        </Link>
        <Link to="/profile" className={navLinkClass}>
          Profile
        </Link>
      </>
    )}
    <Link to="/about" className={navLinkClass}>
      About
    </Link>
  </>
);

const Header = () => {
  const isSignedIn = useIsSignedIn();

  return (
    <header className="px-4 py-3 shadow-sm bg-card flex items-center justify-between">
      <H1 className="text-2xl">
        <Link to={isSignedIn ? '/newsletters' : '/'}>Custom Newsletters</Link>
      </H1>

      {/* Desktop nav */}
      <nav className="hidden sm:flex items-center gap-4 text-sm">
        <NavLinks isSignedIn={isSignedIn} />
      </nav>

      {/* Mobile hamburger */}
      <Sheet>
        <SheetTrigger
          render={<Button variant="ghost" size="icon" className="sm:hidden" />}
          aria-label="Open menu"
        >
          <Menu />
        </SheetTrigger>
        <SheetContent side="right">
          <nav className="flex flex-col gap-6 p-6 pt-12 text-base">
            <SheetClose
              render={<Link to={isSignedIn ? '/newsletters' : '/'} />}
              className={navLinkClass}
            >
              Home
            </SheetClose>
            {isSignedIn && (
              <>
                <SheetClose
                  render={<Link to="/newsletters" />}
                  className={navLinkClass}
                >
                  Newsletters
                </SheetClose>
                <SheetClose
                  render={<Link to="/issues" />}
                  className={navLinkClass}
                >
                  Issues
                </SheetClose>
                <SheetClose
                  render={<Link to="/profile" />}
                  className={navLinkClass}
                >
                  Profile
                </SheetClose>
              </>
            )}
            <SheetClose render={<Link to="/about" />} className={navLinkClass}>
              About
            </SheetClose>
          </nav>
        </SheetContent>
      </Sheet>
    </header>
  );
};

export default Header;
