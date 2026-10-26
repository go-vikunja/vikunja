import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {defineComponent, h} from 'vue'
import Notifications from '@kyvg/vue3-notification'

import Notification from './Notification.vue'
import {ACTION_TOAST_DURATION, success, type Action, type ToastOptions} from '@/message'

const XButton = defineComponent({
	emits: ['click'],
	setup(_, {emit, slots}) {
		return () => h('button', {
			type: 'button',
			onClick: () => emit('click'),
		}, slots.default?.())
	},
})

let wrapper: VueWrapper | null = null

function mountToasts() {
	wrapper = mount(Notification, {
		global: {
			plugins: [Notifications],
			components: {XButton},
		},
	})
}

function addButton(label: string) {
	const element = document.createElement('button')
	element.textContent = label
	document.body.append(element)
	return element
}

function toast() {
	return document.querySelector<HTMLElement>('.vue-notification')
}

function toastButton(label: string) {
	return [...toast()?.querySelectorAll('button') ?? []].find(button => button.textContent === label) ?? null
}

async function showToast(options: ToastOptions = {}, actions: Action[] = [{
	title: 'Undo',
	callback: () => {},
}]) {
	success({message: 'Deleted'}, actions, options)
	await flushPromises()
	vi.advanceTimersToNextFrame()
}

describe('Notification', () => {
	beforeEach(() => {
		// setImmediate stays real so flushPromises works
		vi.useFakeTimers({
			toFake: [
				'setTimeout',
				'clearTimeout',
				'requestAnimationFrame',
				'cancelAnimationFrame',
				'Date',
			],
		})
		mountToasts()
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = null
		vi.useRealTimers()
		document.body.innerHTML = ''
	})

	it('leaves focus on the body without takeFocus', async () => {
		await showToast()

		expect(toastButton('Undo')).not.toBeNull()
		expect(document.activeElement).toBe(document.body)
	})

	it('focuses the action when takeFocus finds the focus on the body', async () => {
		await showToast({takeFocus: true})

		expect(document.activeElement).toBe(toastButton('Undo'))
	})

	it('does not take focus the user holds elsewhere', async () => {
		const elsewhere = addButton('Search')
		elsewhere.focus()

		await showToast({takeFocus: true})

		expect(document.activeElement).toBe(elsewhere)
	})

	it('stays open past its duration while focused and closes once focus leaves', async () => {
		const elsewhere = addButton('Search')
		await showToast({takeFocus: true})

		vi.advanceTimersByTime(ACTION_TOAST_DURATION * 2)
		await flushPromises()
		expect(toast()).not.toBeNull()

		elsewhere.focus()
		vi.advanceTimersByTime(ACTION_TOAST_DURATION)
		await flushPromises()
		expect(toast()).toBeNull()
	})

	it('keeps the focus hold when focus moves within the toast', async () => {
		await showToast({takeFocus: true}, [
			{
				title: 'Undo',
				callback: () => {},
			},
			{
				title: 'Open',
				callback: () => {},
			},
		])
		const undo = toastButton('Undo')!
		const open = toastButton('Open')!

		undo.dispatchEvent(new FocusEvent('focusout', {
			bubbles: true,
			relatedTarget: open,
		}))
		vi.advanceTimersByTime(ACTION_TOAST_DURATION * 2)
		await flushPromises()
		expect(toast()).not.toBeNull()

		undo.dispatchEvent(new FocusEvent('focusout', {
			bubbles: true,
			relatedTarget: null,
		}))
		vi.advanceTimersByTime(ACTION_TOAST_DURATION)
		await flushPromises()
		expect(toast()).toBeNull()
	})

	it('hands focus to the main content when a focused toast closes', async () => {
		const main = document.createElement('main')
		main.id = 'main-content'
		main.tabIndex = -1
		document.body.append(main)
		const undo = vi.fn()
		await showToast({takeFocus: true}, [{
			title: 'Undo',
			callback: undo,
		}])

		toastButton('Undo')!.click()
		await flushPromises()

		expect(undo).toHaveBeenCalledOnce()
		expect(toast()).toBeNull()
		expect(document.activeElement).toBe(main)
	})
})
