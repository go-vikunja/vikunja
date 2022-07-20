import AbstractService from './abstractService'
import ListModel, { type IList } from '../models/list'
import type { IFile } from '@/models/file'

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
	create(listId: IList['id'], file: IFile) {
		return this.uploadFile(
			this.getReplacedRoute(this.paths.create, {listId}),
			file,
			'background',
		)
	}
}
