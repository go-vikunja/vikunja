import {Factory} from '../support/factory'

export class TaskPositionFactory extends Factory {
	static table = 'task_positions'

	static factory() {
		return {
			task_id: '{increment}',
			project_view_id: '{increment}',
			position: '{increment}',
		}
	}
}
