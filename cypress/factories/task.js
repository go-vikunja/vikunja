import faker from 'faker'
import {Factory} from '../support/factory'
import {formatISO} from 'date-fns'

export class TaskFactory extends Factory {
	static table = 'tasks'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			done: false,
			list_id: 1,
			created_by_id: 1,
			is_favorite: false,
			index: '{increment}',
			created: formatISO(now),
			updated: formatISO(now)
		}
	}
}
