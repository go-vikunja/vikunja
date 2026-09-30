import type {Directive} from 'vue'

const FOCUSABLE_TAGS = ['INPUT', 'SELECT', 'TEXTAREA']

function isFocusable(el: HTMLElement): boolean {
	return FOCUSABLE_TAGS.includes(el.tagName) || el.isContentEditable
}

function getFocusableElement(el: HTMLElement): HTMLElement | null {
	if (isFocusable(el)) {
		return el
	}
	// Look for the first focusable child
	return el.querySelector<HTMLElement>('input, select, textarea, [contenteditable="true"]')
}

export const isAutoFocusViewport = () => window.innerWidth > 769

const focus = <Directive<HTMLElement, string>>{
	// When the bound element is inserted into the DOM...
	mounted(el, {modifiers}) {
		// Focus the element only if the viewport is big enough
		// auto focusing elements on mobile can be annoying since in these cases the
		// keyboard always pops up and takes half of the available space on the screen.
		// The threshhold is the same as the breakpoints in css.
		if (isAutoFocusViewport() || modifiers?.always) {
			const target = getFocusableElement(el)
			target?.focus()
		}
	},
}

export default focus
