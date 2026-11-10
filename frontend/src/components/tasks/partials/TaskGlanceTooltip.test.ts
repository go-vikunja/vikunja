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
import {createTaskDraft} from '@/helpers/task'

const HOVER_DELAY = 1000
const WARM_WINDOW = 300

// The warm state is module-level, so each test starts far after the previous one's hide.
let systemTime = 0

const task = createTaskDraft({
	id: 1,
	title: 'Test task',
	identifier: '#1',
	index: 1,
	description: '',
	attachments: [],
	labels: [],
	due_date: '',
	created: new Date().toISOString(),
	created_by: {id: 1, username: 'test', name: 'Test User'},
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
		systemTime += 1_000_000
		vi.setSystemTime(systemTime)
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

describe('TaskGlanceTooltip.vue — warm state', () => {
	beforeEach(() => {
		vi.useFakeTimers()
		systemTime += 1_000_000
		vi.setSystemTime(systemTime)
	})

	afterEach(() => {
		vi.useRealTimers()
		document.body.innerHTML = ''
	})

	async function showAndHideFirst() {
		const first = mountTooltip()
		const trigger = first.get('.task-glance-trigger')
		await trigger.trigger('mouseenter')
		await advancePastDelay()
		expect(findTooltip()).not.toBeNull()
		await trigger.trigger('mouseleave')
		await nextTick()
		first.unmount()
	}

	it('shows the next tooltip immediately after one was just open', async () => {
		await showAndHideFirst()

		const second = mountTooltip()
		await second.get('.task-glance-trigger').trigger('mouseenter')
		vi.advanceTimersByTime(0)
		await nextTick()
		await nextTick()
		expect(findTooltip()).not.toBeNull()

		second.unmount()
	})

	it('uses the full delay again once the warm window passed', async () => {
		await showAndHideFirst()
		vi.advanceTimersByTime(WARM_WINDOW)

		const second = mountTooltip()
		await second.get('.task-glance-trigger').trigger('mouseenter')
		vi.advanceTimersByTime(HOVER_DELAY - 1)
		await nextTick()
		expect(findTooltip()).toBeNull()

		await advancePastDelay()
		expect(findTooltip()).not.toBeNull()

		second.unmount()
	})

	it('does not warm up when the previous tooltip never showed', async () => {
		const first = mountTooltip()
		const trigger = first.get('.task-glance-trigger')
		await trigger.trigger('mouseenter')
		vi.advanceTimersByTime(HOVER_DELAY / 2)
		await trigger.trigger('mouseleave')
		first.unmount()

		const second = mountTooltip()
		await second.get('.task-glance-trigger').trigger('mouseenter')
		vi.advanceTimersByTime(0)
		await nextTick()
		expect(findTooltip()).toBeNull()

		second.unmount()
	})
})
