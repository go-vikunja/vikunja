import {Factory} from '../support/factory'
import {formatISO} from 'date-fns'

export class LabelTaskFactory extends Factory {
	static table = 'label_tasks'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			task_id: 1,
			label_id: 1,
			created: formatISO(now),
		}
	}
}