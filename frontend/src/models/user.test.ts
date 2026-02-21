import {describe, it, expect} from 'vitest'
import {getDisplayName} from './user'
import type {IUser} from '@/modelTypes/IUser'

function makeUser(overrides: Partial<IUser> = {}): IUser {
	return {
		id: 1,
		email: 'test@example.com',
		username: 'testuser',
		name: '',
		exp: 0,
		type: 1,
		created: new Date(),
		updated: new Date(),
		settings: {} as IUser['settings'],
		isLocalUser: true,
		deletionScheduledAt: null,
		...overrides,
	}
}

describe('getDisplayName', () => {
	it('should return the name when set', () => {
		const user = makeUser({name: 'Jane Doe'})
		expect(getDisplayName(user)).toBe('Jane Doe')
	})

	it('should fall back to username when name is empty', () => {
		const user = makeUser({name: '', username: 'janedoe'})
		expect(getDisplayName(user)).toBe('janedoe')
	})
})
