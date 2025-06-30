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

	modelCreateFactory(data: {success: Partial<IAttachment>[] | null}) {
		// Success contains the uploaded attachments
		data.success = (data.success === null ? [] : data.success).map((a: Partial<IAttachment>) => {
			return this.modelFactory(a)
		})
		return data
	}

	getBlobUrl(model: IAttachment, size?: PREVIEW_SIZE): Promise<unknown>
	getBlobUrl(url: string, method?: string, data?: unknown): Promise<unknown>
	getBlobUrl(modelOrUrl: IAttachment | string, sizeOrMethod?: PREVIEW_SIZE | string, data?: unknown): Promise<unknown> {
		if (typeof modelOrUrl === 'string') {
			return super.getBlobUrl(modelOrUrl, sizeOrMethod, data)
		}
		
		const model = modelOrUrl
		let mainUrl = '/tasks/' + model.taskId + '/attachments/' + model.id
		if (sizeOrMethod !== undefined) {
			mainUrl += `?preview_size=${sizeOrMethod}`
		}

		return super.getBlobUrl(mainUrl)
	}

	async download(model: IAttachment) {
		const url = await this.getBlobUrl(model)
		return downloadBlob(url as string, model.file.name)
	}

	/**
	 * Uploads a file to the server
	 * @param files
	 * @returns {Promise<any|never>}
	 */
	create(model: IAttachment): Promise<IAttachment>
	create(model: IAttachment, files: File[] | FileList): Promise<unknown>
	create(model: IAttachment, files?: File[] | FileList): Promise<unknown> {
		if (!files) {
			return super.create(model)
		}
		const data = new FormData()
		for (let i = 0; i < files.length; i++) {
			// TODO: Validation of file size
			data.append('files', new Blob([files[i]]), files[i].name)
		}

		return this.uploadFormData(
			this.getReplacedRoute(this.paths.create, model as Record<string, unknown>),
			data,
		)
	}
}
