import {faker} from '@faker-js/faker'

import {Factory} from '../support/factory'

export class LabelFactory extends Factory {
	static table = 'labels'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			title: faker.lorem.words(2),
			description: faker.lorem.text(10),
			hex_color: (Math.random()*0xFFFFFF<<0).toString(16), // random 6-digit hex number
			created_by_id: 1,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
