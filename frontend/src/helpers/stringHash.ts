// Deterministic non-cryptographic string hash (djb2 variant).
// Used for stable pseudo-random selection keyed on date + user + bucket.
export function stringHash(input: string): number {
	let hash = 5381
	for (let i = 0; i < input.length; i++) {
		// hash * 33 + char, kept in 32-bit range via `| 0`.
		hash = ((hash << 5) + hash + input.charCodeAt(i)) | 0
	}
	// Ensure non-negative.
	return hash >>> 0
}
