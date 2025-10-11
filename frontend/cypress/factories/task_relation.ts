import {Factory} from '../support/factory'

export class TaskRelationFactory extends Factory {
	static table = 'task_relations'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			task_id: '{increment}',
			other_task_id: '{increment}',
			relation_kind: 'related',
			created_by_id: 1,
			created: now.toISOString(),
		}
	}
}
