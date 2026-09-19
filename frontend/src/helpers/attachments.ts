import type {TaskAttachment} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {
	attachmentBlob,
	attachmentKeys,
	type AttachmentIdentity,
	type PreviewSize,
} from '@/client/queries/attachments'
import {captureClientRequestContext, assertClientRequestContext} from '@/client/requestContext'
import {downloadBlob} from '@/helpers/downloadBlob'

// A blob: url for an svg inherits our origin and can script; a data: url cannot.
const SCRIPTABLE_MIME_TYPE = 'image/svg+xml'

// The cache holds the bytes; a url belongs to the caller that asked for it and nobody else may revoke it.
export function fetchAttachmentBlob(attachment: AttachmentIdentity, size?: PreviewSize): Promise<Blob> {
	return queryClient.fetchQuery({
		queryKey: attachmentKeys.blob(attachment.task_id, attachment.id, size),
		queryFn: ({signal}) => attachmentBlob(attachment, size, signal),
		staleTime: Infinity,
		retry: false,
	})
}

export async function fetchAttachmentUrl(attachment: AttachmentIdentity, size?: PreviewSize): Promise<string> {
	const context = captureClientRequestContext()
	const blob = await fetchAttachmentBlob(attachment, size)
	assertClientRequestContext(context)
	const mimeType = blob.type.split(';')[0].trim().toLowerCase()
	// FileReader is absent in iOS Lockdown Mode and some webviews, fall back to a blob url there.
	if (mimeType !== SCRIPTABLE_MIME_TYPE || typeof FileReader === 'undefined') {
		return URL.createObjectURL(blob)
	}

	return new Promise<string>((resolve, reject) => {
		const reader = new FileReader()
		reader.onload = () => {
			// the identity can change while the read runs, so the pre-read fence has to be repeated
			try {
				assertClientRequestContext(context)
			} catch (fenced) {
				reject(fenced)
				return
			}
			if (typeof reader.result === 'string') {
				resolve(reader.result)
				return
			}
			reject(new Error('Attachment could not be read as a data url'))
		}
		reader.onerror = () => reject(reader.error ?? new Error('Attachment could not be read as a data url'))
		reader.readAsDataURL(blob)
	})
}

// The svg defence hands out inert data: urls, which own nothing.
export function releaseAttachmentUrl(url: string | null | undefined) {
	if (url?.startsWith('blob:')) URL.revokeObjectURL(url)
}

export async function downloadAttachment(attachment: TaskAttachment) {
	const url = await fetchAttachmentUrl({id: attachment.id!, task_id: attachment.task_id!})
	downloadBlob(url, attachment.file?.name ?? '')
}

export function generateAttachmentUrl(taskId: number, attachmentId: number) {
	return `${window.API_URL}/tasks/${taskId}/attachments/${attachmentId}`
}
