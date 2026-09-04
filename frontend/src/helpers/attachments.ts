import AttachmentModel from '@/models/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'

import AttachmentService, {type PREVIEW_SIZE} from '@/services/attachment'

const blobService = new AttachmentService()
const blobUrlCache = new Map<string, string>()
const pendingBlobRequests = new Map<string, Promise<string>>()

/**
 * Blob urls are shared between every consumer, so callers must not revoke them.
 * Use clearAttachmentBlobCache() instead.
 */
export function fetchAttachmentBlobUrl(attachment: Pick<IAttachment, 'id' | 'taskId'>, size?: PREVIEW_SIZE): Promise<string> {
	const key = `${attachment.taskId}-${attachment.id}-${size ?? ''}`

	const cached = blobUrlCache.get(key)
	if (cached !== undefined) {
		return Promise.resolve(cached)
	}

	const pending = pendingBlobRequests.get(key)
	if (pending !== undefined) {
		return pending
	}

	const request = blobService.getBlobUrl(attachment, size)
		.then(url => {
			blobUrlCache.set(key, url)
			pendingBlobRequests.delete(key)
			return url
		})
		.catch(e => {
			// drop the rejected promise, else every retry rethrows it
			pendingBlobRequests.delete(key)
			throw e
		})

	pendingBlobRequests.set(key, request)
	return request
}

export function clearAttachmentBlobCache() {
	blobUrlCache.forEach(url => window.URL.revokeObjectURL(url))
	blobUrlCache.clear()
	pendingBlobRequests.clear()
}

export async function uploadFile(taskId: number, file: File, onSuccess?: (url: string) => void): Promise<IAttachment[]> {
	const attachmentService = new AttachmentService()
	const files = [file]

	return await uploadFiles(attachmentService, taskId, files, onSuccess)
}

export async function uploadFiles(
	attachmentService: AttachmentService,
	taskId: number,
	files: File[] | FileList,
	onSuccess?: (attachmentUrl: string) => void,
): Promise<IAttachment[]> {
	const attachmentModel = new AttachmentModel({taskId})
	const response = await attachmentService.create(attachmentModel, files)
	console.debug(`Uploaded attachments for task ${taskId}, response was`, response)

	const uploaded: IAttachment[] = []
	response.success?.map((attachment: IAttachment) => {
		uploaded.push(attachment)
		onSuccess?.(generateAttachmentUrl(taskId, attachment.id))
	})

	if (response.errors !== null) {
		const messages = response.errors.map((e: {message: string}) => e.message)
		throw new Error(messages.join('\n'))
	}

	return uploaded
}

export function generateAttachmentUrl(taskId: number, attachmentId: number) {
	return `${window.API_URL}/tasks/${taskId}/attachments/${attachmentId}`
}
