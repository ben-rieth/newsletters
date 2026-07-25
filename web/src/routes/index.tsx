import { useState } from 'react';
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { Button, buttonVariants } from '#/components/ui/button';
import { Input } from '#/components/ui/input';
import useIsSignedIn from '#/features/auth/queries/hooks/useIsSignedIn';

const NAV_LINKS = [
  { href: '#how-it-works', label: 'How it works' },
  { href: '#what-you-can-add', label: 'What you can add' },
];

const STEPS = [
  {
    n: '01',
    title: 'Group the feeds',
    body: 'Make a newsletter — Entertainment, Morning News, whatever — and paste in RSS or Atom URLs. Each one is checked to make sure it resolves before it’s added.',
  },
  {
    n: '02',
    title: 'Pick when it lands',
    body: 'Daily, weekly, or a specific day, at the time and timezone you choose. Need an extra issue sooner? Schedule a one-off send without touching the routine.',
  },
  {
    n: '03',
    title: 'Read it once',
    body: 'The issue arrives as email and stays in the app, grouped and ready. Past issues keep — there’s nothing to keep up with between sends.',
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

const PREVIEW_ITEMS = [
  {
    source: 'Corridor Crew',
    when: '2h ago',
    title: 'VFX Artists React to Bad & Great CGi',
  },
  {
    source: 'The Criterion Channel',
    when: '4h ago',
    title: 'Now streaming: the complete Kelly Reichardt',
  },
  {
    source: 'A24',
    when: '1d ago',
    title: 'Notes on Film 44 — the sound of an empty room',
  },
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
    <div className="min-h-screen bg-background text-foreground">
      <header className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur">
        <nav className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6 lg:px-10">
          <div className="flex items-center gap-2">
            <span className="h-4 w-1 rounded-full bg-primary" />
            <span className="text-base font-semibold tracking-tight">
              Slowfeed
            </span>
          </div>

          <div className="hidden items-center gap-8 text-sm text-muted-foreground md:flex">
            {NAV_LINKS.map((link) => (
              <a
                key={link.href}
                href={link.href}
                className="transition-colors hover:text-foreground"
              >
                {link.label}
              </a>
            ))}
          </div>

          <div className="flex items-center gap-2">
            {isSignedIn ? (
              <Link to="/issues" className={buttonVariants()}>
                Go to Issues
              </Link>
            ) : (
              <>
                <Link
                  to="/sign-in"
                  className={buttonVariants({
                    variant: 'ghost',
                    className: 'hidden sm:inline-flex',
                  })}
                >
                  Sign in
                </Link>
                <Link to="/sign-up" className={buttonVariants()}>
                  Start reading
                </Link>
              </>
            )}
          </div>
        </nav>
      </header>

      <main>
        <section className="mx-auto max-w-6xl px-6 py-16 lg:px-10 lg:py-24">
          <div className="grid items-center gap-12 lg:grid-cols-[1.1fr_0.9fr]">
            <div className="space-y-8">
              <h1 className="text-5xl font-bold leading-[1.05] tracking-tight sm:text-6xl">
                Read on your schedule, not the algorithm’s.
              </h1>
              <p className="max-w-xl text-lg leading-relaxed text-muted-foreground">
                Slowfeed turns a list of RSS feeds into your own newsletter.
                Paste the URLs of the channels, sites, and blogs you actually
                follow, pick when it arrives, and read the whole lot in one
                sitting — by email or in the app.
              </p>

              {!isSignedIn && (
                <form
                  className="flex max-w-md flex-col gap-3 sm:flex-row"
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
            </div>

            <IssuePreview />
          </div>
        </section>

        <section
          id="how-it-works"
          className="border-t border-border/60 bg-card/30"
        >
          <div className="mx-auto max-w-6xl px-6 py-16 lg:px-10 lg:py-20">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-baseline sm:gap-6">
              <h2 className="text-3xl font-bold tracking-tight">
                Three steps, then nothing
              </h2>
              <p className="text-muted-foreground">
                Setup takes about two minutes per newsletter.
              </p>
            </div>

            <div className="mt-12 grid gap-10 sm:grid-cols-3">
              {STEPS.map((step) => (
                <div key={step.n} className="space-y-3">
                  <div className="border-t-2 border-primary pt-4">
                    <span className="font-mono text-sm text-primary">
                      {step.n}
                    </span>
                  </div>
                  <h3 className="text-lg font-semibold">{step.title}</h3>
                  <p className="text-sm leading-relaxed text-muted-foreground">
                    {step.body}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="what-you-can-add" className="border-t border-border/60">
          <div className="mx-auto grid max-w-6xl gap-12 px-6 py-16 lg:grid-cols-2 lg:px-10 lg:py-20">
            <div className="space-y-5">
              <h2 className="text-3xl font-bold tracking-tight">
                If it has a feed, it works
              </h2>
              <p className="max-w-md leading-relaxed text-muted-foreground">
                Slowfeed takes URLs — that’s the whole input. Most sites still
                publish a feed even when they don’t advertise one, and a YouTube
                channel has one for every channel ID.
              </p>
              <p className="max-w-md leading-relaxed text-muted-foreground">
                Point a feed at a newsletter and add filters to hide the items
                you don’t want, so a noisy source stays in its lane.
              </p>
            </div>

            <div className="divide-y divide-border border-y border-border">
              {FEED_EXAMPLES.map((example) => (
                <div
                  key={example.label}
                  className="grid grid-cols-[9rem_minmax(0,1fr)] items-center gap-4 py-4"
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

        <section className="border-t border-border/60 bg-card/30">
          <div className="mx-auto max-w-6xl space-y-6 px-6 py-16 text-center lg:px-10">
            <h2 className="text-3xl font-bold tracking-tight">
              Start a newsletter you’ll actually finish.
            </h2>
            {!isSignedIn && (
              <Link to="/sign-up" className={buttonVariants({ size: 'lg' })}>
                Start reading
              </Link>
            )}
          </div>
        </section>
      </main>

      <footer className="border-t border-border/60">
        <div className="mx-auto flex max-w-6xl flex-col gap-4 px-6 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between lg:px-10">
          <div className="flex items-center gap-2">
            <span className="h-3.5 w-1 rounded-full bg-primary" />
            <span className="font-semibold text-foreground">Slowfeed</span>
          </div>
          <div className="flex items-center gap-6">
            {!isSignedIn && (
              <Link
                to="/sign-in"
                className="transition-colors hover:text-foreground"
              >
                Sign in
              </Link>
            )}
            <Link
              to="/about"
              className="transition-colors hover:text-foreground"
            >
              About &amp; credits
            </Link>
          </div>
        </div>
      </footer>
    </div>
  );
};

const IssuePreview = () => (
  <div className="overflow-hidden rounded-xl border border-border bg-card shadow-2xl shadow-black/40">
    <div className="grid grid-cols-[8rem_minmax(0,1fr)] divide-x divide-border">
      <div className="space-y-4 p-4">
        <div className="space-y-1 text-sm">
          <div className="flex items-center justify-between rounded-md bg-primary/10 px-2 py-1 text-primary">
            <span className="font-medium">Issues</span>
            <span className="text-xs">3</span>
          </div>
          <p className="px-2 py-1 text-muted-foreground">Saved</p>
        </div>
        <div className="space-y-1">
          <p className="px-2 text-xs uppercase tracking-wider text-muted-foreground">
            Newsletters
          </p>
          <p className="px-2 py-1 text-sm text-foreground">Entertainment</p>
          <p className="px-2 py-1 text-sm text-muted-foreground">
            Morning News
          </p>
          <p className="px-2 py-1 text-sm text-muted-foreground">
            Dev &amp; Releases
          </p>
        </div>
      </div>

      <div className="space-y-4 p-5">
        <div>
          <p className="text-xs text-muted-foreground">
            Entertainment · daily at 08:00
          </p>
          <p className="mt-1 text-lg font-semibold">Friday, 24 July</p>
        </div>
        <div className="space-y-4">
          {PREVIEW_ITEMS.map((item) => (
            <div key={item.title} className="flex gap-3">
              <div className="mt-0.5 size-10 shrink-0 rounded bg-muted" />
              <div className="min-w-0">
                <p className="text-xs text-muted-foreground">
                  {item.source} · {item.when}
                </p>
                <p className="text-sm font-medium leading-snug">{item.title}</p>
              </div>
            </div>
          ))}
        </div>
        <p className="text-xs text-muted-foreground">+ 6 more items</p>
      </div>
    </div>
  </div>
);

export const Route = createFileRoute('/')({
  component: LandingPage,
});
