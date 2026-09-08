/**
 * Security parameters for the OIDC authorization request: a PKCE code verifier and
 * challenge (RFC 7636) and a nonce (OpenID Connect Core 1.0, §3.1.2.1).
 *
 * These have to be unguessable, so they are drawn from the Web Crypto RNG rather than
 * from `createRandomID`, which is backed by `Math.random`.
 */

const CODE_VERIFIER_BYTES = 32
const NONCE_BYTES = 16

export const CODE_VERIFIER_STORAGE_KEY = 'codeVerifier'
export const NONCE_STORAGE_KEY = 'nonce'

function base64UrlEncode(bytes: Uint8Array): string {
	let binary = ''
	bytes.forEach(byte => {
		binary += String.fromCharCode(byte)
	})
	return btoa(binary)
		.replace(/\+/g, '-')
		.replace(/\//g, '_')
		.replace(/=+$/, '')
}

function randomBase64Url(byteLength: number): string {
	const bytes = new Uint8Array(byteLength)
	crypto.getRandomValues(bytes)
	return base64UrlEncode(bytes)
}

/**
 * A high-entropy code verifier as described in RFC 7636, §4.1. 32 random bytes
 * base64url-encode to 43 characters, the minimum length the RFC allows.
 */
export function createCodeVerifier(): string {
	return randomBase64Url(CODE_VERIFIER_BYTES)
}

export function createNonce(): string {
	return randomBase64Url(NONCE_BYTES)
}

/**
 * The S256 challenge for a verifier (RFC 7636, §4.2).
 *
 * Returns undefined when the browser does not expose `crypto.subtle`, which is the case
 * whenever Vikunja is served over plain http: SubtleCrypto is restricted to secure
 * contexts. Callers then have to fall back to an authorization request without PKCE
 * rather than break the login.
 */
export async function createCodeChallenge(verifier: string): Promise<string | undefined> {
	if (typeof crypto === 'undefined' || typeof crypto.subtle === 'undefined') {
		return undefined
	}

	const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier))
	return base64UrlEncode(new Uint8Array(digest))
}
