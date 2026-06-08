import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

// The smart-clock start time: continue from the most recent entry's end so
// consecutive entries don't overlap or leave gaps; with no completed entry to
// continue from, fall back to the user's configured default start (HH:MM) on
// the given day.
export function smartFillStart(recentEntries: ITimeEntry[], defaultStart: string, now: Date): Date {
	const lastEnd = recentEntries
		.map(entry => entry.endTime)
		.filter((end): end is Date => end !== null)
		.sort((a, b) => b.getTime() - a.getTime())[0]
	if (lastEnd !== undefined) {
		return new Date(lastEnd)
	}

	const [hours, minutes] = (defaultStart || '09:00').split(':').map(Number)
	const start = new Date(now)
	start.setHours(hours || 0, minutes || 0, 0, 0)
	return start
}
