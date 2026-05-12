import { Link } from '@tanstack/react-router';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';
import { H1 } from './ui/typography';

const navLinkClass =
  'text-muted-foreground hover:text-foreground transition-colors';

const Header = () => {
  const isSignedIn = useIsSignedIn();

  return (
    <header className="px-4 py-3 shadow-sm bg-card flex items-center justify-between">
      <H1 className="text-2xl">
        <Link to={isSignedIn ? '/newsletters' : '/'}>Custom Newsletters</Link>
      </H1>
      <nav className="flex items-center gap-4 text-sm">
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
      </nav>
    </header>
  );
};

export default Header;
