import AbstractService from './abstractService'
import AttachmentModel from '../models/attachment'
import {formatISO} from 'date-fns'
import {downloadBlob} from '@/helpers/downloadBlob'

export default class AttachmentService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{taskId}/attachments',
			getAll: '/tasks/{taskId}/attachments',
			delete: '/tasks/{taskId}/attachments/{id}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		return model
	}

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

	getBlobUrl(model) {
		return AbstractService.prototype.getBlobUrl.call(this, '/tasks/' + model.taskId + '/attachments/' + model.id)
	}

	download(model) {
		this.getBlobUrl(model)
			.then(url => downloadBlob(url, model.file.name))
	}

	/**
	 * Uploads a file to the server
	 * @param model
	 * @param files
	 * @returns {Promise<any|never>}
	 */
	create(model, files) {
		const data = new FormData()
		for (let i = 0; i < files.length; i++) {
			// TODO: Validation of file size
			data.append('files', new Blob([files[i]]), files[i].name)
		}

		return this.uploadFormData(
			this.getReplacedRoute(this.paths.create, model),
			data,
		)
	}
}
