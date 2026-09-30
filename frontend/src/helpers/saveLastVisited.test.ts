import {describe, it, expect, beforeEach} from 'vitest'

import {getLastVisited, saveLastVisited} from './saveLastVisited'

describe('saveLastVisited', () => {
	beforeEach(() => {
		localStorage.clear()
	})

	it('does not persist one-time tokens', () => {
		saveLastVisited('home', {}, {
			userPasswordReset: 'a',
			userEmailConfirm: 'b',
			accountDeletionConfirm: 'c',
			filter: 'done = false',
		})

		expect(getLastVisited()).toEqual({name: 'home', params: {}, query: {filter: 'done = false'}})
	})
})
