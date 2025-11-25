import {faker} from '@faker-js/faker'
import {Factory} from '../support/factory'

export interface TokenAttributes {
	id: number | '{increment}';
	user_id: number;
	token: string;
	kind: number;
	created: string;
}

export class TokenFactory extends Factory {
	static table = 'user_tokens'

	// The factory method itself produces an object where id is '{increment}' (a string)
	// before it gets processed by the main create() method in the base Factory class.
	static factory(attrs?: Partial<Omit<TokenAttributes, 'id'>>): Omit<TokenAttributes, 'id'> & { id: string } {
		const now = new Date()

		return {
			id: '{increment}', // This is a string
			user_id: 1,      // Default user_id
			token: faker.string.alphanumeric(64),
			kind: 1,         // TokenPasswordReset
			created: now.toISOString(),
			...(attrs ?? {}),
		}
	}
} 
