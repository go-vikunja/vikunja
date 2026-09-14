import {beforeEach, describe, expect, it} from 'vitest'

import {removeToken, saveToken} from '@/helpers/auth'
import {
	assertClientRequestContext,
	captureClientRequestContext,
	isClientRequestContextCurrent,
} from './requestContext'

function token(id: number, type = 1): string {
	const payload = btoa(JSON.stringify({id, type}))
	return `header.${payload}.signature`
}

function expectStaleContext(context: ReturnType<typeof captureClientRequestContext>) {
	expect(isClientRequestContextCurrent(context)).toBe(false)
	try {
		assertClientRequestContext(context)
	} catch (error) {
		expect(error).toMatchObject({name: 'AbortError'})
		return
	}
	throw new Error('Expected stale client request context')
}

describe('client request context', () => {
	beforeEach(() => {
		removeToken()
		window.API_URL = 'https://identity-a.example/api/v1/'
		saveToken(token(1), false)
	})

	it('recognizes the current identity, session, and API URL', () => {
		const context = captureClientRequestContext()

		expect(isClientRequestContextCurrent(context)).toBe(true)
		expect(() => assertClientRequestContext(context)).not.toThrow()
	})

	it('rejects an older session for the same identity', () => {
		const context = captureClientRequestContext()

		removeToken()
		saveToken(token(1), false)

		expectStaleContext(context)
	})

	it('rejects a different identity in the same session', () => {
		const context = captureClientRequestContext()

		saveToken(token(2), false)

		expectStaleContext(context)
	})

	it('rejects a different API URL', () => {
		const context = captureClientRequestContext()

		window.API_URL = 'https://identity-b.example/api/v1/'

		expectStaleContext(context)
	})
})
