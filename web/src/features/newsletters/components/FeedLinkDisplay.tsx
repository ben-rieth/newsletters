import { Label } from '#/components/ui/label';
import { ExternalLink } from 'lucide-react';

interface FeedLinkDisplayProps {
  label: string;
  url: string;
}

const FeedLinkDisplay = ({ label, url }: FeedLinkDisplayProps) => {
  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      <div className="flex items-start gap-1.5">
        <p className="text-sm text-muted-foreground break-all">{url}</p>
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          aria-label="Open feed URL"
          className="mt-0.5 shrink-0 text-muted-foreground hover:text-foreground transition-colors"
        >
          <ExternalLink className="size-3.5" />
        </a>
      </div>
    </div>
  );
};

export default FeedLinkDisplay;
