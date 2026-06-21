import type {Directive, DirectiveBinding} from 'vue'
import {vTooltip} from 'floating-vue'

// When a tooltip target lives inside a <dialog> opened via showModal(), the
// dialog is in the browser's top layer. floating-vue teleports tooltips to
// <body> by default, so they render *below* the dialog's ::backdrop and are
// not visible. Teleporting them into the dialog keeps them in the top layer.
function buildBinding(el: Element, binding: DirectiveBinding): DirectiveBinding {
	const dialog = el.closest('dialog')
	if (!dialog) {
		return binding
	}

	const value = binding.value
	let normalized: Record<string, unknown>
	if (typeof value === 'string') {
		normalized = {content: value}
	} else if (value && typeof value === 'object') {
		normalized = {...value as Record<string, unknown>}
	} else {
		return binding
	}

	if (normalized.container === undefined) {
		normalized.container = dialog
	}

	return {...binding, value: normalized}
}

// Bind via `mounted` rather than `beforeMount` so the element is already
// attached to the DOM — otherwise `el.closest('dialog')` cannot find the
// dialog ancestor.
const tooltip: Directive<Element, unknown> = {
	mounted(el, binding) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		;(vTooltip as any).beforeMount(el, buildBinding(el, binding))
	},
	updated(el, binding) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		;(vTooltip as any).updated(el, buildBinding(el, binding))
	},
	beforeUnmount(el) {
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		;(vTooltip as any).beforeUnmount(el)
	},
}

export default tooltip
