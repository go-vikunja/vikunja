import {calculateItemPosition} from '@/helpers/calculateItemPosition'

export interface ReorderRow {
	taskId: number
	parentId: number | null
	indentLevel: number
	position: number
}

export interface RowReorder {
	position: number
	// Row the drop line is drawn above; rows.length means below the last row.
	indicatorRow: number
}

interface Block {
	start: number
	end: number
	position: number
}

function blockAt(rows: ReorderRow[], start: number): Block {
	let end = start + 1
	while (end < rows.length && rows[end].indentLevel > rows[start].indentLevel) {
		end++
	}
	return {start, end, position: rows[start].position}
}

/**
 * Where a dragged row lands when dropped at `slot` (a row boundary, 0..rows.length) and the
 * position that puts it there. A row only moves among the rows sharing its parent, together with
 * its subtasks. Null when the drop would not change the order.
 */
export function getRowReorder(rows: ReorderRow[], draggedIndex: number, slot: number): RowReorder | null {
	const dragged = rows[draggedIndex]
	if (!dragged) return null

	const siblings: Block[] = []
	let currentIndex = -1
	rows.forEach((row, index) => {
		if (row.parentId !== dragged.parentId || row.indentLevel !== dragged.indentLevel) return
		if (index === draggedIndex) {
			currentIndex = siblings.length
			return
		}
		siblings.push(blockAt(rows, index))
	})

	const insertAt = siblings.filter(block => block.start < slot).length
	if (insertAt === currentIndex) return null

	const before = siblings[insertAt - 1]
	const after = siblings[insertAt]
	if (!before && !after) return null

	return {
		position: calculateItemPosition(before?.position ?? null, after?.position ?? null),
		indicatorRow: after ? after.start : before!.end,
	}
}
