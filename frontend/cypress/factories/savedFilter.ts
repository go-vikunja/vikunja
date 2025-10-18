import {Factory} from '../support/factory'
import {faker} from '@faker-js/faker'

export interface SavedFilterAttributes {
	id: number | '{increment}'
	title: string
	description?: string
	filters: {
		sort_by: string[]
		order_by: string[]
		filter: string
		filter_include_nulls: boolean
		s: string
	}
	owner_id?: number
	is_favorite?: boolean
	created?: string
	updated?: string
}

export class SavedFilterFactory extends Factory {
	static table = 'saved_filters'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			description: '',
			filters: JSON.stringify({
				sort_by: ['due_date', 'id'],
				order_by: ['asc', 'desc'],
				filter: '',
				filter_include_nulls: false,
				s: '',
			}),
			owner_id: 1,
			is_favorite: false,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
