import { Temporal } from 'temporal-polyfill';

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
