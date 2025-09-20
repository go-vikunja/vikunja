import AbstractService from './abstractService'
import AttachmentModel from '../models/attachment'

import type { IAttachment } from '@/modelTypes/IAttachment'

import {downloadBlob} from '@/helpers/downloadBlob'

export enum PREVIEW_SIZE {
	SM = 'sm',
	MD = 'md',
	LG = 'lg',
	XL = 'xl',
}

export default class AttachmentService extends AbstractService<IAttachment> {
	constructor() {
		super({
			create: '/tasks/{taskId}/attachments',
			getAll: '/tasks/{taskId}/attachments',
			delete: '/tasks/{taskId}/attachments/{id}',
		})
	}

	processModel(model: IAttachment) {
		return {
			...model,
			created: new Date(model.created).toISOString(),
		}
	}

	useCreateInterceptor() {
		return false
	}

	modelFactory(data: Partial<IAttachment>) {
		return new AttachmentModel(data)
	}

	modelCreateFactory(data: any) {
		// Success contains the uploaded attachments
		data.success = (data.success === null ? [] : data.success).map((a: any) => {
			return this.modelFactory(a)
		})
		return data
	}

	getAttachmentBlobUrl(model: IAttachment, size?: PREVIEW_SIZE) {
		let mainUrl = '/tasks/' + model.taskId + '/attachments/' + model.id
		if (size !== undefined) {
			mainUrl += `?preview_size=${size}`
		}

		return super.getBlobUrl(mainUrl)
	}

	async download(model: IAttachment) {
		const url = await this.getAttachmentBlobUrl(model)
		return downloadBlob(url, model.file.name)
	}

	/**
	 * Uploads a file to the server
	 * @param files
	 * @returns {Promise<any|never>}
	 */
	createAttachments(model: IAttachment, files: File[] | FileList) {
		const data = new FormData()
		for (let i = 0; i < files.length; i++) {
			const file = files[i]
			if (file) {
				// TODO: Validation of file size
				data.append('files', new Blob([file]), file.name)
			}
		}

		return this.uploadFormData(
			this.getReplacedRoute(this.paths.create, model),
			data,
		)
	}
}
