import {describe, expect, it} from 'vitest'
import {getSensitiveParam} from './sensitiveQueryParams'

describe('getSensitiveParam', () => {
	it('reads the token from the fragment', () => {
		expect(getSensitiveParam({query: {}, hash: '#userPasswordReset=abc'}, 'userPasswordReset')).toBe('abc')
	})

	it('still reads the token from the query of older emails', () => {
		expect(getSensitiveParam({query: {userPasswordReset: 'abc'}, hash: ''}, 'userPasswordReset')).toBe('abc')
	})

	it('uses the first value of a repeated query token', () => {
		expect(getSensitiveParam({query: {userEmailConfirm: ['a', 'b']}, hash: ''}, 'userEmailConfirm')).toBe('a')
	})

	it.each([
		['missing', {query: {}, hash: '#comment-5'}],
		['empty', {query: {userPasswordReset: ''}, hash: '#userPasswordReset='}],
	])('returns undefined for a %s token', (_, route) => {
		expect(getSensitiveParam(route, 'userPasswordReset')).toBeUndefined()
	})
})
