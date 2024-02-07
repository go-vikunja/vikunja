import {Factory} from '../support/factory'
import {faker} from '@faker-js/faker'

export class ProjectFactory extends Factory {
	static table = 'projects'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			owner_id: 1,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}