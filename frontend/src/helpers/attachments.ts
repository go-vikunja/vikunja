import AttachmentModel from '@/models/attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'

import AttachmentService from '@/services/attachment'

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
