export function calculateNearestHours(currentDate: Date = new Date()): number {
	const hours = currentDate.getHours()
	const minutes = currentDate.getMinutes()
	
	// Helper to check if current time is before a given hour breakpoint
	// Returns true if we're before the hour, or at the hour with 0 minutes
	const isBeforeOrAt = (breakpoint: number): boolean => {
		return hours < breakpoint || (hours === breakpoint && minutes === 0)
	}
	
	if (isBeforeOrAt(9) || hours > 21) {
		return 9
	}

	if (isBeforeOrAt(12)) {
		return 12
	}

	if (isBeforeOrAt(15)) {
		return 15
	}

	if (isBeforeOrAt(18)) {
		return 18
	}

	if (isBeforeOrAt(21)) {
		return 21
	}
	
	// After 21:00 with minutes > 0, return 9 for next day
	return 9
}
