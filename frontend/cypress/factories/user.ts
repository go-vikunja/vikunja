import {faker} from '@faker-js/faker'

import {Factory} from '../support/factory'

export class UserFactory extends Factory {
	static table = 'users'

	static factory() {
		const now = new Date()

		return {
            id: '{increment}',
            username: faker.lorem.word(10) + faker.string.uuid(),
            password: '$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.', // 1234
			status: 0,
			issuer: 'local',
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}