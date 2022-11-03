import AbstractService from '../abstractService'

// This service builds on top of the abstract service and basically just hides away method names.
// It enables migration services to be created with minimal overhead and even better method names.
export default class AbstractMigrationFileService extends AbstractService {
	serviceUrlKey = ''

	constructor(serviceUrlKey: string) {
		super({
			create: '/migration/' + serviceUrlKey + '/migrate',
		})
		this.serviceUrlKey = serviceUrlKey
	}

	getStatus() {
		return this.getM('/migration/' + this.serviceUrlKey + '/status')
	}
	
	useCreateInterceptor() {
		return false
	}

	migrate(file: File) {
		return this.uploadFile(
			this.paths.create,
			file,
			'import',
		)
	}
}
