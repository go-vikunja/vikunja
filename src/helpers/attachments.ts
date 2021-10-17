import AttachmentModel from '@/models/attachment'
import FileModel from '@/models/file'

import AttachmentService from '@/services/attachment'
import { store } from '@/store'

export function uploadFile(taskId: number, file: FileModel, onSuccess: () => Function) {
	const attachmentService = new AttachmentService()
	const files = [file]

	return uploadFiles(attachmentService, taskId, files, onSuccess)
}

export async function uploadFiles(attachmentService: AttachmentService, taskId: number, files: FileModel[], onSuccess : Function = () => {}) {
	const attachmentModel = new AttachmentModel({taskId})
	const response = await attachmentService.create(attachmentModel, files)
	console.debug(`Uploaded attachments for task ${taskId}, response was`, response)

	response.success?.map((attachment: AttachmentModel) => {
		store.dispatch('tasks/addTaskAttachment', {
			taskId,
			attachment,
		})
		onSuccess(generateAttachmentUrl(taskId, attachment.id))
	})

	if (response.errors !== null) {
		throw Error(response.errors)
	}
}

export function generateAttachmentUrl(taskId: number, attachmentId: number) : any {
	return `${window.API_URL}/tasks/${taskId}/attachments/${attachmentId}`
}