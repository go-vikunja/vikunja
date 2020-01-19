import AbstractService from '../abstractService'

// This service builds on top of the abstract service and basically just hides away method names.
// It enables migration services to be created with minimal overhead and even better method names.
export default class AbstractMigrationService extends AbstractService {
	constructor(serviceUrlKey) {
		super({
			update: '/migration/'+serviceUrlKey+'/migrate',
			get: '/migration/'+serviceUrlKey+'/auth',
		})
	}

	getAuthUrl() {
		return this.get({})
	}

	migrate(data) {
		return this.update(data)
	}
}