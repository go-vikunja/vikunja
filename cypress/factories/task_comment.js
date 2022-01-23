import faker from '@faker-js/faker'

import {Factory} from '../support/factory'
import {formatISO} from "date-fns"

export class TaskCommentFactory extends Factory {
	static table = 'task_comments'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			comment: faker.lorem.text(3),
			author_id: 1,
			task_id: 1,
			created: formatISO(now),
			updated: formatISO(now)
		}
	}
}
