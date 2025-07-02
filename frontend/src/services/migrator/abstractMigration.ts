import AbstractService from '../abstractService'
import type { IAbstract } from '@/modelTypes/IAbstract'

export interface MigrationConfig extends IAbstract {
	code: string
}

// This service builds on top of the abstract service and basically just hides away method names.
// It enables migration services to be created with minimal overhead and even better method names.
export default class AbstractMigrationService extends AbstractService<MigrationConfig> {
	serviceUrlKey = ''

	constructor(serviceUrlKey: string) {
		super({
			update: '/migration/' + serviceUrlKey + '/migrate',
		})
		this.serviceUrlKey = serviceUrlKey
	}

	getAuthUrl() {
		return this.getM('/migration/' + this.serviceUrlKey + '/auth', { code: '', maxRight: null })
	}

	getStatus() {
		return this.getM('/migration/' + this.serviceUrlKey + '/status', { code: '', maxRight: null })
	}

	migrate(data: MigrationConfig) {
		return this.update(data)
	}
}
