// Minimum spacing between positions. Must survive JSON round-trip.
// Matches backend MinPositionSpacing constant.
const MIN_POSITION_SPACING = 0.01

// Room left between positions when appending, so later inserts can be placed in
// between without a recalculation. Matches the backend's default position formula.
export const POSITION_SPACING = Math.pow(2, 16)

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
		return positionBefore + POSITION_SPACING
	}

	// If we have both a task before and after it, we actually calculate the position
	return positionBefore + (positionAfter - positionBefore) / 2
}

/**
 * Spreads `count` positions evenly over the gap between two neighbors, in ascending
 * order. Needed when inserting several items at once: giving them all the same
 * position leaves their final order up to the api's conflict repair.
 *
 * For count = 1 this returns the same value as calculateItemPosition.
 */
export const calculateItemPositions = (
	count: number,
	positionBefore: number | null = null,
	positionAfter: number | null = null,
): number[] => {
	const spread = (base: number, spacing: number) => Array.from(
		{length: Math.max(count, 0)},
		(_, i) => base + spacing * (i + 1),
	)

	if (positionAfter === null || positionBefore === positionAfter) {
		const base = positionBefore ?? 0
		return spread(base, positionAfter === null ? POSITION_SPACING : MIN_POSITION_SPACING)
	}

	const base = positionBefore ?? 0
	const spacing = (positionAfter - base) / (count + 1)

	// The gap is too small to subdivide without the values colliding after the JSON
	// round-trip. Spread past positionAfter on purpose so the batch doesn't collide with it.
	if (spacing < MIN_POSITION_SPACING) {
		return spread(Math.max(base, positionAfter), MIN_POSITION_SPACING)
	}

	return spread(base, spacing)
}
