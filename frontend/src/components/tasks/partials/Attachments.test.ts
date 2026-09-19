import {describe, it, expect, vi, afterEach} from 'vitest'
import {nextTick} from 'vue'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import Modal from '@/components/misc/Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/Button.vue'
import type {TaskAttachment as IAttachment} from '@/client/generated'
import type {Task as ITask} from '@/client/generated'

const sdk = vi.hoisted(() => ({
	patchTasksRead: vi.fn(),
	taskAttachmentsList: vi.fn(async () => ({data: {items: [attachment], total_pages: 1}})),
	taskAttachmentsUpload: vi.fn(),
}))

const {errorMessage} = vi.hoisted(() => ({errorMessage: vi.fn()}))

vi.mock('@/client/generated', () => sdk)

vi.mock('vue-i18n', async importOriginal => ({
	...(await importOriginal<typeof import('vue-i18n')>()),
	useI18n: () => ({t: (key: string) => key}),
}))

vi.mock('@/message', () => ({error: errorMessage, success: vi.fn()}))

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

function mountAttachments() {
	const wrapper = mount(Attachments, {
		attachTo: document.body,
		props: {task},
		global: {
			plugins: [[VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}]],
			components: {XButton, BaseButton, Modal},
			stubs: {
				Icon: true,
				User: true,
				FilePreview: true,
				AudioPreview: true,
				ProgressBar: true,
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
	sdk.taskAttachmentsUpload.mockReset()
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
