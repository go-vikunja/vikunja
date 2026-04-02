/**
 * Generate a cryptographically random code_verifier (43-128 chars, RFC 7636 Section 4.1).
 * Uses unreserved characters: [A-Z] / [a-z] / [0-9] / "-" / "." / "_" / "~"
 */
export function generateCodeVerifier(): string {
	const array = new Uint8Array(32)
	crypto.getRandomValues(array)
	return base64UrlEncode(array)
}

/**
 * Compute code_challenge = BASE64URL(SHA256(code_verifier)) (RFC 7636 Section 4.2).
 */
export async function generateCodeChallenge(verifier: string): Promise<string> {
	const encoder = new TextEncoder()
	const data = encoder.encode(verifier)
	const digest = await crypto.subtle.digest('SHA-256', data)
	return base64UrlEncode(new Uint8Array(digest))
}

function base64UrlEncode(bytes: Uint8Array): string {
	let binary = ''
	for (const byte of bytes) {
		binary += String.fromCharCode(byte)
	}
	return btoa(binary)
		.replace(/\+/g, '-')
		.replace(/\//g, '_')
		.replace(/=+$/, '')
}
