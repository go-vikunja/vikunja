import {faker} from '@faker-js/faker'
import {Factory} from '../support/factory'

export class TaskFactory extends Factory {
	static table = 'tasks'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			done: false,
			project_id: 1,
			created_by_id: 1,
			index: '{increment}',
			created: now.toISOString(),
			updated: now.toISOString()
		}
	}
}
