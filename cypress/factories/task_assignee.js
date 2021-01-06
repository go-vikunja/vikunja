import {Factory} from '../support/factory'
import {formatISO} from 'date-fns'

export class TaskAssigneeFactory extends Factory {
	static table = 'task_assignees'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			task_id: 1,
			user_id: 1,
			created: formatISO(now),
		}
	}
}