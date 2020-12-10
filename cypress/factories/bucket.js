import faker from 'faker'
import {Factory} from '../support/factory'
import {formatISO} from 'date-fns'

export class BucketFactory extends Factory {
	static table = 'buckets'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			list_id: 1,
			created_by_id: 1,
			created: formatISO(now),
			updated: formatISO(now)
		}
	}
}
