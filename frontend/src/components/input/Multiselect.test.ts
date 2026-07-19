import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, type VueWrapper} from '@vue/test-utils'
import {nextTick} from 'vue'
import Multiselect from './Multiselect.vue'

const searchResults = [
	{title: 'Alpha'},
	{title: 'Beta'},
]

function mountMultiselect() {
	return mount(Multiselect, {
		attachTo: document.body,
		props: {
			modelValue: [],
			searchResults,
			multiple: true,
			label: 'title',
			placeholder: 'Type to search',
			createPlaceholder: 'create',
			selectPlaceholder: 'select',
		},
		global: {
			mocks: {$t: (key: string) => key},
		},
	})
}

// Types into the input and advances past the search debounce + focus timeouts so the result list renders.
async function openResults(wrapper: VueWrapper) {
	const input = wrapper.find('input[role="combobox"]')
	await input.setValue('a')
	await input.trigger('keyup')
	vi.advanceTimersByTime(300)
	await nextTick()
	return input
}

function dispatchEscape(el: Element) {
	const event = new KeyboardEvent('keydown', {key: 'Escape', bubbles: true, cancelable: true})
	const preventDefault = vi.spyOn(event, 'preventDefault')
	const stopPropagation = vi.spyOn(event, 'stopPropagation')
	el.dispatchEvent(event)
	return {event, preventDefault, stopPropagation}
}

describe('Multiselect.vue — combobox Escape semantics', () => {
	beforeEach(() => {
		vi.useFakeTimers()
	})

	afterEach(() => {
		vi.useRealTimers()
		document.body.innerHTML = ''
	})

	it('Escape on the input closes the open list and stops the event, and it stays closed', async () => {
		const wrapper = mountMultiselect()
		const input = await openResults(wrapper)
		expect(wrapper.find('[role="listbox"]').exists()).toBe(true)

		const {preventDefault, stopPropagation} = dispatchEscape(input.element)
		await nextTick()

		expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
		expect(input.attributes('aria-expanded')).toBe('false')
		expect(preventDefault).toHaveBeenCalled()
		expect(stopPropagation).toHaveBeenCalled()

		// The keyup of the same Escape must not re-trigger a search, and the focus
		// timeout must not reopen the list.
		await input.trigger('keyup', {key: 'Escape'})
		vi.advanceTimersByTime(300)
		await nextTick()
		expect(wrapper.find('[role="listbox"]').exists()).toBe(false)

		wrapper.unmount()
	})

	it('Escape on a focused result option closes the list, refocuses the input, and it stays closed', async () => {
		const wrapper = mountMultiselect()
		const input = await openResults(wrapper)

		const option = wrapper.find('[role="option"]').element as HTMLElement
		option.focus()
		expect(document.activeElement).toBe(option)

		dispatchEscape(option)
		await nextTick()

		expect(wrapper.find('[role="listbox"]').exists()).toBe(false)
		expect(document.activeElement).toBe(input.element)

		// Regression: refocusing the input fires @focus, whose 10ms timeout would
		// otherwise reopen the just-closed list. suppressFocusOpen must prevent it.
		vi.advanceTimersByTime(300)
		await nextTick()
		expect(wrapper.find('[role="listbox"]').exists()).toBe(false)

		wrapper.unmount()
	})

	it('Escape on the input with the list closed lets the event through to ancestors', async () => {
		const wrapper = mountMultiselect()
		const input = wrapper.find('input[role="combobox"]')
		expect(wrapper.find('[role="listbox"]').exists()).toBe(false)

		const ancestorListener = vi.fn()
		document.addEventListener('keydown', ancestorListener)

		const {event, preventDefault, stopPropagation} = dispatchEscape(input.element)
		await nextTick()

		expect(preventDefault).not.toHaveBeenCalled()
		expect(stopPropagation).not.toHaveBeenCalled()
		expect(event.defaultPrevented).toBe(false)
		expect(ancestorListener).toHaveBeenCalledTimes(1)

		document.removeEventListener('keydown', ancestorListener)
		wrapper.unmount()
	})
})
