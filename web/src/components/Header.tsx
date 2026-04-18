import { Link } from '@tanstack/react-router';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';
import { H1 } from './ui/typography';

const Header = () => {
  const isSignedIn = useIsSignedIn();

  return (
    <header className="px-4 py-3 shadow-sm bg-card flex items-center justify-between">
      <H1 className="text-2xl">
        <Link to={isSignedIn ? '/newsletters' : '/'}>Custom Newsletters</Link>
      </H1>
      {isSignedIn && (
        <nav className="flex items-center gap-4 text-sm">
          <Link to="/newsletters" className="text-muted-foreground hover:text-foreground transition-colors">
            Newsletters
          </Link>
          <Link to="/profile" className="text-muted-foreground hover:text-foreground transition-colors">
            Profile
          </Link>
        </nav>
      )}
    </header>
  );
};

export default Header;