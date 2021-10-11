import faker from 'faker'

import {Factory} from '../support/factory'
import {formatISO} from "date-fns"

export class UserFactory extends Factory {
	static table = 'users'

	static factory() {
		const now = new Date()

		return {
            id: '{increment}',
            username: faker.lorem.word(10) + faker.datatype.uuid(),
            password: '$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.', // 1234
			status: 0,
			created: formatISO(now),
			updated: formatISO(now)
		}
	}
}