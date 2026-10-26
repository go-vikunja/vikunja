import {afterEach, beforeEach, describe, expect, it} from 'vitest'
import {flushPromises} from '@vue/test-utils'
import {effectScope, ref} from 'vue'

import {captureFocus, useFocusAfterRemoval} from './useFocusAfterRemoval'

function button(label: string) {
	const element = document.createElement('button')
	element.textContent = label
	document.body.append(element)
	return element
}

describe('captureFocus', () => {
	afterEach(() => {
		document.body.innerHTML = ''
	})

	it('moves focus on from a trigger that is still focused', async () => {
		const trigger = button('Empty trash')
		const target = button('Empty state')
		trigger.focus()

		await captureFocus()(() => target)

		expect(document.activeElement).toBe(target)
	})

	it('recovers focus that dropped to the body with a removed trigger', async () => {
		const trigger = button('Restore')
		const target = button('Next restore')
		trigger.focus()
		const moveFocus = captureFocus()
		trigger.remove()

		await moveFocus(() => null, () => target)

		expect(document.activeElement).toBe(target)
	})

	it('leaves focus the user moved elsewhere in the meantime', async () => {
		const trigger = button('Restore')
		const elsewhere = button('Search')
		const target = button('Next restore')
		trigger.focus()
		const moveFocus = captureFocus()
		elsewhere.focus()

		await moveFocus(() => target)

		expect(document.activeElement).toBe(elsewhere)
	})
})

describe('useFocusAfterRemoval', () => {
	let scope: ReturnType<typeof effectScope>

	beforeEach(() => {
		scope = effectScope()
	})

	afterEach(() => {
		scope.stop()
		document.body.innerHTML = ''
	})

	function watchNotice(initiallyShown: boolean, target: HTMLElement) {
		const shown = ref(initiallyShown)
		scope.run(() => useFocusAfterRemoval(() => shown.value, () => target))
		return shown
	}

	it('focuses the target once the notice is gone and focus dropped to the body', async () => {
		const target = button('Main content')
		const shown = watchNotice(true, target)

		shown.value = false
		await flushPromises()

		expect(document.activeElement).toBe(target)
	})

	it('leaves focus the user holds elsewhere', async () => {
		const elsewhere = button('Search')
		const target = button('Main content')
		const shown = watchNotice(true, target)
		elsewhere.focus()

		shown.value = false
		await flushPromises()

		expect(document.activeElement).toBe(elsewhere)
	})

	it('does not move focus when the notice appears', async () => {
		const target = button('Main content')
		const shown = watchNotice(false, target)

		shown.value = true
		await flushPromises()

		expect(document.activeElement).toBe(document.body)
	})
})
