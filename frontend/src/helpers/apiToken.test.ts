import {describe, it, expect} from 'vitest'
import {isApiTokenExpired, oauthClientHost} from './apiToken'

const NOW = Date.parse('2026-01-01T00:00:00Z')

describe('isApiTokenExpired', () => {
	it('reports an expiry in the past as expired', () => {
		expect(isApiTokenExpired({expires_at: '2025-12-31T23:59:59Z'}, NOW)).toBe(true)
	})

	it('reports an expiry in the future as not expired', () => {
		expect(isApiTokenExpired({expires_at: '2026-01-01T00:00:01Z'}, NOW)).toBe(false)
	})

	it.each([
		undefined,
		'',
		'not a date',
		'0001-01-01T00:00:00Z',
	])('treats an unusable expiry (%j) as never expiring', expires_at => {
		expect(isApiTokenExpired({expires_at}, NOW)).toBe(false)
	})
})

describe('oauthClientHost', () => {
	it('returns the host of a metadata document client', () => {
		expect(oauthClientHost({oauth_client_id: 'https://app.example.com:8443/oauth/client.json'})).toBe('app.example.com:8443')
	})

	it.each([
		undefined,
		'',
		'vikunja-desktop',
		'http://app.example.com/client.json',
	])('returns nothing for %j', oauth_client_id => {
		expect(oauthClientHost({oauth_client_id})).toBeUndefined()
	})
})
