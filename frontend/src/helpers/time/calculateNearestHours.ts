export function calculateNearestHours(currentDate: Date = new Date()): number {
	if (currentDate.getHours() <= 9 || currentDate.getHours() > 21) {
		return 9
	}

	if (currentDate.getHours() <= 12) {
		return 12
	}

	if (currentDate.getHours() <= 15) {
		return 15
	}

	if (currentDate.getHours() <= 18) {
		return 18
	}

	if (currentDate.getHours() <= 21) {
		return 21
	}
	
	// Same case as in the first if, will never be called
	return 9
}
