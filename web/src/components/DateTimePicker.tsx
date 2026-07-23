import { useState } from 'react';
import * as chrono from 'chrono-node';
import { CalendarIcon } from 'lucide-react';
import type { Matcher } from 'react-day-picker';
import { Calendar } from '#/components/ui/calendar';
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
} from '#/components/ui/input-group';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '#/components/ui/popover';

type Props = {
  /** The currently selected date and time, or `undefined` if unset. */
  value: Date | undefined;
  onChange: (value: Date | undefined) => void;
  /** Dates before this are disabled in the calendar. */
  minDate?: Date;
  /** Dates after this are disabled in the calendar. */
  maxDate?: Date;
  id?: string;
  placeholder?: string;
};

const pad = (n: number) => n.toString().padStart(2, '0');

const formatDate = (date: Date | undefined) =>
  date
    ? date.toLocaleDateString('en-US', {
        day: '2-digit',
        month: 'long',
        year: 'numeric',
      })
    : '';

const formatTime = (date: Date | undefined) =>
  date ? `${pad(date.getHours())}:${pad(date.getMinutes())}` : '09:00';

/** Combines a calendar date with a `HH:mm` time string into a single Date. */
const combine = (date: Date, time: string): Date => {
  const [hours, minutes] = time.split(':').map(Number);
  const result = new Date(date);
  result.setHours(hours || 0, minutes || 0, 0, 0);
  return result;
};

/**
 * A friendly date + time picker: a natural-language text field ("tomorrow at
 * 9am") backed by a calendar popover, plus a time input. Controlled via
 * `value`/`onChange`, emitting a single combined Date.
 */
export const DateTimePicker = ({
  value,
  onChange,
  minDate,
  maxDate,
  id = 'date-time-picker',
  placeholder = 'Tomorrow at 9am',
}: Props) => {
  const [calendarOpen, setCalendarOpen] = useState(false);
  const [text, setText] = useState(formatDate(value));
  const [month, setMonth] = useState<Date | undefined>(value);
  const [time, setTime] = useState(formatTime(value));

  const handleTextChange = (input: string) => {
    setText(input);
    const results = chrono.parse(input);
    if (results.length === 0) return;
    const parsed = results[0].start;
    const parsedDate = parsed.date();
    setMonth(parsedDate);
    const nextTime = parsed.isCertain('hour')
      ? `${pad(parsedDate.getHours())}:${pad(parsedDate.getMinutes())}`
      : time;
    setTime(nextTime);
    onChange(combine(parsedDate, nextTime));
  };

  const handleDateSelect = (selected: Date | undefined) => {
    if (!selected) return;
    setMonth(selected);
    setText(formatDate(selected));
    setCalendarOpen(false);
    onChange(combine(selected, time));
  };

  const handleTimeChange = (nextTime: string) => {
    setTime(nextTime);
    if (value) {
      onChange(combine(value, nextTime));
    }
  };

  const disabledDates: Matcher[] = [];
  if (minDate) disabledDates.push({ before: minDate });
  if (maxDate) disabledDates.push({ after: maxDate });

  return (
    <div className="flex gap-3">
      <div className="flex-1 space-y-1.5">
        <InputGroup>
          <InputGroupInput
            id={id}
            value={text}
            placeholder={placeholder}
            onChange={(e) => handleTextChange(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'ArrowDown') {
                e.preventDefault();
                setCalendarOpen(true);
              }
            }}
          />
          <InputGroupAddon align="inline-end">
            <Popover open={calendarOpen} onOpenChange={setCalendarOpen}>
              <PopoverTrigger
                render={
                  <InputGroupButton size="icon-xs" aria-label="Pick a date" />
                }
              >
                <CalendarIcon />
              </PopoverTrigger>
              <PopoverContent
                className="w-auto overflow-hidden p-0"
                align="end"
              >
                <Calendar
                  mode="single"
                  selected={value}
                  month={month}
                  onMonthChange={setMonth}
                  disabled={disabledDates}
                  onSelect={handleDateSelect}
                />
              </PopoverContent>
            </Popover>
          </InputGroupAddon>
        </InputGroup>
      </div>

      <input
        type="time"
        aria-label="Time"
        value={time}
        onChange={(e) => handleTimeChange(e.target.value)}
        className="h-7 w-28 rounded-md border border-input bg-input/20 px-2 py-0.5 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 dark:bg-input/30 [&::-webkit-calendar-picker-indicator]:hidden"
      />
    </div>
  );
};
