import { Temporal } from 'temporal-polyfill';

/** Phrases an upcoming send as "today at 08:00" / "Friday at 08:00". */
export const formatUpcoming = (dateStr: string): string => {
  const date = new Date(dateStr);
  const time = date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });

  const now = new Date();
  if (date.toDateString() === now.toDateString()) return `today at ${time}`;

  const tomorrow = new Date(now);
  tomorrow.setDate(tomorrow.getDate() + 1);
  if (date.toDateString() === tomorrow.toDateString()) {
    return `tomorrow at ${time}`;
  }

  const withinWeek = date.getTime() - now.getTime() < 7 * 24 * 60 * 60 * 1000;
  const day = date.toLocaleDateString('en-US', {
    weekday: 'long',
    ...(withinWeek ? {} : { month: 'short', day: 'numeric' }),
  });
  return `${day} at ${time}`;
};

export const formatRelativeTime = (dateStr: string): string => {
  const zone = Temporal.Now.timeZoneId();
  const then = Temporal.Instant.from(dateStr).toZonedDateTimeISO(zone);
  const now = Temporal.Now.zonedDateTimeISO();
  const duration = then.until(now, { largestUnit: 'years' });

  const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

  if (Math.abs(duration.years) >= 1) {
    return rtf.format(-duration.years, 'year');
  }

  if (Math.abs(duration.months) >= 1) {
    return rtf.format(-duration.months, 'month');
  }

  if (Math.abs(duration.days) >= 1) {
    return rtf.format(-duration.days, 'day');
  }

  if (Math.abs(duration.hours) >= 1) {
    return rtf.format(-duration.hours, 'hour');
  }

  if (Math.abs(duration.minutes) >= 1) {
    return rtf.format(-duration.minutes, 'minute');
  }

  return rtf.format(-duration.seconds, 'second');
};
