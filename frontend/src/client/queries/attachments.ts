import {
	queryOptions,
	useMutation,
} from '@tanstack/vue-query'
import {
	taskAttachmentsDelete,
	taskAttachmentsDownload,
	taskAttachmentsList,
	taskAttachmentsUpload,
} from '@/client/generated'
import type {
	TaskAttachment,
	TaskAttachmentsDownloadData,
} from '@/client/generated'
import {i18n} from '@/i18n'
import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {
	invalidateTaskMembership,
	mapTaskEverywhere,
} from './taskCache'

export const attachmentKeys = {
	all: ['attachments'] as const,
	list: (taskId: number) => ['attachments', 'list', taskId] as const,
	blobs: ['attachments', 'blob'] as const,
	blobsFor: (taskId: number, id: number) => [...attachmentKeys.blobs, taskId, id] as const,
	blob: (taskId: number, id: number, size?: PreviewSize) => [...attachmentKeys.blobsFor(taskId, id), size] as const,
}

export type PreviewSize = NonNullable<TaskAttachmentsDownloadData['query']>['preview_size']
export type AttachmentIdentity = Required<Pick<TaskAttachment, 'id' | 'task_id'>>

export function attachmentsQuery(taskId: number) {
	return queryOptions({
		queryKey: attachmentKeys.list(taskId),
		enabled: taskId > 0,
		queryFn: ({signal}) => fetchAllPages(page => taskAttachmentsList({
			path: {task: taskId},
			query: {
				page,
				per_page: 1000,
			},
			signal,
		}).then(({data}) => data)),
	})
}

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
			const append = (current: TaskAttachment[]) => [
				...current.filter(item => !uploaded.some(attachment => attachment.id === item.id)),
				...uploaded,
			]
			client.setQueryData<TaskAttachment[]>(attachmentKeys.list(taskId), current => current && append(current))
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				attachments: append(task.attachments),
			}))
		},
		onSettled: ({taskId}, client) => Promise.all([
			client.invalidateQueries({queryKey: attachmentKeys.list(taskId)}),
			invalidateTaskMembership(client, taskId),
		]),
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
			client.setQueryData<TaskAttachment[]>(attachmentKeys.list(taskId), current => current?.filter(a => a.id !== id))
			mapTaskEverywhere(client, taskId, task => ({
				...task,
				attachments: task.attachments.filter(a => a.id !== id),
				cover_image_attachment_id: task.cover_image_attachment_id === id ? 0 : task.cover_image_attachment_id,
			}))
			client.removeQueries({queryKey: attachmentKeys.blobsFor(taskId, id)})
		},
		onSettled: ({taskId}, client) => Promise.all([
			client.invalidateQueries({queryKey: attachmentKeys.list(taskId)}),
			invalidateTaskMembership(client, taskId),
		]),
		successMessage: () => i18n.global.t('task.attachment.deleteSuccess'),
	})
}

export function useUploadAttachmentsMutation() {
	return useMutation(uploadAttachmentsMutationOptions())
}

export function useDeleteAttachmentMutation() {
	return useMutation(deleteAttachmentMutationOptions())
}
