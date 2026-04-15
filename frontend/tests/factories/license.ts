import {Factory} from '../support/factory'

/**
 * Seeds the license_status row. The backend re-applies the cached response
 * on this seed call, so licensed features take effect immediately.
 */
export class LicenseFactory extends Factory {
	static table = 'license_status'

	static async enable(features: string[]) {
		const now = new Date().toISOString()
		const response = JSON.stringify({
			valid: true,
			features,
			max_users: 0,
			expires_at: '2099-01-01T00:00:00Z',
		})
		await this.seed(this.table, [{
			id: 1,
			instance_id: '00000000-0000-0000-0000-000000000000',
			response,
			validated_at: now,
			created: now,
			updated: now,
		}])
	}

	static async disable() {
		const now = new Date().toISOString()
		await this.seed(this.table, [{
			id: 1,
			instance_id: '00000000-0000-0000-0000-000000000000',
			response: '{}',
			validated_at: null,
			created: now,
			updated: now,
		}])
	}
}
