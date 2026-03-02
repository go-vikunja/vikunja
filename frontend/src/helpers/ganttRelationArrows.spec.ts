import {describe, expect, it} from 'vitest'
import {buildRelationArrows, type GanttBarPosition} from './ganttRelationArrows'
import type {ITask} from '@/modelTypes/ITask'

function makeTask(id: number, overrides: Partial<ITask> = {}): ITask {
	return {
		id,
		title: `Task ${id}`,
		relatedTasks: {},
		...overrides,
	} as ITask
}

describe('buildRelationArrows', () => {
	it('returns empty array when no dependency relations exist', () => {
		const tasks = new Map<number, ITask>([
			[1, makeTask(1)],
			[2, makeTask(2)],
		])
		const positions = new Map<number, GanttBarPosition>([
			[1, {x: 0, y: 20, width: 100, rowIndex: 0}],
			[2, {x: 50, y: 60, width: 100, rowIndex: 1}],
		])

		const result = buildRelationArrows(tasks, positions, new Map())
		expect(result).toHaveLength(0)
	})

	it('creates arrow for blocking relation', () => {
		const tasks = new Map<number, ITask>([
			[1, makeTask(1, {relatedTasks: {blocking: [makeTask(2)]}})],
			[2, makeTask(2, {relatedTasks: {blocked: [makeTask(1)]}})],
		])
		const positions = new Map<number, GanttBarPosition>([
			[1, {x: 0, y: 20, width: 100, rowIndex: 0}],
			[2, {x: 150, y: 60, width: 100, rowIndex: 1}],
		])

		const result = buildRelationArrows(tasks, positions, new Map())

		expect(result).toHaveLength(1)
		expect(result[0].fromTaskId).toBe(1)
		expect(result[0].toTaskId).toBe(2)
		expect(result[0].startX).toBe(100) // x + width
		expect(result[0].endX).toBe(150)   // target x
		expect(result[0].color).toBe('var(--danger)')
		expect(result[0].relationKind).toBe('blocking')
	})

	it('creates arrow for precedes relation', () => {
		const tasks = new Map<number, ITask>([
			[1, makeTask(1, {relatedTasks: {precedes: [makeTask(2)]}})],
			[2, makeTask(2, {relatedTasks: {follows: [makeTask(1)]}})],
		])
		const positions = new Map<number, GanttBarPosition>([
			[1, {x: 0, y: 20, width: 100, rowIndex: 0}],
			[2, {x: 150, y: 60, width: 100, rowIndex: 1}],
		])

		const result = buildRelationArrows(tasks, positions, new Map())

		expect(result).toHaveLength(1)
		expect(result[0].relationKind).toBe('precedes')
		expect(result[0].color).toBe('var(--grey-500)')
	})

	it('skips arrows when target task is not visible', () => {
		const tasks = new Map<number, ITask>([
			[1, makeTask(1, {relatedTasks: {blocking: [makeTask(99)]}})],
		])
		const positions = new Map<number, GanttBarPosition>([
			[1, {x: 0, y: 20, width: 100, rowIndex: 0}],
		])

		const result = buildRelationArrows(tasks, positions, new Map())
		expect(result).toHaveLength(0)
	})

	it('re-routes arrows to parent when child is collapsed', () => {
		const tasks = new Map<number, ITask>([
			[1, makeTask(1, {relatedTasks: {blocking: [makeTask(3)]}})],
			[2, makeTask(2)], // parent of task 3
			[3, makeTask(3, {relatedTasks: {blocked: [makeTask(1)]}})],
		])
		const positions = new Map<number, GanttBarPosition>([
			[1, {x: 0, y: 20, width: 100, rowIndex: 0}],
			[2, {x: 50, y: 60, width: 200, rowIndex: 1}],
			// task 3 has no position (collapsed)
		])
		const hiddenToAncestor = new Map<number, number>([
			[3, 2],
		])

		const result = buildRelationArrows(tasks, positions, hiddenToAncestor)

		expect(result).toHaveLength(1)
		expect(result[0].toTaskId).toBe(2) // re-routed to parent
		expect(result[0].endX).toBe(50)    // parent's x
	})

	it('deduplicates bidirectional relations', () => {
		const tasks = new Map<number, ITask>([
			[1, makeTask(1, {relatedTasks: {blocking: [makeTask(2)]}})],
			[2, makeTask(2, {relatedTasks: {blocked: [makeTask(1)]}})],
		])
		const positions = new Map<number, GanttBarPosition>([
			[1, {x: 0, y: 20, width: 100, rowIndex: 0}],
			[2, {x: 150, y: 60, width: 100, rowIndex: 1}],
		])

		const result = buildRelationArrows(tasks, positions, new Map())

		// Only one arrow, not two
		expect(result).toHaveLength(1)
	})
})

