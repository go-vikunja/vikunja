// Splits items into the fewest rows of at most maxPerRow, as evenly as possible.
// Leftovers go to the top rows, so earlier rows hold at most one more item than later ones.
export function balancedRows<T>(items: T[], maxPerRow: number): T[][] {
	if (items.length === 0) {
		return []
	}

	const rowCount = Math.ceil(items.length / Math.max(1, maxPerRow))
	const baseSize = Math.floor(items.length / rowCount)
	const rowsWithExtra = items.length % rowCount

	const rows: T[][] = []
	let start = 0
	for (let i = 0; i < rowCount; i++) {
		const size = baseSize + (i < rowsWithExtra ? 1 : 0)
		rows.push(items.slice(start, start + size))
		start += size
	}
	return rows
}
