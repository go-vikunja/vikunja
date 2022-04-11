import {Factory} from '../support/factory'
import {formatISO} from "date-fns"

export class UserListFactory extends Factory {
	static table = 'users_lists'

	static factory() {
		const now = new Date()

		return {
			id: '{increment}',
			list_id: 1,
			user_id: 1,
			right: 0,
			created: formatISO(now),
			updated: formatISO(now)
		}
	}
}