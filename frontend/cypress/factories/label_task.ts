import {Factory} from '../support/factory'

export class LabelTaskFactory extends Factory {
	static table = 'label_tasks'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			task_id: 1,
			label_id: 1,
			created: now.toISOString(),
		}
	}
}
