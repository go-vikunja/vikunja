import {Factory} from '../support/factory'
import {formatISO} from "date-fns"
import faker from 'faker'

export class LinkShareFactory extends Factory {
	static table = 'link_shares'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			hash: faker.random.word(32),
			list_id: 1,
			right: 0,
			sharing_type: 0,
			shared_by_id: 1,
			created: formatISO(now),
			updated: formatISO(now)
		}
	}
}
