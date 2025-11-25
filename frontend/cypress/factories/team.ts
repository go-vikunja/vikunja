import {faker} from '@faker-js/faker'
import {Factory} from '../support/factory'

export class TeamFactory extends Factory {
    static table = 'teams'

    static factory() {
        const now = new Date()

        return {
            name: faker.lorem.words(3),
            created_by_id: 1,
			created: now.toISOString(),
			updated: now.toISOString(),
        }
    }
}
