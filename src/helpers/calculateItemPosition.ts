export const calculateItemPosition = (positionBefore: number | null, positionAfter: number | null): number => {
	if (positionBefore === null && positionAfter === null) {
		return 0
	}

	// If there is no task before, our task is the first task in which case we let it have half of the position of the task after it
	if (positionBefore === null && positionAfter !== null) {
		return positionAfter / 2
	}
	
	// If there is no task after it, we just add 2^16 to the last position to have enough room in the future
	if (positionBefore !== null && positionAfter === null) {
		return positionBefore + Math.pow(2, 16)
	}
	
	// If we have both a task before and after it, we acually calculate the position
	// @ts-ignore - can never be null but TS does not seem to understand that
	return positionBefore + (positionAfter - positionBefore) / 2
}