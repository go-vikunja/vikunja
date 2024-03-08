import {describe, expect, it} from 'vitest'
import {transformFilterStringForApi} from '@/helpers/filters'

const nullResolver = (title: string) => null
describe('Filter Transformation', () => {

	const cases = {
		'done': 'done',
		'priority': 'priority',
		'percentDone': 'percent_done',
		'dueDate': 'due_date',
		'startDate': 'start_date',
		'endDate': 'end_date',
		'doneAt': 'done_at',
		'reminders': 'reminders',
		'assignees': 'assignees',
		'labels': 'labels',
	}

	for (const c in cases) {
		it('should transform all filter params for ' + c + ' to snake_case', () => {
			const transformed = transformFilterStringForApi(c + ' = ipsum', nullResolver, nullResolver)

			expect(transformed).toBe(cases[c] + ' = ipsum')
		})
	}
	
	it('should correctly resolve labels', () => {
		const transformed = transformFilterStringForApi(
			'labels = lorem',
			(title: string) => 1,
			nullResolver,
		)
		
		expect(transformed).toBe('labels = 1')
	})

	it('should correctly resolve multiple labels', () => {
		const transformed = transformFilterStringForApi(
			'labels = lorem && dueDate = now && labels = ipsum',
			(title: string) => {
				switch (title) {
					case 'lorem':
						return 1
					case 'ipsum':
						return 2
					default:
						return null
				}
			},
			nullResolver,
		)

		expect(transformed).toBe('labels = 1&& due_date = now && labels = 2')
	})

	it('should correctly resolve projects', () => {
		const transformed = transformFilterStringForApi(
			'project = lorem',
			nullResolver,
			(title: string) => 1,
		)

		expect(transformed).toBe('project = 1')
	})
	
	it('should correctly resolve multiple projects', () => {
		const transformed = transformFilterStringForApi(
			'project = lorem && dueDate = now || project = ipsum',
			nullResolver,
			(title: string) => {
				switch (title) {
					case 'lorem':
						return 1
					case 'ipsum':
						return 2
					default:
						return null
				}
			},
		)

		expect(transformed).toBe('project = 1&& due_date = now || project = 2')
	})
})
