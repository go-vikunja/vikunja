import {faker} from '@faker-js/faker'
import {Factory} from '../support/factory'

export class NamespaceFactory extends Factory {
	static table = 'namespaces'

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
