import {vi} from 'vitest'

type StubbedMethod = 'showModal' | 'close'

/**
 * happy-dom does not implement HTMLDialogElement.showModal/close, so the [open]
 * attribute never flips the way a real browser would. `restore` must run so
 * other test files see the unpatched prototype.
 */
export function stubDialogMethods() {
	const proto = HTMLDialogElement.prototype
	const installed: StubbedMethod[] = []

	if (typeof proto.showModal !== 'function') {
		proto.showModal = function () {}
		installed.push('showModal')
	}
	if (typeof proto.close !== 'function') {
		proto.close = function () {}
		installed.push('close')
	}

	const showModal = vi.spyOn(proto, 'showModal').mockImplementation(function (this: HTMLDialogElement) {
		this.setAttribute('open', '')
	})
	const close = vi.spyOn(proto, 'close').mockImplementation(function (this: HTMLDialogElement) {
		this.removeAttribute('open')
	})

	return {
		showModal,
		close,
		restore() {
			showModal.mockRestore()
			close.mockRestore()
			installed.forEach(method => Reflect.deleteProperty(proto, method))
			installed.length = 0
		},
	}
}
