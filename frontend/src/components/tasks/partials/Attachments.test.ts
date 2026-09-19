import {describe, it, expect, vi, afterEach} from 'vitest'
import {defineComponent, h, nextTick} from 'vue'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {QueryClient, useQuery, VueQueryPlugin} from '@tanstack/vue-query'
import Modal from '@/components/misc/Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/Button.vue'
import type {TaskAttachment as IAttachment} from '@/client/generated'
import type {Task as ITask} from '@/client/generated'

const sdk = vi.hoisted(() => ({
	patchTasksRead: vi.fn(),
	tasksRead: vi.fn(),
	taskAttachmentsUpload: vi.fn(),
	taskAttachmentsDownload: vi.fn(),
	taskAttachmentsDelete: vi.fn(),
}))

const {errorMessage} = vi.hoisted(() => ({errorMessage: vi.fn()}))

vi.mock('@/client/generated', () => sdk)

vi.mock('vue-i18n', async importOriginal => ({
	...(await importOriginal<typeof import('vue-i18n')>()),
	useI18n: () => ({t: (key: string) => key}),
}))

vi.mock('@/message', () => ({error: errorMessage, success: vi.fn()}))

import {normalizeTask, taskKeys, taskQuery} from '@/client/queries/tasks'
import Attachments from './Attachments.vue'

const attachment = {
	id: 1,
	task_id: 1,
	file: {name: 'invoice.pdf', size: 1234, mime: 'application/pdf'},
	created: new Date().toISOString(),
	created_by: {id: 1, username: 'demo'},
} as unknown as IAttachment

const task = {id: 1, attachments: [attachment]} as unknown as ITask

const mounted: VueWrapper[] = []
const renderErrors: unknown[] = []

// Mirrors TaskDetailView: the task detail query owns the attachment list, the component only renders it.
const TaskHost = defineComponent({
	setup() {
		const query = useQuery(taskQuery(1))
		return () => query.data.value === undefined
			? null
			: h(Attachments, {task: query.data.value as ITask})
	},
})

function mountAttachments() {
	const queryClient = new QueryClient({defaultOptions: {queries: {
		retry: false,
		staleTime: Infinity,
	}}})
	queryClient.setQueryData(taskKeys.detail(1), normalizeTask(task))

	const wrapper = mount(TaskHost, {
		attachTo: document.body,
		global: {
			plugins: [[VueQueryPlugin, {queryClient}]],
			components: {XButton, BaseButton, Modal},
			stubs: {
				Icon: true,
				User: true,
				FilePreview: true,
				AudioPreview: true,
				ImageLightbox: true,
				RouterLink: true,
			},
			directives: {tooltip: {}, cy: {}},
			mocks: {$t: (key: string) => key},
			config: {
				errorHandler: (err: unknown) => renderErrors.push(err),
			},
		},
	})
	mounted.push(wrapper)
	return wrapper
}

afterEach(() => {
	mounted.splice(0).forEach(wrapper => wrapper.unmount())
	renderErrors.length = 0
	document.body.innerHTML = ''
	sdk.tasksRead.mockReset()
	sdk.taskAttachmentsUpload.mockReset()
	sdk.taskAttachmentsDownload.mockReset()
	sdk.taskAttachmentsDelete.mockReset()
	errorMessage.mockClear()
})

describe('Attachments delete modal', () => {
	it('does not render the attachment name after the modal was closed', async () => {
		const wrapper = mountAttachments()
		await flushPromises()

		await wrapper.find('.attachment-actions [aria-label="task.attachment.deleteTooltip"]').trigger('click')
		await nextTick()

		expect(document.body.innerHTML).toContain('task.attachment.deleteText1')

		// The modal keeps its slot mounted for the duration of the close
		// transition, so it re-renders once the attachment is already gone.
		wrapper.findComponent(Modal).vm.$emit('close')
		await nextTick()

		expect(renderErrors).toEqual([])
		expect(document.querySelector('dialog.modal-dialog')).not.toBeNull()
		expect(document.body.innerHTML).not.toContain('task.attachment.deleteText1')
	})
})

function dropFile(file: File) {
	const event = new Event('drop', {bubbles: true, cancelable: true})
	Object.defineProperty(event, 'dataTransfer', {
		value: {
			files: [file],
			items: [{kind: 'file', type: file.type}],
			types: ['Files'],
			dropEffect: 'none',
		},
	})
	document.body.dispatchEvent(event)
}

async function pickFiles(wrapper: VueWrapper, files: File[]) {
	const fileInput = wrapper.find('input[type="file"]')
	Object.defineProperty(fileInput.element, 'files', {
		value: files,
		configurable: true,
	})
	await fileInput.trigger('change')
	await flushPromises()
}

describe('Attachments download', () => {
	it('toasts once when the download fails', async () => {
		sdk.taskAttachmentsDownload.mockRejectedValue(new Error('network error'))

		const wrapper = mountAttachments()
		await flushPromises()

		await wrapper.find('.attachment-actions [aria-label="task.attachment.downloadTooltip"]').trigger('click')
		await flushPromises()

		expect(errorMessage).toHaveBeenCalledTimes(1)
	})

	it('does not toast when the download aborts because the session changed', async () => {
		sdk.taskAttachmentsDownload.mockRejectedValue(new DOMException('aborted', 'AbortError'))

		const wrapper = mountAttachments()
		await flushPromises()

		await wrapper.find('.attachment-actions [aria-label="task.attachment.downloadTooltip"]').trigger('click')
		await flushPromises()

		expect(errorMessage).not.toHaveBeenCalled()
	})
})

