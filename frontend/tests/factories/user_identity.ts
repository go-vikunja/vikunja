import {Factory} from '../support/factory'

export class UserIdentityFactory extends Factory {
	static table = 'user_identities'

	static factory() {
		return {
			id: '{increment}',
			user_id: 1,
			issuer: 'https://gone.example.com',
			subject: 'subject-1',
			created: new Date().toISOString(),
		}
	}
}
