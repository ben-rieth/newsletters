import { createFileRoute, Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import IssueDetail from '#/features/issues/components/IssueDetail';
import { issueDetailOptions } from '#/features/issues/queries/issues';

const IssuePage = () => {
  const { issueId } = Route.useParams();
  const { data: issue } = useSuspenseQuery(issueDetailOptions(issueId));

  return (
    <div className="p-6 space-y-6">
      <Link
        to="/issues"
        className="text-sm text-muted-foreground hover:underline"
      >
        ← Issues
      </Link>
      <IssueDetail issue={issue} />
    </div>
  );
};

export const Route = createFileRoute('/issues/$issueId')({
  component: IssuePage,
  loader: async ({ context, params }) => {
    await context.queryClient.ensureQueryData(
      issueDetailOptions(params.issueId),
    );
  },
});
