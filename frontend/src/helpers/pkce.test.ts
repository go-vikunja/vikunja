import {describe, it, expect, vi, afterEach} from 'vitest'

import {createCodeChallenge, createCodeVerifier, createNonce, createState} from './pkce'

const BASE64_URL = /^[A-Za-z0-9\-_]+$/

afterEach(() => {
	vi.unstubAllGlobals()
})

describe('createCodeVerifier', () => {
	it('is base64url with no padding', () => {
		expect(createCodeVerifier()).toMatch(BASE64_URL)
	})

	it('is 43 characters, the shortest RFC 7636 allows', () => {
		expect(createCodeVerifier()).toHaveLength(43)
	})

	it('differs every time', () => {
		expect(createCodeVerifier()).not.toBe(createCodeVerifier())
	})
})

describe.each([
	['createNonce', createNonce],
	['createState', createState],
])('%s', (_name, create) => {
	it('is base64url with no padding', () => {
		expect(create()).toMatch(BASE64_URL)
	})

	it('differs every time', () => {
		expect(create()).not.toBe(create())
	})
})

describe('createCodeChallenge', () => {
	const verifier = 'dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk'

	// RFC 7636, appendix B.
	it('matches the challenge from the RFC test vector', async () => {
		expect(await createCodeChallenge(verifier)).toBe('E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM')
	})

	it('returns undefined when the browser exposes no subtle crypto', async () => {
		vi.stubGlobal('crypto', {getRandomValues: crypto.getRandomValues.bind(crypto)})

		expect(await createCodeChallenge(verifier)).toBeUndefined()
	})

	it('returns undefined when hashing fails', async () => {
		vi.stubGlobal('crypto', {
			getRandomValues: crypto.getRandomValues.bind(crypto),
			subtle: {digest: () => Promise.reject(new Error('no'))},
		})

		expect(await createCodeChallenge(verifier)).toBeUndefined()
	})
})
