import AbstractService from '../abstractService'

export type MigrationConfig = { code: string }

// This service builds on top of the abstract service and basically just hides away method names.
// It enables migration services to be created with minimal overhead and even better method names.
export default class AbstractMigrationService extends AbstractService<any> {
	serviceUrlKey = ''

	constructor(serviceUrlKey: string) {
		super({
			update: '/migration/' + serviceUrlKey + '/migrate',
		})
		this.serviceUrlKey = serviceUrlKey
	}

	getAuthUrl() {
		return this.getM('/migration/' + this.serviceUrlKey + '/auth', {} as any)
	}

	getStatus() {
		return this.getM('/migration/' + this.serviceUrlKey + '/status', {} as any)
	}

	migrate(data: MigrationConfig) {
		return this.update(data)
	}
}
