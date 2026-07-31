import { createFileRoute } from '@tanstack/react-router';
import { Globe, GitBranch } from 'lucide-react';
import { SettingsRow, SettingsSection } from '#/components/SettingsRow';
import { useMobileHeader } from '#/components/MobileHeader';

const STACK = [
  {
    heading: 'Backend',
    items: ['Go', 'Huma (OpenAPI framework)', 'sqlc + PostgreSQL'],
  },
  {
    heading: 'Frontend',
    items: [
      'React 19 + TanStack Start',
      'TanStack Router & Query',
      'Tailwind CSS v4 + shadcn/ui',
      'openapi-fetch',
    ],
  },
];

const AboutPage = () => {
  useMobileHeader({ title: 'About' });

  return (
    <div className="mx-auto max-w-3xl space-y-10 px-4 py-6 md:px-6 md:py-8 lg:px-10">
      <header className="space-y-2">
        <h1 className="hidden text-2xl font-semibold tracking-tight md:block">
          About Slowfeed
        </h1>
        <p className="text-sm text-muted-foreground">
          A personal RSS-to-email digest tool. This page covers who builds it,
          what it runs on, and the work it stands on.
        </p>
      </header>

      <SettingsSection>
        <SettingsRow
          title="Maker"
          description="Slowfeed is built and maintained by Ben Riethmeier."
        >
          <div className="flex flex-wrap gap-4 text-sm">
            <a
              href="https://benriethmeier.dev"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 text-muted-foreground transition-colors hover:text-foreground"
            >
              <Globe className="size-4" />
              Website
            </a>
            <a
              href="https://github.com/ben-rieth"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 text-muted-foreground transition-colors hover:text-foreground"
            >
              <GitBranch className="size-4" />
              GitHub
            </a>
          </div>
        </SettingsRow>

        <SettingsRow
          title="Tech stack"
          description="The tools Slowfeed is built with."
        >
          <div className="grid gap-x-6 gap-y-4 sm:grid-cols-2">
            {STACK.map((group) => (
              <div key={group.heading} className="space-y-1.5">
                <p className="text-xs font-semibold uppercase tracking-wide text-foreground">
                  {group.heading}
                </p>
                <ul className="space-y-1 text-sm text-muted-foreground">
                  {group.items.map((item) => (
                    <li key={item}>{item}</li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </SettingsRow>

        <SettingsRow
          title="Attribution"
          description="Assets used under open licenses."
        >
          <p className="text-sm text-muted-foreground">
            Favicon: &ldquo;Newspaper&rdquo; emoji (U+1F4F0) by Twitter, Inc and
            other contributors.{' '}
            <a
              href="https://github.com/twitter/twemoji/blob/master/assets/svg/1f4f0.svg"
              target="_blank"
              rel="noopener noreferrer"
              className="underline underline-offset-4 transition-colors hover:text-foreground"
            >
              Source
            </a>
            . Licensed under{' '}
            <a
              href="https://creativecommons.org/licenses/by/4.0/"
              target="_blank"
              rel="noopener noreferrer"
              className="underline underline-offset-4 transition-colors hover:text-foreground"
            >
              CC BY 4.0
            </a>
            .
          </p>
        </SettingsRow>
      </SettingsSection>
    </div>
  );
};

export const Route = createFileRoute('/about')({
  component: AboutPage,
});
