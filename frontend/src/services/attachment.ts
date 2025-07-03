import AbstractService from './abstractService'
import AttachmentModel from '../models/attachment'

import type { IAttachment, IAttachmentUploadResponse } from '@/modelTypes/IAttachment'
import type { Method } from 'axios'

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
			created: new Date(model.created),
		}
	}

	useCreateInterceptor() {
		return false
	}

	modelFactory(data: Partial<IAttachment>) {
		return new AttachmentModel(data)
	}

	modelCreateFactory(data: Partial<IAttachment>): IAttachment {
		return this.modelFactory(data)
	}

	// Special factory for file upload responses
	processUploadResponse(data: IAttachmentUploadResponse): IAttachmentUploadResponse {
		// Success contains the uploaded attachments
		if (data.success) {
			data.success = data.success.map((a: Partial<IAttachment>) => {
				return this.modelFactory(a)
			})
		}
		return data
	}

	getBlobUrl(model: IAttachment, size?: PREVIEW_SIZE): Promise<unknown>
	getBlobUrl(url: string, method?: string, data?: unknown): Promise<unknown>
	getBlobUrl(modelOrUrl: IAttachment | string, sizeOrMethod?: PREVIEW_SIZE | string, data?: unknown): Promise<unknown> {
		if (typeof modelOrUrl === 'string') {
			return super.getBlobUrl(modelOrUrl, sizeOrMethod as Method, data as Record<string, unknown> | undefined)
		}
		
		const model = modelOrUrl
		let mainUrl = '/tasks/' + model.taskId + '/attachments/' + model.id
		if (sizeOrMethod !== undefined) {
			mainUrl += `?preview_size=${sizeOrMethod}`
		}

		return super.getBlobUrl(mainUrl, 'GET', {})
	}

	async download(model: IAttachment) {
		const url = await this.getBlobUrl(model)
		return downloadBlob(url as string, model.file.name)
	}

	/**
	 * Uploads a file to the server
	 * @param files
	 * @returns {Promise<IAttachment|IAttachmentUploadResponse>}
	 */
	create(model: IAttachment): Promise<IAttachment>
	create(model: IAttachment, files: File[] | FileList): Promise<IAttachmentUploadResponse>
	create(model: IAttachment, files?: File[] | FileList): Promise<IAttachment | IAttachmentUploadResponse> {
		if (!files) {
			return super.create(model)
		}
		const data = new FormData()
		for (let i = 0; i < files.length; i++) {
			// TODO: Validation of file size
			data.append('files', new Blob([files[i]]), files[i].name)
		}

		return this.uploadFormData(
			this.getReplacedRoute(this.paths.create, model as unknown as Record<string, unknown>),
			data,
		)
	}
}
