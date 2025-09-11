export function roundToNaturalDayBoundary(date: Date, isStart = false): Date {
	const d = new Date(date)
	if (isStart || d.getHours() < 12) {
		d.setHours(0, 0, 0, 0)
	} else {
		d.setHours(23, 59, 59, 999)
	}
	return d
}
