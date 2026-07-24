import { Field, FieldLabel } from '#/components/ui/field';
import { Input } from '#/components/ui/input';

type Props = {
  id: string;
  label: string;
  hour: number;
  minute: number;
  onChange: (hour: number, minute: number) => void;
};

export const TimeField = ({ id, label, hour, minute, onChange }: Props) => {
  const value = `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`;

  return (
    <Field>
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      <Input
        id={id}
        type="time"
        value={value}
        className="appearance-none bg-background [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
        onChange={(e) => {
          const [h, m] = (e.target.value || '00:00').split(':').map(Number);
          onChange(isNaN(h) ? 0 : h, isNaN(m) ? 0 : m);
        }}
      />
    </Field>
  );
};
