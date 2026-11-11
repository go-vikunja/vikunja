import {describe, it, expect, beforeEach, vi} from 'vitest'

type VirtualReference = {getBoundingClientRect: () => DOMRect}

const computePosition = vi.fn((reference: VirtualReference, _floating: HTMLElement) => {
	const {x, y} = reference.getBoundingClientRect()
	return Promise.resolve({x, y})
})
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

let referenceRect: DOMRect | null = new DOMRect(10, 20, 3, 4)

function create() {
	const content = document.createElement('span')
	const contextElement = document.createElement('div')
	return createSuggestionPopup(document.body, content, () => referenceRect, contextElement)
}

describe('createSuggestionPopup', () => {
	beforeEach(() => {
		document.body.innerHTML = ''
		autoUpdateCallback = null
		referenceRect = new DOMRect(10, 20, 3, 4)
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

	it('repositions against the current reference rect', async () => {
		const popup = create()
		await vi.waitFor(() => expect(popup.element.style.left).toBe('10px'))

		referenceRect = new DOMRect(80, 140, 3, 4)
		popup.reposition()

		await vi.waitFor(() => expect(popup.element.style.left).toBe('80px'))
		expect(popup.element.style.top).toBe('140px')
	})

	it('reads the reference rect fresh on every autoUpdate tick', async () => {
		const popup = create()
		await vi.waitFor(() => expect(popup.element.style.left).toBe('10px'))

		referenceRect = new DOMRect(45, 55, 3, 4)
		autoUpdateCallback?.()

		await vi.waitFor(() => expect(popup.element.style.left).toBe('45px'))
	})

	it('falls back to an empty rect when the reference is gone', () => {
		referenceRect = null
		create()

		const reference = computePosition.mock.calls[0][0]
		expect(reference.getBoundingClientRect().x).toBe(0)
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
})
