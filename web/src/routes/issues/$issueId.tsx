import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import IssueDetail from '#/features/issues/components/IssueDetail';
import { issueDetailOptions } from '#/features/issues/queries/issues';

const IssuePage = () => {
  const { issueId } = Route.useParams();
  const { data: issue } = useSuspenseQuery(issueDetailOptions(issueId));

  return (
    <div className="mx-auto max-w-3xl px-6 py-8 lg:px-10">
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
