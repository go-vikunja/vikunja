import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {nextTick} from 'vue'
import Modal from './Modal.vue'

const globalMocks = {
	global: {
		mocks: {
			$t: (key: string) => key,
		},
	},
}

// jsdom does not implement HTMLDialogElement.showModal/close.
// Provide stubs so that the [open] attribute — which CSS and our tests
// check — is flipped the same way the real browser would.
let showModalSpy: ReturnType<typeof vi.spyOn>
let closeSpy: ReturnType<typeof vi.spyOn>
let installedShowModal = false
let installedClose = false

beforeEach(() => {
	const proto = HTMLDialogElement.prototype
	if (typeof proto.showModal !== 'function') {
		proto.showModal = function () {}
		installedShowModal = true
	}
	if (typeof proto.close !== 'function') {
		proto.close = function () {}
		installedClose = true
	}
	showModalSpy = vi.spyOn(proto, 'showModal').mockImplementation(function (this: HTMLDialogElement) {
		this.setAttribute('open', '')
	})
	closeSpy = vi.spyOn(proto, 'close').mockImplementation(function (this: HTMLDialogElement) {
		this.removeAttribute('open')
	})
})

afterEach(() => {
	showModalSpy.mockRestore()
	closeSpy.mockRestore()
	// Remove the prototype stubs we installed, so other test files see the
	// original (unpatched) shape of HTMLDialogElement.
	if (installedShowModal) {
		// @ts-expect-error — removing the method we added
		delete HTMLDialogElement.prototype.showModal
		installedShowModal = false
	}
	if (installedClose) {
		// @ts-expect-error — removing the method we added
		delete HTMLDialogElement.prototype.close
		installedClose = false
	}
	document.body.innerHTML = ''
})

