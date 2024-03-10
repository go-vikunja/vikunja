import {describe, expect, it} from 'vitest'
import {transformFilterStringForApi, transformFilterStringFromApi} from '@/helpers/filters'

const nullTitleToIdResolver = (title: string) => null
const nullIdToTitleResolver = (id: number) => null
describe('Filter Transformation', () => {

	const fieldCases = {
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
	
	describe('For api', () => {
		for (const c in fieldCases) {
			it('should transform all filter params for ' + c + ' to snake_case', () => {
				const transformed = transformFilterStringForApi(c + ' = ipsum', nullTitleToIdResolver, nullTitleToIdResolver)

				expect(transformed).toBe(fieldCases[c] + ' = ipsum')
			})
		}

		it('should correctly resolve labels', () => {
			const transformed = transformFilterStringForApi(
				'labels = lorem',
				(title: string) => 1,
				nullTitleToIdResolver,
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
				nullTitleToIdResolver,
			)

			expect(transformed).toBe('labels = 1&& due_date = now && labels = 2')
		})

		it('should correctly resolve projects', () => {
			const transformed = transformFilterStringForApi(
				'project = lorem',
				nullTitleToIdResolver,
				(title: string) => 1,
			)

			expect(transformed).toBe('project = 1')
		})

		it('should correctly resolve multiple projects', () => {
			const transformed = transformFilterStringForApi(
				'project = lorem && dueDate = now || project = ipsum',
				nullTitleToIdResolver,
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

	describe('To API', () => {
		for (const c in fieldCases) {
			it('should transform all filter params for ' + c + ' to snake_case', () => {
				const transformed = transformFilterStringFromApi(fieldCases[c] + ' = ipsum', nullTitleToIdResolver, nullTitleToIdResolver)

				expect(transformed).toBe(c + ' = ipsum')
			})
		}

		it('should correctly resolve labels', () => {
			const transformed = transformFilterStringFromApi(
				'labels = 1',
				(id: number) => 'lorem',
				nullIdToTitleResolver,
			)

			expect(transformed).toBe('labels = lorem')
		})
		
		it('should correctly resolve multiple labels', () => {
			const transformed = transformFilterStringFromApi(
				'labels = 1 && due_date = now && labels = 2',
				(id: number) => {
					switch (id) {
						case 1:
							return 'lorem'
						case 2:
							return 'ipsum'
						default:
							return null
					}
				},
				nullIdToTitleResolver,
			)

			expect(transformed).toBe('labels = lorem&& dueDate = now && labels = ipsum')
		})

		it('should correctly resolve projects', () => {
			const transformed = transformFilterStringFromApi(
				'project = 1',
				nullIdToTitleResolver,
				(id: number) => 'lorem',
			)

			expect(transformed).toBe('project = lorem')
		})

		it('should correctly resolve multiple projects', () => {
			const transformed = transformFilterStringFromApi(
				'project = lorem && due_date = now || project = ipsum',
				nullIdToTitleResolver,
				(id: number) => {
					switch (id) {
						case 1:
							return 'lorem'
						case 2:
							return 'ipsum'
						default:
							return null
					}
				},
			)

			expect(transformed).toBe('project = lorem && dueDate = now || project = ipsum')
		})
	})
})
