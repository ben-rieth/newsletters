import { useState } from 'react';
import { Select as SelectPrimitive } from '@base-ui/react/select';
import { CheckIcon, ChevronDownIcon, SearchIcon } from 'lucide-react';

const COMMON_TIMEZONES = [
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'America/Anchorage',
  'Pacific/Honolulu',
  'America/Sao_Paulo',
  'America/Toronto',
  'America/Vancouver',
  'America/Mexico_City',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Europe/Moscow',
  'Africa/Cairo',
  'Africa/Johannesburg',
  'Asia/Dubai',
  'Asia/Kolkata',
  'Asia/Bangkok',
  'Asia/Singapore',
  'Asia/Shanghai',
  'Asia/Tokyo',
  'Asia/Seoul',
  'Australia/Sydney',
  'Australia/Perth',
  'Pacific/Auckland',
  'UTC',
];

const now = new Date();

const getGenericName = (tz: string): string =>
  new Intl.DateTimeFormat('en', { timeZone: tz, timeZoneName: 'longGeneric' })
    .formatToParts(now)
    .find((p) => p.type === 'timeZoneName')?.value ?? tz;

type TimezoneOption = { iana: string; genericName: string };

const toOption = (tz: string): TimezoneOption => ({
  iana: tz,
  genericName: getGenericName(tz),
});

const USER_TIMEZONE = Intl.DateTimeFormat().resolvedOptions().timeZone;

const SUGGESTED: TimezoneOption[] = [
  USER_TIMEZONE,
  ...COMMON_TIMEZONES.filter((tz) => tz !== USER_TIMEZONE),
]
  .filter((tz) => {
    try {
      Intl.DateTimeFormat('en', { timeZone: tz });
      return true;
    } catch {
      return false;
    }
  })
  .map(toOption);

const SUGGESTED_SET = new Set(SUGGESTED.map((tz) => tz.iana));

const ALL_OTHERS: TimezoneOption[] = Intl.supportedValuesOf('timeZone')
  .filter((tz) => !SUGGESTED_SET.has(tz))
  .map(toOption);

type Props = {
  id?: string;
  value: string;
  onValueChange: (value: string) => void;
};

export const TimezoneSelect = ({ id, value, onValueChange }: Props) => {
  const [search, setSearch] = useState('');

  const q = search.trim().toLowerCase();
  const isSearching = q.length > 0;
  const matches = (tz: TimezoneOption) =>
    tz.genericName.toLowerCase().includes(q) ||
    tz.iana.toLowerCase().includes(q);

  const filteredSuggested = isSearching ? SUGGESTED.filter(matches) : SUGGESTED;
  const filteredOthers = isSearching ? ALL_OTHERS.filter(matches) : [];

  const selected =
    SUGGESTED.find((tz) => tz.iana === value) ??
    ALL_OTHERS.find((tz) => tz.iana === value);

  const renderItem = ({ iana, genericName }: TimezoneOption) => (
    <SelectPrimitive.Item
      key={iana}
      value={iana}
      className="relative flex min-h-7 w-full cursor-default items-center gap-2 rounded-md px-2 py-1 pr-7 text-xs/relaxed outline-hidden select-none focus:bg-accent focus:text-accent-foreground data-disabled:pointer-events-none data-disabled:opacity-50"
    >
      <SelectPrimitive.ItemText className="flex flex-1 gap-1.5">
        <span>{genericName}</span>
        <span className="text-muted-foreground">{iana}</span>
      </SelectPrimitive.ItemText>
      <SelectPrimitive.ItemIndicator className="pointer-events-none absolute right-2 flex items-center justify-center">
        <CheckIcon className="size-3.5" />
      </SelectPrimitive.ItemIndicator>
    </SelectPrimitive.Item>
  );

  return (
    <SelectPrimitive.Root
      value={value}
      onValueChange={(value) => value != null && onValueChange(value)}
      onOpenChange={(open) => {
        if (!open) {
          setSearch('');
        }
      }}
    >
      <SelectPrimitive.Trigger
        id={id}
        className="flex h-7 w-full items-center justify-between gap-1.5 rounded-md border border-input bg-input/20 px-2 text-xs/relaxed whitespace-nowrap transition-colors outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50 data-placeholder:text-muted-foreground dark:bg-input/30 dark:hover:bg-input/50 [&_svg]:pointer-events-none [&_svg]:shrink-0"
      >
        <SelectPrimitive.Value placeholder="Select timezone">
          {selected && (
            <span className="flex items-center gap-1.5 truncate">
              <span>{selected.genericName}</span>
              <span className="text-muted-foreground">{selected.iana}</span>
            </span>
          )}
        </SelectPrimitive.Value>
        <SelectPrimitive.Icon
          render={
            <ChevronDownIcon className="pointer-events-none size-3.5 shrink-0 text-muted-foreground" />
          }
        />
      </SelectPrimitive.Trigger>

      <SelectPrimitive.Portal>
        <SelectPrimitive.Positioner sideOffset={4} className="isolate z-50">
          <SelectPrimitive.Popup className="flex w-(--anchor-width) min-w-32 origin-(--transform-origin) flex-col overflow-hidden rounded-lg bg-popover text-popover-foreground shadow-md ring-1 ring-foreground/10 duration-100 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95">
            {/* Search input — stopPropagation prevents Select from intercepting keystrokes */}
            <div className="border-b border-border/50 p-1">
              <div className="flex items-center gap-1.5 rounded-md bg-input/20 px-2">
                <SearchIcon className="size-3.5 shrink-0 text-muted-foreground" />
                <input
                  className="h-7 flex-1 bg-transparent text-xs/relaxed outline-none placeholder:text-muted-foreground"
                  placeholder="Search timezone..."
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  onKeyDown={(e) => e.stopPropagation()}
                />
              </div>
            </div>

            <SelectPrimitive.List className="no-scrollbar max-h-64 overflow-y-auto p-1">
              {filteredSuggested.length > 0 && (
                <SelectPrimitive.Group className="scroll-my-1 p-1">
                  {!isSearching && (
                    <SelectPrimitive.GroupLabel className="px-2 py-1.5 text-xs text-muted-foreground">
                      Suggested
                    </SelectPrimitive.GroupLabel>
                  )}
                  {filteredSuggested.map(renderItem)}
                </SelectPrimitive.Group>
              )}

              {filteredSuggested.length > 0 && filteredOthers.length > 0 && (
                <div className="-mx-1 my-1 h-px bg-border/50" />
              )}

              {filteredOthers.length > 0 && (
                <SelectPrimitive.Group className="scroll-my-1 p-1">
                  {filteredOthers.map(renderItem)}
                </SelectPrimitive.Group>
              )}

              {filteredSuggested.length === 0 &&
                filteredOthers.length === 0 && (
                  <p className="py-6 text-center text-xs/relaxed text-muted-foreground">
                    No timezone found.
                  </p>
                )}
            </SelectPrimitive.List>
          </SelectPrimitive.Popup>
        </SelectPrimitive.Positioner>
      </SelectPrimitive.Portal>
    </SelectPrimitive.Root>
  );
};
