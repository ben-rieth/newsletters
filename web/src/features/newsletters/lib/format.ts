import type { components } from '#/api/schema'

type Newsletter = components['schemas']['Newsletter']

export const DAY_NAMES = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
]

const padTime = (n: number) => String(n).padStart(2, '0')

// Assumes sendDay follows JS Date.getDay() convention: 0=Sunday, 6=Saturday
export const formatSchedule = (n: Newsletter): string => {
  const time = `${padTime(n.sendHour)}:${padTime(n.sendMinute)}`

  if (n.frequency === 'daily') {
    return `Daily at ${time}`
  }

  if (n.frequency === 'weekly') {
    const day = n.sendDay != null ? (DAY_NAMES[n.sendDay] ?? 'Unknown') : 'Unknown'
    return `Every ${day} at ${time}`
  }

  if (n.frequency === 'monthly') {
    return `Day ${n.sendDay ?? '?'} of month at ${time}`
  }

  return time
}