describe('Modal.vue — open race condition (#2590)', () => {
	it('opens the dialog when enabled flips false → true', async () => {
		const wrapper = mount(Modal, {
			...globalMocks,
			attachTo: document.body,
			props: {enabled: false},
			slots: {default: '<p class="test-body">hi</p>'},
		})

		// Pre-condition: dialog is not yet in the DOM.
		expect(document.querySelector('dialog.modal-dialog')).toBeNull()

		// Flip enabled → true. This is the failure path in the bug report.
		// The fix must call showModal() deterministically — i.e. once the
		// <dialog> element is mounted via the dialogRef watcher, not via a
		// nextTick that may fire before the mount flush under Electron.
		await wrapper.setProps({enabled: true})
		await flushPromises()
		await nextTick()

		const dialog = document.querySelector('dialog.modal-dialog') as HTMLDialogElement | null
		expect(dialog).not.toBeNull()
		expect(dialog!.hasAttribute('open')).toBe(true)
		expect(showModalSpy).toHaveBeenCalledTimes(1)

		wrapper.unmount()
	})

	it('calls showModal synchronously with the render flush, not via a deferred nextTick (#2590)', async () => {
		// Regression guard: the buggy implementation scheduled showModal() via
		// nextTick *after* setting showDialog = true, so the call landed in a
		// microtask that could fire before the <dialog> mount flush under
		// Electron/Chromium. The fix invokes showModal() from a watch on the
		// dialogRef template ref, which Vue populates during the same flush
		// that mounts the element. That means by the time `await nextTick()`
		// resolves after the first state change, the dialog must already have
		// [open] set — no additional flushPromises or extra ticks required.
		const wrapper = mount(Modal, {
			...globalMocks,
			attachTo: document.body,
			props: {enabled: false},
			slots: {default: '<p class="test-body">hi</p>'},
		})
		expect(document.querySelector('dialog.modal-dialog')).toBeNull()

		// Flip enabled and wait exactly one render flush. After this, the
		// dialog is mounted AND showModal has been called.
		wrapper.setProps({enabled: true})
		await nextTick()

		const dialog = document.querySelector('dialog.modal-dialog') as HTMLDialogElement | null
		expect(dialog).not.toBeNull()
		expect(showModalSpy).toHaveBeenCalled()
		expect(showModalSpy.mock.instances[0]).toBe(dialog)
		expect(dialog!.hasAttribute('open')).toBe(true)

		wrapper.unmount()
	})

	it('calls showModal on the exact dialog element that is mounted (race regression)', async () => {
		// This test asserts the fix's contract: whenever the <dialog> element
		// is mounted (i.e. dialogRef becomes non-null), showModal() is called
		// on *that* element. The buggy implementation instead relied on a
		// nextTick callback whose timing could fire before the dialog mounted,
		// skipping the showModal() call entirely and leaving .open === false.
		const wrapper = mount(Modal, {
			...globalMocks,
			attachTo: document.body,
			props: {enabled: true},
			slots: {default: '<p class="test-body">hi</p>'},
		})
		await flushPromises()
		await nextTick()

		const dialog = document.querySelector('dialog.modal-dialog') as HTMLDialogElement | null
		expect(dialog).not.toBeNull()
		// The fingerprint from the bug report: element is mounted but .open
		// is false because showModal() was never called. The fix guarantees
		// these two always agree.
		expect(dialog!.hasAttribute('open')).toBe(true)
		expect(showModalSpy).toHaveBeenCalled()
		expect(showModalSpy.mock.instances[0]).toBe(dialog)

		wrapper.unmount()
	})

	it('closes the dialog when enabled flips true → false', async () => {
		const wrapper = mount(Modal, {
			...globalMocks,
			attachTo: document.body,
			props: {enabled: true},
			slots: {default: '<p class="test-body">hi</p>'},
		})
		await flushPromises()
		await nextTick()

		// Sanity: open.
		expect(document.querySelector('dialog.modal-dialog')?.hasAttribute('open')).toBe(true)

		await wrapper.setProps({enabled: false})
		// Wait past the 150ms closeTimer (real timers — fake timers interact
		// badly with Vue's scheduler).
		await new Promise(resolve => setTimeout(resolve, 200))
		await flushPromises()
		await nextTick()

		expect(document.querySelector('dialog.modal-dialog')).toBeNull()

		wrapper.unmount()
	})

	it('does not open the dialog if enabled flips back to false before mount', async () => {
		// Regression guard: the dialogRef watcher fires once the <dialog>
		// element mounts. If props.enabled has flipped back to false by the
		// time the mount happens, the watcher must not call showModal().
		const wrapper = mount(Modal, {
			...globalMocks,
			attachTo: document.body,
			props: {enabled: false},
			slots: {default: '<p class="test-body">hi</p>'},
		})

		// Flip enabled true then false within the same tick, before the mount
		// flush can complete.
		wrapper.setProps({enabled: true})
		wrapper.setProps({enabled: false})
		await flushPromises()
		await nextTick()
		await new Promise(resolve => setTimeout(resolve, 200))
		await flushPromises()
		await nextTick()

		// showModal must not have been called — the final prop state is
		// disabled.
		expect(showModalSpy).not.toHaveBeenCalled()
		expect(document.querySelector('dialog.modal-dialog')).toBeNull()

		wrapper.unmount()
	})

	it('clears data-closing when re-opened mid-close transition', async () => {
		// Regression guard: if the user toggles enabled back to true while the
		// 150ms close transition is still in flight, the <dialog> is still
		// mounted and [open], so the dialogRef watcher does not re-fire. Make
		// sure openDialog() clears the leftover data-closing flag itself;
		// otherwise the dialog stays stuck at opacity 0.
		const wrapper = mount(Modal, {
			...globalMocks,
			attachTo: document.body,
			props: {enabled: true},
			slots: {default: '<p class="test-body">hi</p>'},
		})
		await flushPromises()
		await nextTick()

		const dialog = document.querySelector('dialog.modal-dialog') as HTMLDialogElement
		expect(dialog.hasAttribute('open')).toBe(true)

		// Start closing — this sets data-closing and schedules the unmount.
		await wrapper.setProps({enabled: false})
		await nextTick()
		expect(dialog.dataset.closing).toBe('')

		// Re-open well before the 150ms close timer fires.
		await wrapper.setProps({enabled: true})
		await nextTick()

		expect(dialog.dataset.closing).toBeUndefined()
		expect(dialog.hasAttribute('open')).toBe(true)

		wrapper.unmount()
	})
})
