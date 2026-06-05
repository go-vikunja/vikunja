import {faker} from '@faker-js/faker'
import {Factory} from '../support/factory'

// The `events` column is a JSON-serialized string in XORM; pass it pre-encoded.
export class WebhookFactory extends Factory {
	static table = 'webhooks'

	static factory() {
		const now = new Date()
		return {
			id: '{increment}',
			project_id: 1,
			target_url: faker.internet.url(),
			events: JSON.stringify(['task.created']),
			created_by_id: 1,
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
