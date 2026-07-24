import { formatRelativeTime } from '#/utils/format';
import type {
  DetailedIssue,
  IssueFeed,
  IssueItem,
} from '#/features/issues/queries/issues';

interface IssueDetailProps {
  issue: DetailedIssue;
}

const IssueDetail = ({ issue }: IssueDetailProps) => {
  const feeds = issue.feeds ?? [];
  const feedCount = feeds.length;
  const itemCount = feeds.reduce(
    (sum, feed) => sum + (feed.items ?? []).length,
    0,
  );

  const sentDate = new Date(issue.sentAt).toLocaleDateString('en-US', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  });

  const sortedFeeds = [...feeds].sort((a: IssueFeed, b: IssueFeed) => {
    const aRead = (a.items ?? []).every((i) => i.state === 'read');
    const bRead = (b.items ?? []).every((i) => i.state === 'read');
    if (aRead !== bRead) return aRead ? 1 : -1;
    return a.title.localeCompare(b.title);
  });

  return (
    <div className="space-y-8">
      <header className="border-b border-border pb-6">
        <h1 className="text-3xl font-bold tracking-tight">
          {issue.newsletterName}
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          {sentDate} · {itemCount} {itemCount === 1 ? 'item' : 'items'} from{' '}
          {feedCount} {feedCount === 1 ? 'feed' : 'feeds'}
        </p>
      </header>

      {feeds.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          No content in this issue.
        </p>
      ) : (
        <div className="space-y-8">
          {sortedFeeds.map((feed: IssueFeed, feedIdx: number) => (
            <section key={feedIdx} className="space-y-4">
              <a
                href={feed.webUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="text-xs font-semibold uppercase tracking-wide text-primary hover:underline"
              >
                {feed.title}
              </a>

              <ul className="space-y-5">
                {(feed.items ?? []).map((item: IssueItem) => {
                  const read = item.state === 'read';
                  return (
                    <li
                      key={item.itemId}
                      className="border-b border-border/50 pb-5 last:border-0"
                    >
                      <a
                        href={`/api/link/${item.token}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className={`block text-lg font-semibold leading-snug transition-colors hover:text-primary ${
                          read ? 'text-muted-foreground' : 'text-foreground'
                        }`}
                      >
                        {item.title}
                      </a>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {formatRelativeTime(item.publishDate)}
                        {read && ' · read'}
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
