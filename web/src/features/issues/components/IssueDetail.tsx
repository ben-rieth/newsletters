import type {
  DetailedIssue,
  IssueFeed,
  IssueItem,
} from '#/features/issues/queries/issues';

interface IssueDetailProps {
  issue: DetailedIssue;
}

const IssueDetail = ({ issue }: IssueDetailProps) => {
  const sentDate = new Date(issue.sentAt).toLocaleDateString(undefined, {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });

  return (
    <div className="max-w-2xl space-y-8">
      <div className="border-b pb-4">
        <h1 className="text-2xl font-bold">{issue.newsletterName}</h1>
        <p className="text-sm text-muted-foreground mt-1">Sent {sentDate}</p>
      </div>

      {(issue.feeds ?? []).length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No content in this issue.
        </p>
      ) : (
        <div className="space-y-10">
          {[...(issue.feeds ?? [])]
            .sort((a: IssueFeed, b: IssueFeed) => {
              const aRead = (a.items ?? []).every((i) => i.state === 'read');
              const bRead = (b.items ?? []).every((i) => i.state === 'read');
              if (aRead !== bRead) return aRead ? 1 : -1;
              return a.title.localeCompare(b.title);
            })
            .map((feed: IssueFeed, feedIdx: number) => (
              <section key={feedIdx}>
                <a
                  href={feed.webUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-lg font-semibold hover:underline"
                >
                  {feed.title}
                </a>

                <ul className="mt-3 space-y-3">
                  {(feed.items ?? []).map((item: IssueItem) => {
                    const read = item.state === 'read';
                    return (
                      <li key={item.itemId}>
                        <a
                          href={`/api/link/${item.token}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className={`text-sm font-medium hover:underline ${read ? 'text-muted-foreground line-through' : 'text-foreground'}`}
                        >
                          {item.title}
                        </a>
                        <p className="text-xs text-muted-foreground mt-0.5">
                          {new Date(item.publishDate).toLocaleDateString(
                            undefined,
                            {
                              month: 'short',
                              day: 'numeric',
                              year: 'numeric',
                            },
                          )}
                        </p>
                      </li>
                    );
                  })}
                </ul>
              </section>
            ))}
        </div>
      )}
    </div>
  );
};

export default IssueDetail;
