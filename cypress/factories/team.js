import faker from 'faker'
import {Factory} from '../support/factory'
import {formatISO} from 'date-fns'

export class TeamFactory extends Factory {
    static table = 'teams'

    static factory() {
        const now = new Date()

        return {
            name: faker.lorem.words(3),
            created_by_id: 1,
			created: formatISO(now),
			updated: formatISO(now)
        }
    }
}
