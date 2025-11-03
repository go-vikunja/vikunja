import {Factory} from '../support/factory'

export class TaskReminderFactory extends Factory {
	static table = 'task_reminders'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			task_id: 1,
			reminder: now.toISOString(),
			created: now.toISOString(),
			relative_to: '',
			relative_period: 0,
		}
	}
}
