import {createHash, randomUUID} from 'node:crypto'
import {Factory} from '../support/factory'

// Mirrors pkg/models/sessions.go HashSessionToken(). Unsalted because refresh
// tokens are high-entropy (128 random bytes hex-encoded), not human passwords.
export function hashSessionToken(raw: string): string {
	return createHash('sha256').update(raw).digest('hex')
}

export class SessionFactory extends Factory {
	static table = 'sessions'

	static factory() {
		const now = new Date()
		return {
			id: randomUUID(),
			user_id: 1,
			token_hash: hashSessionToken('placeholder-override-me'),
			device_info: 'Firefox on Linux',
			ip_address: '192.0.2.5',
			is_long_session: false,
			last_active: now.toISOString(),
			created: now.toISOString(),
		}
	}
}
