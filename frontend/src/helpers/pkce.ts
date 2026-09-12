// Values here must be unguessable, so they come from crypto.getRandomValues rather than
// createRandomID, which is Math.random-backed.

const CODE_VERIFIER_BYTES = 32
const NONCE_BYTES = 16
const STATE_BYTES = 16

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

// RFC 7636 §4.1: 32 random bytes base64url to 43 characters, the RFC minimum.
export function createCodeVerifier(): string {
	return randomBase64Url(CODE_VERIFIER_BYTES)
}

export function createNonce(): string {
	return randomBase64Url(NONCE_BYTES)
}

export function createState(): string {
	return randomBase64Url(STATE_BYTES)
}

// Returns undefined when crypto.subtle is unavailable (non-secure context, e.g. plain http)
// or fails — callers then send the request without PKCE rather than break the login.
export async function createCodeChallenge(verifier: string): Promise<string | undefined> {
	if (typeof crypto.subtle === 'undefined') {
		return undefined
	}

	try {
		const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier))
		return base64UrlEncode(new Uint8Array(digest))
	} catch {
		return undefined
	}
}