describe('Attachments upload', () => {
	it('uploads the remaining files after one request fails', async () => {
		sdk.taskAttachmentsUpload
			.mockRejectedValueOnce(new Error('413 Payload Too Large'))
			.mockResolvedValueOnce({data: {success: []}})

		const wrapper = mountAttachments()
		await flushPromises()

		const fileInput = wrapper.find('input[type="file"]')
		Object.defineProperty(fileInput.element, 'files', {
			value: [
				new File(['x'], 'too-big.zip', {type: 'application/zip'}),
				new File(['y'], 'cover-image.png', {type: 'image/png'}),
			],
			configurable: true,
		})

		await fileInput.trigger('change')
		await flushPromises()

		expect(sdk.taskAttachmentsUpload).toHaveBeenCalledTimes(2)
		expect(errorMessage).toHaveBeenCalledTimes(1)
	})

	it('toasts every per-file failure with its code, which the message alone would lose', async () => {
		sdk.taskAttachmentsUpload.mockResolvedValue({data: {errors: [
			{code: 4014},
			{
				code: 4015,
				message: 'file is too large',
			},
		]}})

		const wrapper = mountAttachments()
		await flushPromises()

		const fileInput = wrapper.find('input[type="file"]')
		Object.defineProperty(fileInput.element, 'files', {
			value: [new File(['x'], 'too-big.zip', {type: 'application/zip'})],
			configurable: true,
		})

		await fileInput.trigger('change')
		await flushPromises()

		expect(errorMessage).toHaveBeenCalledTimes(2)
		expect(errorMessage).toHaveBeenNthCalledWith(1, {code: 4014})
		expect(errorMessage).toHaveBeenNthCalledWith(2, {
			code: 4015,
			message: 'file is too large',
		})
	})

	it('shows the bar for the whole of a single-file upload', async () => {
		let finishUpload: (result: unknown) => void = () => {}
		sdk.taskAttachmentsUpload.mockReturnValue(new Promise(resolve => {
			finishUpload = resolve
		}))

		const wrapper = mountAttachments()
		await flushPromises()

		expect(wrapper.find('progress').exists()).toBe(false)

		await pickFiles(wrapper, [new File(['x'], 'cover-image.png', {type: 'image/png'})])

		const bar = wrapper.find('progress')
		expect(bar.exists()).toBe(true)
		expect(bar.attributes('value')).toBeUndefined()
		expect(bar.attributes('aria-label')).toBe('task.attachment.upload')

		finishUpload({data: {success: []}})
		await flushPromises()

		expect(wrapper.find('progress').exists()).toBe(false)
	})

	it('advances the bar by one file as each upload completes', async () => {
		const finishers: ((result: unknown) => void)[] = []
		sdk.taskAttachmentsUpload.mockImplementation(() => new Promise(resolve => {
			finishers.push(resolve)
		}))

		const wrapper = mountAttachments()
		await flushPromises()

		await pickFiles(wrapper, [
			new File(['a'], 'a.png', {type: 'image/png'}),
			new File(['b'], 'b.png', {type: 'image/png'}),
			new File(['c'], 'c.png', {type: 'image/png'}),
			new File(['d'], 'd.png', {type: 'image/png'}),
		])

		expect(wrapper.find('progress').attributes('value')).toBeUndefined()

		for (const completed of [1, 2, 3]) {
			finishers[completed - 1]({data: {success: []}})
			await flushPromises()

			expect(wrapper.find('progress').attributes('value')).toBe(`${completed * 25}`)
		}

		finishers[3]({data: {success: []}})
		await flushPromises()

		expect(wrapper.find('progress').exists()).toBe(false)
	})

	it('renders the uploaded file without refetching the task', async () => {
		sdk.taskAttachmentsUpload.mockResolvedValue({data: {success: [{
			id: 2,
			task_id: 1,
			file: {
				name: 'cover-image.png',
				size: 42,
				mime: 'image/png',
			},
		}]}})

		const wrapper = mountAttachments()
		await flushPromises()

		expect(wrapper.findAll('.attachment')).toHaveLength(1)

		await pickFiles(wrapper, [new File(['x'], 'cover-image.png', {type: 'image/png'})])

		expect(wrapper.findAll('.attachment')).toHaveLength(2)
		expect(wrapper.text()).toContain('cover-image.png')
		expect(sdk.tasksRead).not.toHaveBeenCalled()
	})

	it('drops the deleted row without refetching the task', async () => {
		sdk.taskAttachmentsDelete.mockResolvedValue({data: {message: 'deleted'}})

		const wrapper = mountAttachments()
		await flushPromises()

		await wrapper.find('.attachment-actions [aria-label="task.attachment.deleteTooltip"]').trigger('click')
		await nextTick()
		wrapper.findComponent(Modal).vm.$emit('submit')
		await flushPromises()

		expect(sdk.taskAttachmentsDelete).toHaveBeenCalledWith({path: {
			task: 1,
			attachment: 1,
		}})
		expect(wrapper.findAll('.attachment')).toHaveLength(0)
		expect(sdk.tasksRead).not.toHaveBeenCalled()
	})

	it('ignores a drop while another batch is still uploading', async () => {
		let finishUpload: (result: unknown) => void = () => {}
		sdk.taskAttachmentsUpload.mockReturnValue(new Promise(resolve => {
			finishUpload = resolve
		}))

		const wrapper = mountAttachments()
		await flushPromises()

		const fileInput = wrapper.find('input[type="file"]')
		Object.defineProperty(fileInput.element, 'files', {
			value: [new File(['x'], 'first.png', {type: 'image/png'})],
			configurable: true,
		})
		await fileInput.trigger('change')
		await flushPromises()

		dropFile(new File(['y'], 'second.png', {type: 'image/png'}))
		await flushPromises()

		expect(sdk.taskAttachmentsUpload).toHaveBeenCalledTimes(1)

		finishUpload({data: {success: []}})
		await flushPromises()
	})
})
