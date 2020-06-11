import AbstractService from './abstractService'
import ListModel from '../models/list'

export default class BackgroundUploadService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listId}/backgrounds/upload',
		})
	}

	uploadProgress = 0

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

		let data = new FormData()
		data.append('background', new Blob([file]), file.name);

		const cancel = this.setLoading()
		return this.http.put(
			this.getReplacedRoute(this.paths.create, {listId: listId}),
			data,
			{
				headers: {
					'Content-Type':
						'multipart/form-data; boundary=' + data._boundary,
				},
				onUploadProgress: progressEvent => {
					this.uploadProgress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
				}
			}
		)
			.catch(error => {
				return this.errorHandler(error)
			})
			.then(response => {
				return Promise.resolve(this.modelCreateFactory(response.data))
			})
			.finally(() => {
				this.uploadProgress = 0
				cancel()
			})
	}
}
