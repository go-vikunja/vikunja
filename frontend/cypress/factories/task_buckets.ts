import {Factory} from '../support/factory'

export class TaskBucketFactory extends Factory {
	static table = 'task_buckets'

	static factory() {
		return {
			task_id: '{increment}',
			bucket_id: '{increment}',
			project_view_id: '{increment}',
		}
	}
}
