import faker from 'faker'
import {Factory} from '../support/factory'
import {formatISO} from 'date-fns'

export class NamespaceFactory extends Factory {
	static table = 'namespaces'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			owner_id: 1,
			created: formatISO(now),
			updated: formatISO(now)
		}
	}
}
