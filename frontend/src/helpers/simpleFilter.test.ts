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
			kind: 'before',
			to: 'now',
		}}, 'due_date < now'],
		[{due: {
			kind: 'between',
			from: 'now/Q',
			to: 'now/Q+1Q',
		}},
		'due_date > now/Q && due_date < now/Q+1Q'],
		[{due: {
			kind: 'between',
			from: '2026-11-01T00:00:00.000Z',
			to: '2026-11-02T00:00:00.000Z',
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

	it.each<[string, Partial<SimpleFilter>]>([
		['done = false && assignees = alice', {
			showDone: false,
			assignees: ['alice'],
		}],
		['labels = 1', {labelIds: [1]}],
		['project = 2', {projectIds: [2]}],
		['project_id in 2, 3', {projectIds: [2, 3]}],
	])('accepts equivalent query %s', (query, fields) => {
		expect(fromQuery(query)).toEqual({
			...emptySimpleFilter(),
			...fields,
		})
	})

	it.each([
		'done = false || priority >= 3', 'due_date < "2026-99-01T00:00:00.000Z"', 'priority >= 3 && done = false',
		'due_date > now', 'assignees in "a||b"', 'assignees in alice\\', 'labels = 1, 2',
	])('rejects noncanonical query %s', query => {
		expect(fromQuery(query)).toBeNull()
	})
})
