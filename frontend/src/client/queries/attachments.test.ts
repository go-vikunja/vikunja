import {
	beforeEach,
	describe,
	expect,
	it,
	vi,
} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {kanbanKeys} from './kanban'
import {
	normalizeTask,
	taskKeys,
} from './tasks'
import {
	attachmentKeys,
	deleteAttachmentMutationOptions,
	uploadAttachmentsMutationOptions,
} from './attachments'
import {error, success} from '@/message'

const sdk = vi.hoisted(() => ({
	taskAttachmentsUpload: vi.fn(),
	taskAttachmentsDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

let client: QueryClient
beforeEach(() => {
	vi.clearAllMocks()
	client = new QueryClient({defaultOptions: {queries: {retry: false}}})
})

describe('attachments', () => {
	it('keeps partial upload successes in mounted caches', async () => {
		const attachment = {
			id: 3,
			task_id: 1,
		}
		const file = new File(['hello'], 'hello.txt')
		client.setQueryData(taskKeys.detail(1), normalizeTask({
			id: 1,
			attachments: [],
		}))
		sdk.taskAttachmentsUpload.mockResolvedValue({data: {
			success: [attachment],
			errors: [{message: 'full'}],
		}})
		const result = await client.getMutationCache().build(client, uploadAttachmentsMutationOptions())
			.execute({
				taskId: 1,
				files: [file],
			})
		expect(result.errors).toEqual([{message: 'full'}])
		expect(sdk.taskAttachmentsUpload).toHaveBeenCalledWith({
			path: {task: 1},
			body: {files: [file]},
		})
		expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({attachments: [attachment]})
		expect(client.getQueryState(taskKeys.detail(1))?.isInvalidated).toBe(true)
	})

	it('uploads without creating an absent task detail', async () => {
		const attachment = {
			id: 3,
			task_id: 1,
		}
		sdk.taskAttachmentsUpload.mockResolvedValue({data: {success: [attachment]}})

		await client.getMutationCache().build(client, uploadAttachmentsMutationOptions())
			.execute({
				taskId: 1,
				files: [new File(['hello'], 'hello.txt')],
			})

		expect(client.getQueryData(taskKeys.detail(1))).toBeUndefined()
	})

	it('toasts a failed upload once through the default options', async () => {
		const failed = new Error('failed to save file: no space left on device')
		sdk.taskAttachmentsUpload.mockRejectedValue(failed)

		const mutation = client.getMutationCache().build(client, uploadAttachmentsMutationOptions())

		await expect(mutation.execute({
			taskId: 1,
			files: [new File([''], 'a.png')],
		})).rejects.toThrow(failed)
		expect(error).toHaveBeenCalledExactlyOnceWith(failed)
	})

	it('leaves the toast to the editor when notifications are suppressed', async () => {
		sdk.taskAttachmentsUpload.mockRejectedValue(new Error('failed to save file: no space left on device'))

		const mutation = client.getMutationCache().build(client, uploadAttachmentsMutationOptions(() => false))

		await expect(mutation.execute({
			taskId: 1,
			files: [new File([''], 'a.png')],
		})).rejects.toThrow('failed to save file: no space left on device')
		expect(error).not.toHaveBeenCalled()
	})

	it('deletes an attachment and clears its cover without creating absent lists', async () => {
		client.setQueryData(taskKeys.detail(1), normalizeTask({
			id: 1,
			attachments: [{id: 3}],
			cover_image_attachment_id: 3,
		}))
		sdk.taskAttachmentsDelete.mockResolvedValue({data: {message: 'deleted'}})
		const boardKey = [...kanbanKeys.all, 'test-board']
		client.setQueryData(boardKey, {buckets: []})
		client.setQueryData(attachmentKeys.blob(1, 3), 'blob:original')
		client.setQueryData(attachmentKeys.blob(1, 3, 'md'), 'blob:md')
		client.setQueryData(attachmentKeys.blob(1, 4), 'blob:other')
		const before = client.getQueryCache().getAll().length
		await client.getMutationCache().build(client, deleteAttachmentMutationOptions())
			.execute({
				taskId: 1,
				id: 3,
			})
		expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({
			attachments: [],
			cover_image_attachment_id: 0,
		})
		expect(client.getQueryData(attachmentKeys.blob(1, 3))).toBeUndefined()
		expect(client.getQueryData(attachmentKeys.blob(1, 3, 'md'))).toBeUndefined()
		expect(client.getQueryData(attachmentKeys.blob(1, 4))).toBe('blob:other')
		expect(client.getQueryCache().getAll()).toHaveLength(before - 2)
		expect(client.getQueryState(boardKey)?.isInvalidated).toBe(true)
		expect(success).toHaveBeenCalledWith({message: 'The attachment was successfully deleted.'})
	})
})
