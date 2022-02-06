export function getNextWeekDate(): Date {
	return new Date((new Date()).getTime() + 7 * 24 * 60 * 60 * 1000)
}
