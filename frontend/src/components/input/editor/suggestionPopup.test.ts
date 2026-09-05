import {describe, it, expect, beforeEach, vi} from 'vitest'

type VirtualReference = {getBoundingClientRect: () => DOMRect}

const computePosition = vi.fn((_reference: VirtualReference, _floating: HTMLElement) => Promise.resolve({x: 10, y: 20}))
const stopAutoUpdate = vi.fn()
let autoUpdateCallback: (() => void) | null = null

vi.mock('@floating-ui/dom', () => ({
	computePosition: (reference: VirtualReference, floating: HTMLElement) => computePosition(reference, floating),
	autoUpdate: (_reference: unknown, _floating: unknown, update: () => void) => {
		autoUpdateCallback = update
		update()
		return stopAutoUpdate
	},
	flip: () => ({}),
	offset: () => ({}),
	shift: () => ({}),
}))

import {createSuggestionPopup} from './suggestionPopup'

const rect = () => new DOMRect(1, 2, 3, 4)

function create() {
	const content = document.createElement('span')
	return createSuggestionPopup(document.body, content, rect())
}

describe('createSuggestionPopup', () => {
	beforeEach(() => {
		document.body.innerHTML = ''
		autoUpdateCallback = null
		computePosition.mockClear()
		stopAutoUpdate.mockClear()
	})

	it('mounts the content into the container and positions it', async () => {
		const popup = create()

		expect(popup.element.parentElement).toBe(document.body)
		expect(computePosition).toHaveBeenCalledTimes(1)

		await vi.waitFor(() => expect(popup.element.style.left).toBe('10px'))
		expect(popup.element.style.top).toBe('20px')
	})

	it('removes the element and stops autoUpdate on destroy', () => {
		const popup = create()

		popup.destroy()

		expect(stopAutoUpdate).toHaveBeenCalledTimes(1)
		expect(popup.element.parentElement).toBeNull()
	})

	it('does not measure anything when autoUpdate fires after destroy', () => {
		const popup = create()
		computePosition.mockClear()

		popup.destroy()
		autoUpdateCallback?.()

		expect(computePosition).not.toHaveBeenCalled()
	})

	it('ignores a destroy that happens while a position is being computed', async () => {
		const popup = create()

		popup.destroy()
		await vi.waitFor(() => expect(computePosition).toHaveBeenCalled())

		expect(popup.element.style.left).toBe('0px')
	})

	it('is idempotent on repeated destroy', () => {
		const popup = create()

		popup.destroy()
		popup.destroy()

		expect(stopAutoUpdate).toHaveBeenCalledTimes(1)
	})

	it('uses the latest reference rect for subsequent updates', () => {
		const popup = create()
		const next = new DOMRect(5, 6, 7, 8)

		popup.setReferenceRect(next)
		autoUpdateCallback?.()

		const reference = computePosition.mock.calls[1][0]
		expect(reference.getBoundingClientRect()).toBe(next)
	})
})
