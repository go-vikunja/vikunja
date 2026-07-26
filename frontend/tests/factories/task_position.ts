import {Factory} from '../support/factory'

export class TaskPositionFactory extends Factory {
	static table = 'task_positions'

	static factory() {
		return {
			task_id: '{increment}',
			project_view_id: 1,
			// Same spacing the api uses. Cramped positions make the api's conflict
			// repair give up and fully recalculate the view, which reorders by task id
			// and papers over ordering bugs.
			position: (i: number) => i * Math.pow(2, 16),
		}
	}
}
