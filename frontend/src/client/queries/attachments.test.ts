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
	attachmentsQuery,
	deleteAttachmentMutationOptions,
	uploadAttachmentsMutationOptions,
} from './attachments'
import {success} from '@/message'

const sdk = vi.hoisted(() => ({
	taskAttachmentsList: vi.fn(),
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
	it('lists through the generated SDK', async () => {
		sdk.taskAttachmentsList.mockResolvedValue({data: {
			items: [{
				id: 3,
				task_id: 1,
			}],
			total_pages: 1,
		}})
		expect(await client.fetchQuery(attachmentsQuery(1))).toEqual([{
			id: 3,
			task_id: 1,
		}])
		expect(sdk.taskAttachmentsList).toHaveBeenCalledWith({
			path: {task: 1},
			query: {
				page: 1,
				per_page: 1000,
			},
			signal: expect.any(AbortSignal),
		})
	})

	it('keeps partial upload successes in mounted caches', async () => {
		const attachment = {
			id: 3,
			task_id: 1,
		}
		const file = new File(['hello'], 'hello.txt')
		client.setQueryData(attachmentKeys.list(1), [])
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
		expect(client.getQueryData(attachmentKeys.list(1))).toEqual([attachment])
		expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({attachments: [attachment]})
		expect(client.getQueryState(attachmentKeys.list(1))?.isInvalidated).toBe(true)
	})

	it('uploads without creating an absent list', async () => {
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

		expect(client.getQueryData(attachmentKeys.list(1))).toBeUndefined()
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
