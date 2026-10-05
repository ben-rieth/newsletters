import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '#/components/ui/select';
import useUpdateIssueRetention from '../queries/hooks/useUpdateIssueRetention';

const RETENTION_OPTIONS = [
  { value: 0, label: 'Forever' },
  { value: 7, label: 'One week' },
  { value: 30, label: 'One month' },
  { value: 90, label: 'Three months' },
  { value: 180, label: 'Six months' },
  { value: 365, label: 'One year' },
];

type Props = {
  issueRetentionDays: number;
};

const IssueRetentionSetting = ({ issueRetentionDays }: Props) => {
  const updateRetention = useUpdateIssueRetention();

  // A value set before these presets existed would otherwise leave the trigger
  // blank, making it look like nothing is configured at all.
  const items = RETENTION_OPTIONS.some(
    (option) => option.value === issueRetentionDays,
  )
    ? RETENTION_OPTIONS
    : [
        ...RETENTION_OPTIONS,
        { value: issueRetentionDays, label: `${issueRetentionDays} days` },
      ];

  return (
    <Select
      value={issueRetentionDays}
      onValueChange={(value: number | null) => {
        if (value !== null) updateRetention.mutate(value);
      }}
      items={items}
      disabled={updateRetention.isPending}
    >
      <SelectTrigger className="w-full" aria-label="Keep issues for">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {items.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            {option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
};

export default IssueRetentionSetting;
