import {afterEach, describe, expect, it} from 'vitest'

import {captureFocus} from './useFocusAfterRemoval'

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
