import {describe, it, expect, vi, afterEach} from 'vitest'
import {nextTick} from 'vue'
import {mount, type VueWrapper} from '@vue/test-utils'
import Attachments from './Attachments.vue'
import Modal from '@/components/misc/Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/Button.vue'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {ITask} from '@/modelTypes/ITask'

vi.mock('@/services/attachment', () => ({
	default: class {
		loading = false
		uploadProgress = 0
	},
}))

vi.mock('@/stores/tasks', () => ({useTaskStore: () => ({isLoading: false})}))
vi.mock('@/stores/auth', () => ({useAuthStore: () => ({info: {id: 1}, limit: () => null, usage: () => 0, adjustUsage: vi.fn()})}))
vi.mock('@/stores/projects', () => ({useProjectStore: () => ({projects: {}})}))

vi.mock('vue-i18n', async importOriginal => ({
	...(await importOriginal<typeof import('vue-i18n')>()),
	useI18n: () => ({t: (key: string) => key}),
}))

vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn(), upgradeActions: () => []}))

const attachment = {
	id: 1,
	taskId: 1,
	file: {name: 'invoice.pdf', size: 1234, mime: 'application/pdf'},
	created: new Date(),
	createdBy: {id: 1, username: 'demo'},
} as unknown as IAttachment

const task = {id: 1, attachments: [attachment]} as unknown as ITask

const mounted: VueWrapper[] = []
const renderErrors: unknown[] = []

function mountAttachments() {
	const wrapper = mount(Attachments, {
		attachTo: document.body,
		props: {task},
		global: {
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
})

describe('Attachments delete modal', () => {
	it('does not render the attachment name after the modal was closed', async () => {
		const wrapper = mountAttachments()

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
