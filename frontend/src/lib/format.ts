export function formatDate(value?: string | null, withTime = false): string {
  if (!value) return '—'
  return new Intl.DateTimeFormat('en', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    ...(withTime ? { hour: 'numeric', minute: '2-digit' } : {}),
  }).format(new Date(value))
}

export function relativeTime(value?: string | null): string {
  if (!value) return 'Never'
  const date = new Date(value).getTime()
  const seconds = Math.round((date - Date.now()) / 1000)
  const formatter = new Intl.RelativeTimeFormat('en', { numeric: 'auto' })
  const ranges: Array<[Intl.RelativeTimeFormatUnit, number]> = [
    ['year', 31_536_000],
    ['month', 2_592_000],
    ['week', 604_800],
    ['day', 86_400],
    ['hour', 3_600],
    ['minute', 60],
  ]
  for (const [unit, divisor] of ranges) {
    if (Math.abs(seconds) >= divisor) return formatter.format(Math.round(seconds / divisor), unit)
  }
  return 'Just now'
}

export function scoreTone(score: number, inverse = false): string {
  const value = inverse ? 100 - score : score
  if (value >= 70) return '#9ef01a'
  if (value >= 40) return '#f6c453'
  return '#ff5d73'
}

export function titleCase(value: string): string {
  return value.replace(/_/g, ' ').replace(/\b\w/g, (character) => character.toUpperCase())
}

