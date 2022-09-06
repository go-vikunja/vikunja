import AbstractService from './abstractService'
import ListModel from '@/models/list'

import type { IList } from '@/modelTypes/IList'
import type { IFile } from '@/modelTypes/IFile'

export default class BackgroundUploadService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listId}/backgrounds/upload',
		})
	}

	useCreateInterceptor() {
		return false
	}

	modelCreateFactory(data: Partial<IList>) {
		return new ListModel(data)
	}

	/**
	 * Uploads a file to the server
	 */
	create(listId: IList['id'], file: IFile) {
		return this.uploadFile(
			this.getReplacedRoute(this.paths.create, {listId}),
			file,
			'background',
		)
	}
}
