import {randomBytes} from 'node:crypto'
import {Factory} from '../support/factory'

export class ApiTokenFactory extends Factory {
	static table = 'api_tokens'

	static factory() {
		const now = new Date()
		return {
			id: '{increment}',
			title: 'Token',
			token_sha256: randomBytes(32).toString('hex'),
			permissions: JSON.stringify({mcp: ['access']}),
			expires_at: new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000).toISOString(),
			owner_id: 1,
			created: now.toISOString(),
		}
	}
}
