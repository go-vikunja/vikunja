import {describe, expect, it} from 'vitest'
import {buildGanttTaskTree} from './ganttTaskTree'
import type {Task as ITask} from '@/client/generated'
import {normalizeTask, type TaskResponse} from '@/client/queries/tasks'

function makeTask(id: number, overrides: Partial<ITask> = {}): TaskResponse {
	return normalizeTask({
		id,
		title: `Task ${id}`,
		start_date: new Date('2026-03-01').toISOString(),
		end_date: new Date('2026-03-10').toISOString(),
		due_date: undefined,
		done: false,
		related_tasks: {},
		...overrides,
	})
}

describe('buildGanttTaskTree', () => {
	it('returns flat list when no relations exist', () => {
		const tasks = new Map<number, TaskResponse>([
			[1, makeTask(1)],
			[2, makeTask(2)],
		])

		const result = buildGanttTaskTree(tasks)

		expect(result).toHaveLength(2)
		expect(result[0].task.id).toBe(1)
		expect(result[0].indentLevel).toBe(0)
		expect(result[0].isParent).toBe(false)
		expect(result[1].task.id).toBe(2)
		expect(result[1].indentLevel).toBe(0)
	})

	it('nests subtasks under parents in depth-first order', () => {
		const child1 = makeTask(2, {
			related_tasks: {parenttask: [makeTask(1)]},
		})
		const child2 = makeTask(3, {
			related_tasks: {parenttask: [makeTask(1)]},
		})
		const parent = makeTask(1, {
			related_tasks: {subtask: [makeTask(2), makeTask(3)]},
		})

		const tasks = new Map<number, TaskResponse>([
			[1, parent],
			[2, child1],
			[3, child2],
		])

		const result = buildGanttTaskTree(tasks)

		expect(result).toHaveLength(3)
		expect(result[0].task.id).toBe(1)
		expect(result[0].indentLevel).toBe(0)
		expect(result[0].isParent).toBe(true)
		expect(result[0].childIds).toEqual([2, 3])
		expect(result[1].task.id).toBe(2)
		expect(result[1].indentLevel).toBe(1)
		expect(result[2].task.id).toBe(3)
		expect(result[2].indentLevel).toBe(1)
	})

	it('handles multi-level nesting', () => {
		const grandchild = makeTask(3, {
			related_tasks: {parenttask: [makeTask(2)]},
		})
		const child = makeTask(2, {
			related_tasks: {
				parenttask: [makeTask(1)],
				subtask: [makeTask(3)],
			},
		})
		const parent = makeTask(1, {
			related_tasks: {subtask: [makeTask(2)]},
		})

		const tasks = new Map<number, TaskResponse>([
			[1, parent],
			[2, child],
			[3, grandchild],
		])

		const result = buildGanttTaskTree(tasks)

		expect(result).toHaveLength(3)
		expect(result[0].indentLevel).toBe(0) // parent
		expect(result[1].indentLevel).toBe(1) // child
		expect(result[1].isParent).toBe(true)
		expect(result[2].indentLevel).toBe(2) // grandchild
	})

	it('caps indent level at max depth', () => {
		// Build a chain: 1 -> 2 -> 3 -> 4 -> 5 -> 6
		const tasks = new Map<number, TaskResponse>()
		for (let i = 1; i <= 6; i++) {
			const related_tasks: ITask['related_tasks'] = {}
			if (i > 1) related_tasks.parenttask = [makeTask(i - 1)]
			if (i < 6) related_tasks.subtask = [makeTask(i + 1)]
			tasks.set(i, makeTask(i, {related_tasks}))
		}

		const result = buildGanttTaskTree(tasks)

		expect(result[4].indentLevel).toBe(4) // level 4 (0-indexed)
		expect(result[5].indentLevel).toBe(4) // capped at 4
	})

	it('calculates derived dates for dateless parents from children', () => {
		const child1 = makeTask(2, {
			start_date: new Date('2026-03-05').toISOString(),
			end_date: new Date('2026-03-10').toISOString(),
			related_tasks: {parenttask: [makeTask(1)]},
		})
		const child2 = makeTask(3, {
			start_date: new Date('2026-03-01').toISOString(),
			end_date: new Date('2026-03-15').toISOString(),
			related_tasks: {parenttask: [makeTask(1)]},
		})
		const parent = makeTask(1, {
			start_date: undefined,
			end_date: undefined,
			due_date: undefined,
			related_tasks: {subtask: [makeTask(2), makeTask(3)]},
		})

		const tasks = new Map<number, TaskResponse>([
			[1, parent],
			[2, child1],
			[3, child2],
		])

		const result = buildGanttTaskTree(tasks)

		expect(result[0].derivedStartDate?.toISOString()).toContain('2026-03-01')
		expect(result[0].derivedEndDate?.toISOString()).toContain('2026-03-15')
		expect(result[0].hasDerivedDates).toBe(true)
	})

	it('derives dates for diamond-shaped relation graphs in linear time', () => {
		// Every node links to both nodes of the next layer: 2^30 paths without memoization.
		const layers = 30
		const id = (layer: number, i: number) => layer * 2 + i + 1
		const tasks = new Map<number, TaskResponse>()
		tasks.set(0, makeTask(0, {
			start_date: undefined,
			end_date: undefined,
			related_tasks: {subtask: [makeTask(id(0, 0)), makeTask(id(0, 1))]},
		}))
		for (let layer = 0; layer < layers; layer++) {
			for (let i = 0; i < 2; i++) {
				const isLast = layer === layers - 1
				const parents = layer === 0
					? [makeTask(0)]
					: [makeTask(id(layer - 1, 0)), makeTask(id(layer - 1, 1))]
				tasks.set(id(layer, i), makeTask(id(layer, i), {
					start_date: isLast ? new Date(`2026-03-0${i + 1}`).toISOString() : undefined,
					end_date: isLast ? new Date(`2026-03-1${i + 1}`).toISOString() : undefined,
					related_tasks: {
						parenttask: parents,
						subtask: isLast ? [] : [makeTask(id(layer + 1, 0)), makeTask(id(layer + 1, 1))],
					},
				}))
			}
		}

		const start = performance.now()
		const result = buildGanttTaskTree(tasks)
		expect(performance.now() - start).toBeLessThan(1000)

		expect(result).toHaveLength(tasks.size)
		expect(result[0].derivedStartDate?.toISOString()).toContain('2026-03-01')
		expect(result[0].derivedEndDate?.toISOString()).toContain('2026-03-12')
	})

	it('does not loop forever on cyclic subtask relations', () => {
		const tasks = new Map<number, TaskResponse>([
			[1, makeTask(1, {
				start_date: undefined,
				end_date: undefined,
				related_tasks: {subtask: [makeTask(2)]},
			})],
			[2, makeTask(2, {
				start_date: undefined,
				end_date: undefined,
				related_tasks: {parenttask: [makeTask(1), makeTask(3)], subtask: [makeTask(3)]},
			})],
			[3, makeTask(3, {
				start_date: new Date('2026-03-05').toISOString(),
				end_date: new Date('2026-03-20').toISOString(),
				related_tasks: {parenttask: [makeTask(2)], subtask: [makeTask(2)]},
			})],
		])

		const result = buildGanttTaskTree(tasks)

		expect(result).toHaveLength(3)
		expect(result[0].derivedStartDate?.toISOString()).toContain('2026-03-05')
		expect(result[0].derivedEndDate?.toISOString()).toContain('2026-03-20')
	})
})
