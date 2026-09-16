import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount} from '@vue/test-utils'
import {nextTick, ref} from 'vue'

vi.mock('@vueuse/core', async (importOriginal) => ({
	...await importOriginal<typeof import('@vueuse/core')>(),
	useMediaQuery: () => ref(true),
}))

vi.mock('@floating-ui/dom', () => ({
	computePosition: async () => ({x: 0, y: 0}),
	flip: () => ({}),
	offset: () => ({}),
	shift: () => ({}),
}))

// The real one reaches into the auth store for the date display setting.
vi.mock('@/helpers/time/formatDate', () => ({
	formatDisplayDate: () => '2024-01-01',
}))

import TaskGlanceTooltip from './TaskGlanceTooltip.vue'
import TaskModel from '@/models/task'
import UserModel from '@/models/user'

const HOVER_DELAY = 1000

const task = new TaskModel({
	id: 1,
	title: 'Test task',
	identifier: '#1',
	index: 1,
	description: '',
	attachments: [],
	labels: [],
	dueDate: null,
	created: new Date(),
	createdBy: new UserModel({id: 1, username: 'test', name: 'Test User'}),
})

function mountTooltip() {
	return mount(TaskGlanceTooltip, {
		attachTo: document.body,
		props: {task},
		slots: {
			default: '<a href="/tasks/1" id="task-link">Test task</a>',
		},
		global: {
			mocks: {$t: (key: string) => key},
			stubs: {
				Labels: true,
				ChecklistSummary: true,
				CommentCount: true,
				Icon: true,
				'i18n-t': {template: '<span><slot /></span>'},
				CustomTransition: {template: '<div><slot /></div>'},
			},
		},
	})
}

function findTooltip() {
	return document.body.querySelector('[role="tooltip"]')
}

async function advancePastDelay() {
	vi.advanceTimersByTime(HOVER_DELAY)
	await nextTick()
	await nextTick()
}

describe('TaskGlanceTooltip.vue — keyboard accessibility', () => {
	beforeEach(() => {
		vi.useFakeTimers()
	})

	afterEach(() => {
		vi.useRealTimers()
		document.body.innerHTML = ''
	})

	it('shows the tooltip on focus and describes the focused element', async () => {
		const wrapper = mountTooltip()
		const link = wrapper.get('#task-link')

		await link.trigger('focusin')
		expect(findTooltip()).toBeNull()

		await advancePastDelay()

		const tooltip = findTooltip()
		expect(tooltip).not.toBeNull()
		expect(link.attributes('aria-describedby')).toBe(tooltip?.id)
		expect(tooltip?.id).toBeTruthy()

		wrapper.unmount()
	})

	it('hides the tooltip and removes aria-describedby on Escape', async () => {
		const wrapper = mountTooltip()
		const link = wrapper.get('#task-link')

		await link.trigger('focusin')
		await advancePastDelay()
		expect(findTooltip()).not.toBeNull()

		const event = new KeyboardEvent('keydown', {key: 'Escape', bubbles: true, cancelable: true})
		const stopPropagation = vi.spyOn(event, 'stopPropagation')
		document.dispatchEvent(event)
		await nextTick()

		expect(stopPropagation).toHaveBeenCalled()
		expect(findTooltip()).toBeNull()
		expect(link.attributes('aria-describedby')).toBeUndefined()

		wrapper.unmount()
	})

	it('does not swallow Escape when no tooltip is shown', async () => {
		const wrapper = mountTooltip()

		const event = new KeyboardEvent('keydown', {key: 'Escape', bubbles: true, cancelable: true})
		const stopPropagation = vi.spyOn(event, 'stopPropagation')
		document.dispatchEvent(event)
		await nextTick()

		expect(stopPropagation).not.toHaveBeenCalled()

		wrapper.unmount()
	})

	it('hides the tooltip on focusout', async () => {
		const wrapper = mountTooltip()
		const link = wrapper.get('#task-link')

		await link.trigger('focusin')
		await advancePastDelay()
		expect(findTooltip()).not.toBeNull()

		await link.trigger('focusout')
		await nextTick()

		expect(findTooltip()).toBeNull()
		expect(link.attributes('aria-describedby')).toBeUndefined()

		wrapper.unmount()
	})

	it('still shows on hover and hides on mouseleave', async () => {
		const wrapper = mountTooltip()
		const trigger = wrapper.get('.task-glance-trigger')

		await trigger.trigger('mouseenter')
		await advancePastDelay()
		expect(findTooltip()).not.toBeNull()

		await trigger.trigger('mouseleave')
		await nextTick()
		expect(findTooltip()).toBeNull()

		wrapper.unmount()
	})

	it('removes the document keydown listener on unmount', async () => {
		const removeEventListener = vi.spyOn(document, 'removeEventListener')
		const wrapper = mountTooltip()

		await wrapper.get('#task-link').trigger('focusin')
		await advancePastDelay()

		wrapper.unmount()

		expect(removeEventListener).toHaveBeenCalledWith('keydown', expect.any(Function))
		removeEventListener.mockRestore()
	})
})
