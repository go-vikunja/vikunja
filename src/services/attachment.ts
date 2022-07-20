import AbstractService from './abstractService'
import AttachmentModel from '../models/attachment'
import {formatISO} from 'date-fns'
import {downloadBlob} from '@/helpers/downloadBlob'
import type FileModel from '@/models/file'

export default class AttachmentService extends AbstractService<AttachmentModel> {
	constructor() {
		super({
			create: '/tasks/{taskId}/attachments',
			getAll: '/tasks/{taskId}/attachments',
			delete: '/tasks/{taskId}/attachments/{id}',
		})
	}

	processModel(model: AttachmentModel) {
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

	getBlobUrl(model: AttachmentModel) {
		return AbstractService.prototype.getBlobUrl.call(this, '/tasks/' + model.taskId + '/attachments/' + model.id)
	}

	async download(model: AttachmentModel) {
		const url = await this.getBlobUrl(model)
		return downloadBlob(url, model.file.name)
	}

	/**
	 * Uploads a file to the server
	 * @param files
	 * @returns {Promise<any|never>}
	 */
	create(model: AttachmentModel, files: FileModel[]) {
		const data = new FormData()
		for (let i = 0; i < files.length; i++) {
			// TODO: Validation of file size
			data.append('files', new Blob([JSON.stringify(files[i], null, 2)]), files[i].name)
		}

		return this.uploadFormData(
			this.getReplacedRoute(this.paths.create, model),
			data,
		)
	}
}
