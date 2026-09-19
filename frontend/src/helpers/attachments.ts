import type {QueryKey} from '@tanstack/vue-query'
import type {TaskAttachment} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {
	attachmentBlob,
	attachmentKeys,
	uploadAttachmentsMutationOptions,
	type AttachmentIdentity,
	type PreviewSize,
} from '@/client/queries/attachments'
import {captureClientRequestContext, assertClientRequestContext} from '@/client/requestContext'
import {downloadBlob} from '@/helpers/downloadBlob'

const isAttachmentBlobKey = (queryKey: QueryKey) =>
	attachmentKeys.blobs.every((segment, index) => queryKey[index] === segment)

queryClient.getQueryCache().subscribe(event => {
	if (event.type !== 'removed' || !isAttachmentBlobKey(event.query.queryKey)) {
		return
	}

	const url = event.query.state.data
	if (typeof url === 'string') URL.revokeObjectURL(url)
})

// A blob: url for these inherits our origin and can script; a data: url cannot.
const SCRIPTABLE_MIME_TYPES = new Set([
	'image/svg+xml',
	'text/html',
	'application/xhtml+xml',
	'text/xml',
	'application/xml',
])

export async function attachmentBlobUrl(attachment: AttachmentIdentity, size?: PreviewSize, signal?: AbortSignal) {
	const context = captureClientRequestContext()
	const blob = await attachmentBlob(attachment, size, signal)
	assertClientRequestContext(context)
	signal?.throwIfAborted()
	const mimeType = blob.type.split(';')[0].trim().toLowerCase()
	// FileReader is absent in iOS Lockdown Mode and some webviews, fall back to a blob url there.
	if (SCRIPTABLE_MIME_TYPES.has(mimeType) && typeof FileReader !== 'undefined') {
		return new Promise<string>((resolve, reject) => {
			const reader = new FileReader()
			reader.onload = () => typeof reader.result === 'string'
				? resolve(reader.result)
				: reject(new Error('Attachment could not be read as a data url'))
			reader.onerror = () => reject(reader.error ?? new Error('Attachment could not be read as a data url'))
			reader.readAsDataURL(blob)
		})
	}
	return URL.createObjectURL(blob)
}

// Shared editor/board URLs live until their query is evicted; callers must not revoke them.
export function fetchAttachmentBlobUrl(attachment: AttachmentIdentity, size?: PreviewSize): Promise<string> {
	return queryClient.fetchQuery({
		queryKey: attachmentKeys.blob(attachment.task_id, attachment.id, size),
		queryFn: ({signal}) => attachmentBlobUrl(attachment, size, signal),
		staleTime: Infinity,
		// fetchQuery registers no observer, so the default gcTime would revoke a url the editor or a cover still shows
		gcTime: Infinity,
		retry: false,
	})
}

export async function downloadAttachment(attachment: TaskAttachment) {
	const url = await attachmentBlobUrl({id: attachment.id!, task_id: attachment.task_id!})
	downloadBlob(url, attachment.file?.name ?? '')
}

export async function uploadFile(
	taskId: number,
	file: File,
	onSuccess?: (url: string) => void,
): Promise<TaskAttachment[]> {
	// the editor callers toast the rejection themselves, a mutation toast would duplicate it
	const options = uploadAttachmentsMutationOptions(() => false)
	const result = await queryClient.getMutationCache().build(queryClient, options)
		.execute({taskId, files: [file]})
	const uploaded = result.success ?? []
	for (const attachment of uploaded) onSuccess?.(generateAttachmentUrl(taskId, attachment.id!))
	if (result.errors?.length) throw new Error(result.errors.map(error => error.message).join('\n'))
	return uploaded
}

/**
 * Uploads report each finished file through a callback rather than their return
 * value, so the urls have to be collected there. The rejection still has to be
 * forwarded, else a failed upload leaves a dangling rejected promise behind.
 */
export function uploadFilesForEditor(
	upload: (file: File, onSuccess: (attachmentUrl: string) => void) => Promise<unknown>,
	files: File[] | FileList,
): Promise<string[]> {
	return Promise.all(Array.from(files).map(file => new Promise<string>((resolve, reject) => {
		upload(file, resolve).catch(reject)
	})))
}

export function generateAttachmentUrl(taskId: number, attachmentId: number) {
	return `${window.API_URL}/tasks/${taskId}/attachments/${attachmentId}`
}
