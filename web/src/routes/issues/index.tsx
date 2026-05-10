import { useState } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { H2 } from '#/components/ui/typography';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import IssuesList from '#/features/issues/components/IssuesList';
import { issuesOptions } from '#/features/issues/queries/issues';

const IssuesPage = () => {
  const { data: issues } = useSuspenseQuery(issuesOptions);
  const [filter, setFilter] = useState<string | null>(null);

  const allIssues = issues ?? [];
  const newsletterNames = [
    ...new Set(allIssues.map((i) => i.newsletterName)),
  ].sort();
  const filtered =
    filter === null
      ? allIssues
      : allIssues.filter((i) => i.newsletterName === filter);

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <H2>Issues</H2>
        {newsletterNames.length > 1 && (
          <Select value={filter} onValueChange={(val) => setFilter(val)}>
            <SelectTrigger>
              <SelectValue placeholder="Select a newsletter" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={null} label="All Newsletters">
                All Newsletters
              </SelectItem>
              {newsletterNames.map((name) => (
                <SelectItem key={name} value={name}>
                  {name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </div>
      <IssuesList issues={filtered} />
    </div>
  );
};

export const Route = createFileRoute('/issues/')({
  component: IssuesPage,
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(issuesOptions);
  },
});
