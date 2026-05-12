import { Link } from '@tanstack/react-router';
import type { Issue } from '#/features/newsletters/queries/issues';

interface IssuesListProps {
  issues: Issue[];
}

const IssuesList = ({ issues }: IssuesListProps) => {
  if (issues.length === 0) {
    return (
      <p className="text-sm text-muted-foreground py-8 text-center">
        No issues yet — newsletters will appear here after they're first sent.
      </p>
    );
  }

  const sorted = [...issues].sort(
    (a, b) => new Date(b.sentAt).getTime() - new Date(a.sentAt).getTime(),
  );

  return (
    <div className="divide-y border rounded-md">
      {sorted.map((issue) => (
        <Link
          key={issue.issueId}
          to="/issues/$issueId"
          params={{ issueId: issue.issueId }}
          className="flex items-center justify-between px-4 py-3 hover:bg-muted/50 transition-colors"
        >
          <span className="font-medium text-sm">{issue.newsletterName}</span>
          <span className="text-sm text-muted-foreground">
            {new Date(issue.sentAt).toLocaleDateString(undefined, {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
            })}
          </span>
        </Link>
      ))}
    </div>
  );
};

export default IssuesList;
