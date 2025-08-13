import {Factory} from '../support/factory'

export class UserProjectFactory extends Factory {
	static table = 'users_projects'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			project_id: 1,
			user_id: 1,
			permission: 0,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
