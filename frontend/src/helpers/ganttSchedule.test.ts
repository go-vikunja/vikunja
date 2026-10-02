import {describe, expect, it} from 'vitest'

import type {TaskResponse} from '@/client/queries/tasks'
import {buildWbs, criticalPath, criticalPathInput, durationInDays, edgeKey, taskSpan} from './ganttSchedule'

function node(task: Partial<TaskResponse>, derived: {start?: Date, end?: Date} = {}) {
	return {
		task: {id: 1, done: false, related_tasks: {}, ...task} as TaskResponse,
		derivedStartDate: derived.start ?? null,
		derivedEndDate: derived.end ?? null,
		hasDerivedDates: Boolean(derived.start || derived.end),
	}
}

describe('taskSpan', () => {
	const d = (day: number) => new Date(2026, 10, day)

	it('own dates win, the due date stands in for a missing end', () => {
		expect(taskSpan(node({start_date: d(2).toISOString(), end_date: d(5).toISOString()}))).toEqual({start: d(2), end: d(5)})
		expect(taskSpan(node({start_date: d(2).toISOString(), due_date: d(6).toISOString()}))).toEqual({start: d(2), end: d(6)})
		expect(taskSpan(node({start_date: d(2).toISOString(), end_date: d(5).toISOString(), due_date: d(9).toISOString()})).end).toEqual(d(5))
	})

	it('a parent without dates spans its children, one with dates keeps them', () => {
		expect(taskSpan(node({}, {start: d(3), end: d(8)}))).toEqual({start: d(3), end: d(8)})
		expect(taskSpan(node({start_date: d(1).toISOString(), end_date: d(2).toISOString()}, {start: d(3), end: d(8)}))).toEqual({start: d(1), end: d(2)})
	})

	it('no dates at all is no span', () => {
		expect(taskSpan(node({}))).toEqual({start: null, end: null})
	})

	it('a milestone is one point: its end, else its due date, else its start, else the children', () => {
		expect(taskSpan(node({is_milestone: true, end_date: d(4).toISOString(), start_date: d(1).toISOString()}))).toEqual({start: d(4), end: d(4)})
		expect(taskSpan(node({is_milestone: true, due_date: d(5).toISOString()}))).toEqual({start: d(5), end: d(5)})
		expect(taskSpan(node({is_milestone: true, start_date: d(6).toISOString()}))).toEqual({start: d(6), end: d(6)})
		expect(taskSpan(node({is_milestone: true}, {end: d(7)}))).toEqual({start: d(7), end: d(7)})
		expect(taskSpan(node({is_milestone: true}))).toEqual({start: null, end: null})
	})
})

describe('criticalPathInput', () => {
	const d = (day: number) => new Date(2026, 10, day)
	const task = (id: number, o: Partial<TaskResponse> = {}) => ({id, done: false, related_tasks: {}, ...o} as TaskResponse)

	it('keeps dated tasks that are not done, and every precedes link', () => {
		const a = task(1, {related_tasks: {precedes: [task(2)]} as TaskResponse['related_tasks']})
		const b = task(2)
		const done = task(3, {done: true})
		const undated = task(4)
		const rows = [
			{task: a, span: {start: d(1), end: d(3)}},
			{task: b, span: {start: d(3), end: d(5)}},
			{task: done, span: {start: d(1), end: d(2)}},
			{task: undated, span: {start: null, end: null}},
			{task: task(5), span: {start: d(1), end: null}},
		]

		const input = criticalPathInput(rows, [a, b, done, undated])

		expect(input.tasks.map(t => t.id)).toEqual([1, 2])
		expect(input.tasks[0]).toEqual({id: 1, start: d(1).getTime(), end: d(3).getTime()})
		expect(input.edges).toEqual([{from: 1, to: 2}])
	})

	it('a dateless parent that spans its children is part of the path', () => {
		const parent = node({id: 10}, {start: d(2), end: d(6)})
		const input = criticalPathInput([{task: parent.task, span: taskSpan(parent)}], [parent.task])
		expect(input.tasks).toEqual([{id: 10, start: d(2).getTime(), end: d(6).getTime()}])
	})
})

const DAY = 24 * 60 * 60 * 1000
const t = (id: number, startDay: number, endDay: number) => ({id, start: startDay * DAY, end: endDay * DAY})

describe('buildWbs', () => {
	const node = (id: number, indentLevel: number) => ({task: {id}, indentLevel})

	it('numbers a flat list and a nested tree', () => {
		const wbs = buildWbs([node(1, 0), node(2, 1), node(3, 1), node(4, 2), node(5, 0), node(6, 1)])
		expect(wbs.get(1)).toBe('1')
		expect(wbs.get(2)).toBe('1.1')
		expect(wbs.get(3)).toBe('1.2')
		expect(wbs.get(4)).toBe('1.2.1')
		expect(wbs.get(5)).toBe('2')
		expect(wbs.get(6)).toBe('2.1')
	})

	it('restarts the counter of a level under a new parent', () => {
		const wbs = buildWbs([node(1, 0), node(2, 1), node(3, 0), node(4, 1)])
		expect(wbs.get(2)).toBe('1.1')
		expect(wbs.get(4)).toBe('2.1')
	})

	it('empty in, empty out', () => {
		expect(buildWbs([]).size).toBe(0)
	})
})

