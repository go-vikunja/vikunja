import AbstractService from './abstractService'
import AttachmentModel from '../models/attachment'
import {formatISO} from 'date-fns'

export default class AttachmentService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{taskId}/attachments',
			getAll: '/tasks/{taskId}/attachments',
			delete: '/tasks/{taskId}/attachments/{id}',
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		return model
	}

	uploadProgress = 0

	useCreateInterceptor() {
		return false
	}

	modelFactory(data) {
		return new AttachmentModel(data)
	}

	modelCreateFactory(data) {
		// Success contains the uploaded attachments
		data.success = (data.success === null ? [] : data.success).map(a => {
			return this.modelFactory(a)
		})
		return data
	}

	download(model) {
		this.http({
			url: '/tasks/' + model.taskId + '/attachments/' + model.id,
			method: 'GET',
			responseType: 'blob',
		}).then((response) => {
			const url = window.URL.createObjectURL(new Blob([response.data]));
			const link = document.createElement('a');
			link.href = url;
			link.setAttribute('download', model.file.name);
			link.click();
			window.URL.revokeObjectURL(url);
		});
	}

	/**
	 * Uploads a file to the server
	 * @param model
	 * @param files
	 * @returns {Promise<any|never>}
	 */
	create(model, files) {

		let data = new FormData()
		for (let i = 0; i < files.length; i++) {
			// TODO: Validation of file size
			data.append('files', new Blob([files[i]]), files[i].name);
		}

		const cancel = this.setLoading()
		return this.http.put(
			this.getReplacedRoute(this.paths.create, model),
			data,
			{
				headers: {
					'Content-Type':
						'multipart/form-data; boundary=' + data._boundary,
				},
				onUploadProgress: progressEvent => {
					this.uploadProgress = Math.round( (progressEvent.loaded * 100) / progressEvent.total );
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
