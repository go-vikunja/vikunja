import {describe, expect, it} from 'vitest'
import {emptySimpleFilter, fromQuery, toQuery, type SimpleFilter} from './simpleFilter'

describe('simple filters', () => {
	it.each<[Partial<SimpleFilter>, string]>([
		[{}, ''],
		[{showDone: false}, 'done = false'],
		[{labelIds: [2, 1]}, 'labels in 2, 1'],
		[{assignees: ['alice', 'bob']}, 'assignees in alice, bob'],
		[{projectIds: [3, 4]}, 'project in 3, 4'],
		[{minPriority: 3}, 'priority >= 3'],
		[{due: {
			range: null,
			includeOverdue: true,
		}}, 'due_date < now'],
		[{due: {
			range: {
				from: 'now/d',
				to: 'now/d+1d',
			},
			includeOverdue: false,
		}},
		'due_date > now/d && due_date < now/d+1d'],
		[{due: {
			range: {
				from: '2026-11-01T00:00:00.000Z',
				to: '2026-11-02T00:00:00.000Z',
			},
			includeOverdue: false,
		}},
		'due_date > "2026-11-01T00:00:00.000Z" && due_date < "2026-11-02T00:00:00.000Z"'],
		[{
			showDone: false,
			labelIds: [2, 1],
			assignees: ['alice'],
			projectIds: [4],
			minPriority: 5,
		}, 'done = false && labels in 2, 1 && assignees in alice && project in 4 && priority >= 5'],
	])('generates and round trips %j', (fields, query) => {
		const form = {
			...emptySimpleFilter(),
			...fields,
		}
		expect(toQuery(form)).toBe(query)
		expect(fromQuery(query)).toEqual(form)
	})

	it('restores only the effective upper bound when including overdue', () => {
		const due = {
			range: {
				from: 'now/d',
				to: 'now/d+1d',
			},
			includeOverdue: true,
		}
		expect(toQuery({
			...emptySimpleFilter(),
			due,
		})).toBe('due_date < now/d+1d')
		expect(fromQuery('due_date < now/d+1d')?.due).toEqual({
			range: {
				from: '',
				to: 'now/d+1d',
			},
			includeOverdue: true,
		})
	})
	it.each([
		'done = false || priority >= 3', 'priority > 3', 'priority >= 3 && done = false',
		'labels in 01', 'priority >= 6', 'due_date < "2026-99-01T00:00:00.000Z"', 'due_date < "now"',
		'due_date > now', 'assignees in "a||b"',
	])('rejects noncanonical query %s', query => {
		expect(fromQuery(query)).toBeNull()
	})
})
