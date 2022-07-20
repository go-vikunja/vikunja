import AbstractService from './abstractService'
import ListModel from '../models/list'
import type FileModel from '@/models/file'

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
	 * @param file
	 * @returns {Promise<any|never>}
	 */
	create(listId: ListModel['id'], file: FileModel) {
		return this.uploadFile(
			this.getReplacedRoute(this.paths.create, {listId}),
			file,
			'background',
		)
	}
}
