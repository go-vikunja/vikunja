import {Factory} from '../support/factory'
import {formatISO} from "date-fns"
import faker from '@faker-js/faker'

export class ListFactory extends Factory {
	static table = 'lists'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			owner_id: 1,
			namespace_id: 1,
			created: formatISO(now),
			updated: formatISO(now)
		}
	}
}