import {Factory} from '../support/factory'

export class SavedFilterFactory extends Factory {
	static table = 'saved_filters'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: 'Test Filter',
			description: '',
			filters: '{"filter":"","filter_include_nulls":false,"s":""}',
			owner_id: 1,
			is_favorite: false,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
