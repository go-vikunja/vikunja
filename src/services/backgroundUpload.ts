import AbstractService from './abstractService'
import ListModel from '../models/list'

export default class BackgroundUploadService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listId}/backgrounds/upload',
		})
	}

	useCreateInterceptor() {
		return false
	}

	modelCreateFactory(data) {
		return new ListModel(data)
	}

	/**
	 * Uploads a file to the server
	 * @param listId
	 * @param file
	 * @returns {Promise<any|never>}
	 */
	create(listId, file) {
		return this.uploadFile(
			this.getReplacedRoute(this.paths.create, {listId: listId}),
			file,
			'background',
		)
	}
}
