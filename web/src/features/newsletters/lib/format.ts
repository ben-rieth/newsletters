import type { components } from '#/api/schema';

type Newsletter = components['schemas']['Newsletter'];

export const DAY_NAMES = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
];

const padTime = (n: number) => String(n).padStart(2, '0');

// Assumes sendDay follows JS Date.getDay() convention: 0=Sunday, 6=Saturday
export const formatSchedule = (n: Newsletter): string => {
  const time = `${padTime(n.sendHour)}:${padTime(n.sendMinute)}`;

  if (n.frequency === 'daily') {
    return `Daily at ${time}`;
  }

  if (n.frequency === 'weekly') {
    const day = DAY_NAMES[n.sendDay] ?? 'Unknown';
    return `Every ${day} at ${time}`;
  }

  if (n.frequency === 'monthly') {
    return `${n.sendDay}${getOrdinalEnding(n.sendDay)} of the month at ${time}`;
  }

  return time;
};

const getOrdinalEnding = (dayOfMonth: number) => {
  const mod100 = dayOfMonth % 100;
  if (mod100 >= 11 && mod100 <= 13) return 'th';
  switch (dayOfMonth % 10) {
    case 1:
      return 'st';
    case 2:
      return 'nd';
    case 3:
      return 'rd';
    default:
      return 'th';
  }
};
