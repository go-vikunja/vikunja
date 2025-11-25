import {Factory} from '../support/factory'
import {faker} from '@faker-js/faker'

export class LinkShareFactory extends Factory {
	static table = 'link_shares'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			hash: faker.lorem.word(32),
			project_id: 1,
			permission: 0,
			sharing_type: 0,
			shared_by_id: 1,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
