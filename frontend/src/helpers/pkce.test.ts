import {describe, it, expect, vi, afterEach} from 'vitest'

import {createCodeChallenge, createCodeVerifier, createNonce} from './pkce'

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

describe('createNonce', () => {
	it('is base64url with no padding', () => {
		expect(createNonce()).toMatch(BASE64_URL)
	})

	it('differs every time', () => {
		expect(createNonce()).not.toBe(createNonce())
	})
})

describe('createCodeChallenge', () => {
	// RFC 7636, appendix B.
	it('matches the challenge from the RFC test vector', async () => {
		const verifier = 'dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk'

		expect(await createCodeChallenge(verifier)).toBe('E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM')
	})

	// SubtleCrypto is restricted to secure contexts, so an instance served over plain
	// http has no way to hash the verifier. The caller has to notice and leave PKCE out
	// of the authorization request rather than send a broken challenge.
	it('returns undefined when the browser exposes no subtle crypto', async () => {
		vi.stubGlobal('crypto', {getRandomValues: crypto.getRandomValues.bind(crypto)})

		expect(await createCodeChallenge('dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk')).toBeUndefined()
	})
})
