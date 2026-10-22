import {Factory} from '../support/factory'

export class NotificationFactory extends Factory {
	static table = 'notifications'

	static factory() {
		return {
			id: '{increment}',
			notifiable_id: 1,
			name: 'task.comment',
			notification: JSON.stringify({}),
			project_id: 0,
			created: new Date().toISOString(),
		}
	}
}
