import {describe, it, expect} from 'vitest'

import {getDisplayName} from '@/helpers/user'
import type {User as IUser} from '@/client/generated'

function makeUser(overrides: Partial<IUser> = {}): IUser {
	return {
		id: 1,
		username: 'testuser',
		...overrides,
	}
}

describe('getDisplayName', () => {
	it('falls back when generated optional fields are absent', () => {
		expect(getDisplayName({username: 'sam'})).toBe('sam')
		expect(getDisplayName({})).toBe('')
	})

	it('should return the name when set', () => {
		const user = makeUser({name: 'Jane Doe'})
		expect(getDisplayName(user)).toBe('Jane Doe')
	})

	it('should fall back to username when name is empty', () => {
		const user = makeUser({name: '', username: 'janedoe'})
		expect(getDisplayName(user)).toBe('janedoe')
	})
})

