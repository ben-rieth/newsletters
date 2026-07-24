import { LayoutGrid, Table2 } from 'lucide-react';
import { ToggleGroup, ToggleGroupItem } from '#/components/ui/toggle-group';

export type View = 'table' | 'cards';

type Props = {
  value: View;
  onChange: (value: View) => void;
};

export const ViewToggle = ({ value, onChange }: Props) => (
  <ToggleGroup
    value={[value]}
    onValueChange={(values) => {
      if (values[0]) {
        onChange(values[0] as View);
      }
    }}
    variant="outline"
    size="default"
  >
    <ToggleGroupItem value="table" aria-label="Table view">
      <Table2 className="size-4" />
    </ToggleGroupItem>
    <ToggleGroupItem value="cards" aria-label="Card view">
      <LayoutGrid className="size-4" />
    </ToggleGroupItem>
  </ToggleGroup>
);
