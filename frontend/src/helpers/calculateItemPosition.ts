// Minimum spacing between positions. Must survive JSON round-trip.
// Matches backend MinPositionSpacing constant.
const MIN_POSITION_SPACING = 0.01

export const calculateItemPosition = (
	positionBefore: number | null = null,
	positionAfter: number | null = null,
): number => {
	// Both neighbors have the same position (conflict)
	if (positionBefore !== null && positionAfter !== null && positionBefore === positionAfter) {
		// Nudge slightly above to maintain ordering intent
		return positionAfter + MIN_POSITION_SPACING
	}

	if (positionBefore === null) {
		if (positionAfter === null) {
			return 0
		}

		// If there is no task before it, place it at half the position of the task after
		return positionAfter / 2
	}

	// If there is no task after it, we just add 2^16 to the last position to have enough room in the future
	if (positionAfter === null) {
		return positionBefore + Math.pow(2, 16)
	}

	// If we have both a task before and after it, we actually calculate the position
	return positionBefore + (positionAfter - positionBefore) / 2
}
