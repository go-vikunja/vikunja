import {describe, expect, it} from 'vitest'
import {getRowReorder, type ReorderRow} from './ganttRowReorder'

// A, B (with subtask B1), C
const ROWS: ReorderRow[] = [
	{taskId: 1, parentId: null, indentLevel: 0, position: 100},
	{taskId: 2, parentId: null, indentLevel: 0, position: 200},
	{taskId: 3, parentId: 2, indentLevel: 1, position: 50},
	{taskId: 4, parentId: null, indentLevel: 0, position: 300},
]

describe('getRowReorder', () => {
	it('moves a row to the top', () => {
		expect(getRowReorder(ROWS, 3, 0)).toEqual({position: 50, indicatorRow: 0})
	})

	it('moves a row to the bottom', () => {
		expect(getRowReorder(ROWS, 0, 4)).toEqual({position: 300 + 2 ** 16, indicatorRow: 4})
	})

	it('moves a row between two siblings', () => {
		expect(getRowReorder(ROWS, 3, 1)).toEqual({position: 150, indicatorRow: 1})
	})

	it('drops after a sibling together with its subtasks', () => {
		expect(getRowReorder(ROWS, 0, 2)).toEqual({position: 250, indicatorRow: 3})
	})

	it('returns null when the row would stay where it is', () => {
		expect(getRowReorder(ROWS, 0, 0)).toBeNull()
		expect(getRowReorder(ROWS, 0, 1)).toBeNull()
		expect(getRowReorder(ROWS, 1, 3)).toBeNull()
	})

	it('returns null for an only child', () => {
		expect(getRowReorder(ROWS, 2, 0)).toBeNull()
	})

	it('keeps a subtask among its siblings', () => {
		const rows: ReorderRow[] = [
			{taskId: 1, parentId: null, indentLevel: 0, position: 100},
			{taskId: 2, parentId: 1, indentLevel: 1, position: 10},
			{taskId: 3, parentId: 1, indentLevel: 1, position: 20},
			{taskId: 4, parentId: null, indentLevel: 0, position: 200},
		]
		expect(getRowReorder(rows, 2, 0)).toEqual({position: 5, indicatorRow: 1})
		expect(getRowReorder(rows, 1, 4)).toEqual({position: 20 + 2 ** 16, indicatorRow: 3})
	})
})
