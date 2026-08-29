import AbstractService from '../abstractService'
import {apiV2Url} from '@/helpers/fetcher'

export type MigrationConfig =
	| { code: string }
	| { url: string, token: string }
	| { url: string, username: string, password: string }

// This service builds on top of the abstract service and basically just hides away method names.
// It enables migration services to be created with minimal overhead and even better method names.
export default class AbstractMigrationService extends AbstractService<MigrationConfig> {
	serviceUrlKey = ''

	constructor(serviceUrlKey: string) {
		super({
			update: apiV2Url(`migration/${serviceUrlKey}/migrate`),
		})
		this.serviceUrlKey = serviceUrlKey
	}

	getAuthUrl() {
		return this.getM(apiV2Url(`migration/${this.serviceUrlKey}/auth`))
	}

	getStatus() {
		return this.getM(apiV2Url(`migration/${this.serviceUrlKey}/status`))
	}

	migrate(data: MigrationConfig) {
		return this.update(data)
	}
}
