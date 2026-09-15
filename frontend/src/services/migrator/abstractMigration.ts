import AbstractService from '../abstractService'
import {apiV2Url} from '@/helpers/fetcher'

export type MigrationConfig =
	| { code: string }
	| { url: string, token: string }
	| { url: string, username: string, password: string }

// '' means no failure. Anything but 'detail' is a closed set the frontend translates itself;
// 'detail' carries untranslatable text from the user's own data in error_message.
export type MigrationErrorKind = '' | 'reported' | 'interrupted' | 'credentials' | 'queue' | 'upload' | 'detail'

export interface MigrationStatus {
	started_at: string | null
	finished_at: string | null
	error_kind: MigrationErrorKind
	error_message: string
}

interface MigrationStatusEndpoint {
	getM(url: string): Promise<unknown>
}

// The status route answers with a migration status, not the model the migration services are typed on.
export function getMigrationStatus(service: MigrationStatusEndpoint, url: string): Promise<MigrationStatus> {
	return service.getM(url) as Promise<MigrationStatus>
}

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
		return getMigrationStatus(this, apiV2Url(`migration/${this.serviceUrlKey}/status`))
	}

	migrate(data: MigrationConfig) {
		return this.update(data)
	}
}