describe('durationInDays', () => {
	it('counts both days', () => {
		expect(durationInDays(new Date(2026, 10, 2), new Date(2026, 10, 2))).toBe(1)
		expect(durationInDays(new Date(2026, 10, 2), new Date(2026, 10, 6))).toBe(5)
	})

	it('ignores the time of day', () => {
		expect(durationInDays(new Date(2026, 10, 2, 23, 0), new Date(2026, 10, 3, 1, 0))).toBe(2)
	})

	it('milestones have none, missing dates have none', () => {
		expect(durationInDays(new Date(2026, 10, 2), new Date(2026, 10, 6), true)).toBe(0)
		expect(durationInDays(null, new Date(2026, 10, 6))).toBeNull()
		expect(durationInDays(new Date(2026, 10, 2), null)).toBeNull()
	})

	it('an end before the start is still at least one day', () => {
		expect(durationInDays(new Date(2026, 10, 6), new Date(2026, 10, 2))).toBe(1)
	})
})

describe('criticalPath', () => {
	it('a contiguous chain is critical, a task with slack and an unlinked task are not', () => {
		// A(0-3) -> B(3-6) -> C(6-9); D(0-2) -> C has a gap of 4 days; E has no links
		const tasks = [t(1, 0, 3), t(2, 3, 6), t(3, 6, 9), t(4, 0, 2), t(5, 0, 9)]
		const edges = [{from: 1, to: 2}, {from: 2, to: 3}, {from: 4, to: 3}]

		const res = criticalPath(tasks, edges)

		expect([...res.taskIds].sort()).toEqual([1, 2, 3])
		expect(res.edges.has(edgeKey(1, 2))).toBe(true)
		expect(res.edges.has(edgeKey(2, 3))).toBe(true)
		expect(res.edges.has(edgeKey(4, 3))).toBe(false)
		expect(res.taskIds.has(5)).toBe(false)
	})

	it('a gap on the chain breaks it: only what is tight to the finish is critical', () => {
		// A(0-3) -> B(5-8): A has 2 days of slack
		const res = criticalPath([t(1, 0, 3), t(2, 5, 8)], [{from: 1, to: 2}])

		expect([...res.taskIds]).toEqual([2])
		expect(res.edges.size).toBe(0)
	})

	it('of two branches the longer one is critical', () => {
		// A(0-2) -> B(2-4) -> D(4-6); A -> C(2-3) -> D
		const tasks = [t(1, 0, 2), t(2, 2, 4), t(3, 2, 3), t(4, 4, 6)]
		const res = criticalPath(tasks, [{from: 1, to: 2}, {from: 1, to: 3}, {from: 2, to: 4}, {from: 3, to: 4}])

		expect([...res.taskIds].sort()).toEqual([1, 2, 4])
		expect(res.taskIds.has(3)).toBe(false)
	})

	it('a violated dependency (successor starts before the predecessor ends) is critical', () => {
		const res = criticalPath([t(1, 0, 5), t(2, 3, 6)], [{from: 1, to: 2}])
		expect(res.taskIds.has(1)).toBe(true)
		expect(res.taskIds.has(2)).toBe(true)
		expect(res.edges.has(edgeKey(1, 2))).toBe(true)
	})

	it('tolerates a minute of rounding between end and start', () => {
		const a = {id: 1, start: 0, end: 3 * DAY - 30_000}
		const b = {id: 2, start: 3 * DAY, end: 6 * DAY}
		const res = criticalPath([a, b], [{from: 1, to: 2}])
		expect(res.taskIds.has(1)).toBe(true)
	})

	it('cycles do not hang and are left out', () => {
		const res = criticalPath([t(1, 0, 2), t(2, 2, 4), t(3, 4, 6)], [{from: 1, to: 2}, {from: 2, to: 1}, {from: 2, to: 3}])
		expect(res.taskIds.has(1)).toBe(false)
		expect(res.taskIds.has(2)).toBe(false)
	})

	it('ignores self links, duplicates and links to tasks that are not loaded', () => {
		const res = criticalPath([t(1, 0, 2), t(2, 2, 4)], [{from: 1, to: 1}, {from: 1, to: 2}, {from: 1, to: 2}, {from: 1, to: 99}])
		expect([...res.taskIds].sort()).toEqual([1, 2])
	})

	it('no dependencies, nothing critical', () => {
		const res = criticalPath([t(1, 0, 2), t(2, 1, 5)], [])
		expect(res.taskIds.size).toBe(0)
		expect(res.edges.size).toBe(0)
	})
})
