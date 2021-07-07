export const generateAttachmentUrl = (taskId, attachmentId) => {
	return `${window.API_URL}/tasks/${taskId}/attachments/${attachmentId}`
}