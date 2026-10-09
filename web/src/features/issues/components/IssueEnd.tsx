import { Link } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { cn } from '#/lib/utils';
import { buttonVariants } from '#/components/ui/button';
import { issuesOptions } from '../queries/issues';
import type { Issue } from '../queries/issues';

const newestFirst = (a: Issue, b: Issue) =>
  new Date(b.sentAt).getTime() - new Date(a.sentAt).getTime();

const findNextUnread = (issues: Issue[], currentId: string) => {
  const sorted = issues.toSorted(newestFirst);
  const isNextCandidate = (issue: Issue) =>
    issue.state === 'unread' && issue.issueId !== currentId;
  const currentIndex = sorted.findIndex((i) => i.issueId === currentId);

  return (
    sorted.slice(currentIndex + 1).find(isNextCandidate) ??
    sorted.find(isNextCandidate)
  );
};

export const IssueEnd = ({ issueId }: { issueId: string }) => {
  const { data } = useQuery(issuesOptions);
  const next = findNextUnread(data ?? [], issueId);

  return (
    <footer className="animate-in border-t border-border pt-8 duration-500 ease-out fade-in slide-in-from-bottom-2">
      <p className="font-serif text-xl font-medium">That’s the whole issue.</p>
      {next ? (
        <Link
          to="/issues/$issueId"
          params={{ issueId: next.issueId }}
          className={cn(buttonVariants({ variant: 'outline' }), 'mt-4')}
        >
          Next: {next.newsletterName},{' '}
          {new Date(next.sentAt).toLocaleDateString('en-US', {
            weekday: 'long',
          })}
        </Link>
      ) : (
        <p className="mt-2 text-sm text-muted-foreground">
          Nothing else is waiting. See you at the next send.
        </p>
      )}
    </footer>
  );
};
