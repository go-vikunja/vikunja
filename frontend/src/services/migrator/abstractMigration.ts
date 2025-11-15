import AbstractService from '../abstractService'

export type MigrationConfig = {
	code: string
	server_url?: string
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

	getAuthUrl(serverUrl?: string) {
		if (serverUrl) {
			return this.post('/migration/' + this.serviceUrlKey + '/auth', {server_url: serverUrl})
		}
		return this.getM('/migration/' + this.serviceUrlKey + '/auth')
	}

	getStatus() {
		return this.getM('/migration/' + this.serviceUrlKey + '/status')
	}

	migrate(data: MigrationConfig) {
		return this.update(data)
	}
}
