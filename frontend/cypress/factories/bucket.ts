import {faker} from '@faker-js/faker'
import {Factory} from '../support/factory'

export class BucketFactory extends Factory {
	static table = 'buckets'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(3),
			project_view_id: '{increment}',
			created_by_id: 1,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
