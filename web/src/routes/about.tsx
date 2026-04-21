import { createFileRoute } from '@tanstack/react-router';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '#/components/ui/card';
import { H2 } from '#/components/ui/typography';
import {
  Rss,
  CalendarClock,
  Mail,
  Layers,
  Globe,
  GitBranch,
} from 'lucide-react';

const features = [
  {
    icon: Rss,
    title: 'RSS Feed Aggregation',
    description:
      'Add any RSS or Atom feed as a source for your newsletter content.',
  },
  {
    icon: Layers,
    title: 'Multiple Newsletters',
    description:
      'Organize content into separate newsletters, each with its own feeds and schedule.',
  },
  {
    icon: CalendarClock,
    title: 'Flexible Scheduling',
    description:
      'Send daily, weekly, or monthly — pick the day and time that works for you.',
  },
  {
    icon: Mail,
    title: 'Email Delivery',
    description:
      'Compiled digests are delivered straight to your inbox on schedule.',
  },
];

const AboutPage = () => {
  return (
    <div className="p-6 space-y-6">
      <H2>About</H2>

      <Card>
        <CardHeader>
          <CardTitle>What is Custom Newsletters?</CardTitle>
          <CardDescription>A personal RSS-to-email digest tool</CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground space-y-2">
          <p>
            Custom Newsletters lets you curate content from any RSS or Atom feed
            and receive it as a scheduled email digest. Instead of checking
            feeds manually, your reading lands in your inbox on a schedule you
            control.
          </p>
          <p>
            Create multiple newsletters to keep topics separate — tech news,
            blogs, podcasts, or whatever you follow — each with its own sources
            and delivery schedule.
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Features</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="grid sm:grid-cols-2 gap-4">
            {features.map(({ icon: Icon, title, description }) => (
              <li key={title} className="flex gap-3">
                <Icon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                <div className="space-y-0.5">
                  <p className="text-sm font-medium">{title}</p>
                  <p className="text-sm text-muted-foreground">{description}</p>
                </div>
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>How It Works</CardTitle>
        </CardHeader>
        <CardContent>
          <ol className="space-y-3 text-sm text-muted-foreground list-decimal list-inside">
            <li>Create a newsletter and give it a name.</li>
            <li>Add one or more RSS feed URLs as content sources.</li>
            <li>Set a delivery schedule — daily, weekly, or monthly.</li>
            <li>Receive a compiled digest in your inbox at the chosen time.</li>
          </ol>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Built by Ben Riethmeier</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground space-y-4">
          <div className="flex flex-wrap gap-4">
            <a
              href="https://benriethmeier.dev"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 hover:text-foreground transition-colors"
            >
              <Globe className="size-4" />
              Website
            </a>
            <a
              href="https://github.com/ben-rieth"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 hover:text-foreground transition-colors"
            >
              <GitBranch className="size-4" />
              GitHub
            </a>
          </div>
          <div className="grid sm:grid-cols-2 gap-x-6 gap-y-4">
            <div className="space-y-1.5">
              <p className="text-xs font-semibold uppercase tracking-wide text-foreground">
                Backend
              </p>
              <ul className="space-y-1 text-muted-foreground">
                <li>Go</li>
                <li>Huma (OpenAPI framework)</li>
                <li>sqlc + PostgreSQL</li>
              </ul>
            </div>
            <div className="space-y-1.5">
              <p className="text-xs font-semibold uppercase tracking-wide text-foreground">
                Frontend
              </p>
              <ul className="space-y-1 text-muted-foreground">
                <li>React 19 + TanStack Start</li>
                <li>TanStack Router &amp; Query</li>
                <li>Tailwind CSS v4 + shadcn/ui</li>
                <li>openapi-fetch</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export const Route = createFileRoute('/about')({
  component: AboutPage,
});
