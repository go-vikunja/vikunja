import {useMutation} from '@tanstack/vue-query'
import {
	taskAttachmentsDelete,
	taskAttachmentsDownload,
	taskAttachmentsUpload,
} from '@/client/generated'
import type {
	TaskAttachment,
	TaskAttachmentsDownloadData,
} from '@/client/generated'
import {i18n} from '@/i18n'
import {contextMutationOptions} from './contextMutation'
import {
	invalidateTaskMembership,
	mapTaskEverywhere,
} from './taskCache'

// The task detail and every list response already carry `attachments`, so there is no list query.
export const attachmentKeys = {
	blobs: ['attachments', 'blob'] as const,
	blobsFor: (taskId: number, id: number) => [...attachmentKeys.blobs, taskId, id] as const,
	blob: (taskId: number, id: number, size?: PreviewSize) => [...attachmentKeys.blobsFor(taskId, id), size] as const,
}

export type PreviewSize = NonNullable<TaskAttachmentsDownloadData['query']>['preview_size']
export type AttachmentIdentity = Required<Pick<TaskAttachment, 'id' | 'task_id'>>

export async function attachmentBlob(attachment: AttachmentIdentity, size?: PreviewSize, signal?: AbortSignal) {
	const {data} = await taskAttachmentsDownload({
		path: {
			task: attachment.task_id,
			attachment: attachment.id,
		},
		query: {preview_size: size},
		parseAs: 'blob',
		signal,
	})
	if (!(data instanceof Blob)) throw new Error('Attachment response was not a blob')
	return data
}

export function uploadAttachmentsMutationOptions(shouldNotify: () => boolean = () => true) {
	return contextMutationOptions({
		mutationFn: async ({taskId, files}: {
			taskId: number,
			files: File[],
		}) =>
			(await taskAttachmentsUpload({
				path: {task: taskId},
				body: {files},
			})).data,
		onSuccess: (result, {taskId}, client) => {
			const uploaded = result.success ?? []
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				attachments: [
					...task.attachments.filter(item => !uploaded.some(attachment => attachment.id === item.id)),
					...uploaded,
				],
			}))
		},
		onSettled: ({taskId}, client) => invalidateTaskMembership(client, taskId),
		toastError: shouldNotify,
	})
}

export function deleteAttachmentMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, id}: {
			taskId: number,
			id: number,
		}) =>
			(await taskAttachmentsDelete({path: {
				task: taskId,
				attachment: id,
			}})).data,
		onSuccess: (_data, {taskId, id}, client) => {
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				attachments: task.attachments.filter(a => a.id !== id),
				cover_image_attachment_id: task.cover_image_attachment_id === id ? 0 : task.cover_image_attachment_id,
			}))
			client.removeQueries({queryKey: attachmentKeys.blobsFor(taskId, id)})
		},
		onSettled: ({taskId}, client) => invalidateTaskMembership(client, taskId),
		successMessage: () => i18n.global.t('task.attachment.deleteSuccess'),
	})
}

export function useUploadAttachmentsMutation(shouldNotify?: () => boolean) {
	return useMutation(uploadAttachmentsMutationOptions(shouldNotify))
}

export function useDeleteAttachmentMutation() {
	return useMutation(deleteAttachmentMutationOptions())
}
