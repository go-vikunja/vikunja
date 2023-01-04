import {Factory} from '../support/factory'

export class UserListFactory extends Factory {
	static table = 'users_lists'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			list_id: 1,
			user_id: 1,
			right: 0,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}