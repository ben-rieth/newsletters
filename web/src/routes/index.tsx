import { useState } from 'react';
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { Button, buttonVariants } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';

const STEPS = [
  {
    n: '01',
    title: 'Group the feeds',
    body: 'Make a newsletter — Entertainment, Morning News, whatever — and paste in RSS or Atom URLs.',
  },
  {
    n: '02',
    title: 'Pick when it lands',
    body: 'Daily, weekly, or a specific day, at the time and timezone you choose.',
  },
  {
    n: '03',
    title: 'Read it once',
    body: 'The issue arrives as email and stays in the app. There’s nothing to keep up with between sends.',
  },
];

const FEED_EXAMPLES = [
  {
    label: 'YouTube channel',
    url: 'youtube.com/feeds/videos.xml?channel_id=…',
  },
  { label: 'News site', url: 'reuters.com/world/rss' },
  { label: 'Blog or Substack', url: 'yourfavourite.substack.com/feed' },
  { label: 'Podcast', url: 'feeds.simplecast.com/…' },
  { label: 'Software releases', url: 'github.com/owner/repo/releases.atom' },
];

const LandingPage = () => {
  const isSignedIn = useIsSignedIn();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');

  const startReading = () => {
    navigate({
      to: '/sign-up',
      search: email.trim() ? { email: email.trim() } : {},
    });
  };

  return (
    <div className="min-h-full bg-background text-foreground">
      <header className="mx-auto flex h-16 max-w-6xl items-center justify-between px-5 sm:px-6 lg:px-10">
        <div className="flex items-center gap-2">
          <span className="h-4 w-1 rounded-full bg-primary" />
          <span className="text-base font-semibold tracking-tight">
            Slowfeed
          </span>
        </div>

        {isSignedIn ? (
          <Link to="/issues" className={buttonVariants()}>
            Go to Issues
          </Link>
        ) : (
          <Link to="/sign-in" className={buttonVariants({ variant: 'ghost' })}>
            Sign in
          </Link>
        )}
      </header>

      <main>
        <section className="mx-auto max-w-6xl px-5 py-20 sm:px-6 sm:py-28 lg:px-10 lg:py-36">
          <h1 className="max-w-4xl font-serif text-display font-medium text-balance">
            Read on your schedule, not the algorithm’s.
          </h1>
          <p className="mt-8 max-w-xl text-lg leading-relaxed text-muted-foreground">
            Slowfeed turns a list of RSS feeds into your own newsletter. Paste
            the URLs you actually follow, pick when it arrives, read the lot in
            one sitting.
          </p>

          {!isSignedIn && (
            <form
              className="mt-10 flex max-w-md flex-col gap-3 sm:flex-row"
              onSubmit={(e) => {
                e.preventDefault();
                startReading();
              }}
            >
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                aria-label="Email address"
                className="h-11 flex-1"
              />
              <Button type="submit" size="lg">
                Start reading
              </Button>
            </form>
          )}
        </section>

        <section className="border-t border-border">
          <div className="mx-auto max-w-6xl px-5 py-14 sm:px-6 sm:py-20 lg:px-10">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-baseline sm:gap-6">
              <h2 className="font-serif text-3xl font-medium tracking-tight sm:text-4xl">
                Three steps, then nothing
              </h2>
              <p className="text-muted-foreground">
                Setup takes about two minutes per newsletter.
              </p>
            </div>

            <div className="mt-12 grid gap-10 sm:grid-cols-3">
              {STEPS.map((step) => (
                <div key={step.n} className="border-t border-border pt-4">
                  <span className="font-serif text-xl tabular-nums text-muted-foreground">
                    {step.n}
                  </span>
                  <h3 className="mt-3 text-lg font-semibold">{step.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                    {step.body}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="border-t border-border">
          <div className="mx-auto grid max-w-6xl gap-10 px-5 py-14 sm:px-6 sm:py-20 lg:grid-cols-2 lg:gap-16 lg:px-10">
            <div>
              <h2 className="font-serif text-3xl font-medium tracking-tight sm:text-4xl">
                If it has a feed, it works
              </h2>
              <p className="mt-5 max-w-md leading-relaxed text-muted-foreground">
                Slowfeed takes URLs — that’s the whole input. Most sites still
                publish a feed even when they don’t advertise one. Add filters
                to drop the items you don’t want, so a noisy source stays in its
                lane.
              </p>
            </div>

            <div className="divide-y divide-border border-y border-border">
              {FEED_EXAMPLES.map((example) => (
                <div
                  key={example.label}
                  className="grid gap-1 py-4 sm:grid-cols-[9rem_minmax(0,1fr)] sm:items-center sm:gap-4"
                >
                  <span className="text-sm font-medium">{example.label}</span>
                  <span className="truncate font-mono text-sm text-muted-foreground">
                    {example.url}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t border-border">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-5 py-8 text-sm text-muted-foreground sm:px-6 lg:px-10">
          <div className="flex items-center gap-2">
            <span className="h-3.5 w-1 rounded-full bg-primary" />
            <span className="font-semibold text-foreground">Slowfeed</span>
          </div>
          <Link to="/about" className="transition-colors hover:text-foreground">
            About &amp; credits
          </Link>
        </div>
      </footer>
    </div>
  );
};

export const Route = createFileRoute('/')({
  component: LandingPage,
});
