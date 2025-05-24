export const calculateItemPosition = (
	positionBefore: number | null = null,
	positionAfter: number | null = null,
): number => {
	if (positionBefore === null) {
		if (positionAfter === null) {
			return 0
		}
		
		// If there is no task after it, we just add 2^16 to the last position to have enough room in the future
		return positionAfter / 2
	}
	
	// If there is no task after it, we just add 2^16 to the last position to have enough room in the future
	if (positionAfter === null) {
		return positionBefore + Math.pow(2, 16)
	}
	
	// If we have both a task before and after it, we actually calculate the position
	return positionBefore + (positionAfter - positionBefore) / 2
}
