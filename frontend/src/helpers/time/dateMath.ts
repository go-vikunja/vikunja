export function startOfDay(date: Date): Date {
	const result = new Date(date)
	result.setHours(0, 0, 0, 0)
	return result
}

export function addDays(date: Date, days: number): Date {
	const result = new Date(date)
	result.setDate(result.getDate() + days)
	return result
}

export function isSameDay(a: Date | null | undefined, b: Date | null | undefined): boolean {
	if (!a || !b) {
		return false
	}
	return a.getFullYear() === b.getFullYear()
		&& a.getMonth() === b.getMonth()
		&& a.getDate() === b.getDate()
}

export function isDayBetween(date: Date, start: Date, end: Date): boolean {
	const t = startOfDay(date).getTime()
	return t > startOfDay(start).getTime() && t < startOfDay(end).getTime()
}
