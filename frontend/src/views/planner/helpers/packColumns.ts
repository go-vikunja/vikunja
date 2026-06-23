export interface PackedItem<T> {
	item: T
	col: number
	cols: number
}

/**
 * Lays out overlapping intervals into side-by-side columns (Google-Calendar
 * style). Items are grouped into clusters of transitively overlapping
 * intervals; within a cluster each item gets the lowest free column index and
 * every item in that cluster shares the same total column count.
 *
 * `getStart`/`getEnd` return comparable numbers (e.g. minutes from midnight).
 * Intervals that merely touch (`a.end === b.start`) do not count as overlapping.
 */
export function packColumns<T>(
	items: T[],
	getStart: (item: T) => number,
	getEnd: (item: T) => number,
): PackedItem<T>[] {
	const sorted = [...items].sort((a, b) => getStart(a) - getStart(b) || getEnd(a) - getEnd(b))

	const result: PackedItem<T>[] = []
	let cluster: PackedItem<T>[] = []
	let clusterEnd = -Infinity
	let columnEnds: number[] = []

	const flush = () => {
		const cols = columnEnds.length
		cluster.forEach(packed => packed.cols = cols)
		result.push(...cluster)
		cluster = []
		columnEnds = []
		clusterEnd = -Infinity
	}

	for (const item of sorted) {
		const start = getStart(item)
		const end = getEnd(item)

		// A gap to everything placed so far closes the current cluster.
		if (start >= clusterEnd && cluster.length > 0) {
			flush()
		}

		let col = columnEnds.findIndex(colEnd => colEnd <= start)
		if (col === -1) {
			col = columnEnds.length
			columnEnds.push(end)
		} else {
			columnEnds[col] = end
		}

		cluster.push({item, col, cols: 1})
		clusterEnd = Math.max(clusterEnd, end)
	}

	if (cluster.length > 0) {
		flush()
	}

	return result
}
