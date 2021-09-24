import AttachmentModel from '@/models/attachment'
import FileModel from '@/models/file'

import AttachmentService from '@/services/attachment'
import { store } from '@/store'

export function uploadFile(taskId: number, file: FileModel, onSuccess: () => Function) {
	const attachmentService = new AttachmentService()
	const files = [file]

	return uploadFiles(attachmentService, taskId, files, onSuccess)
}

export function uploadFiles(attachmentService: AttachmentService, taskId: number, files: FileModel[], onSuccess : Function = () => {}) {
	const attachmentModel = new AttachmentModel({taskId})
	attachmentService.create(attachmentModel, files)
		.then(r => {
			console.debug(`Uploaded attachments for task ${taskId}, response was`, r)
			if (r.success !== null) {
				r.success.forEach((attachment: AttachmentModel) => {
					store.dispatch('tasks/addTaskAttachment', {
						taskId,
						attachment,
					})
					onSuccess(generateAttachmentUrl(taskId, attachment.id))
				})
			}
			if (r.errors !== null) {
				throw Error(r.errors)
			}
		})
}

export function generateAttachmentUrl(taskId: number, attachmentId: number) : any {
	return `${window.API_URL}/tasks/${taskId}/attachments/${attachmentId}`
}