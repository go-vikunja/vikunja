export function getScrollParent(element: HTMLElement): HTMLElement {
	let el: HTMLElement | null = element

	while (el) {
		const overflowY = getComputedStyle(el).overflowY
		if (['auto', 'scroll', 'overlay'].includes(overflowY)) {
			return el
		}
		el = el.parentElement
	}

	return (document.scrollingElement as HTMLElement | null) ?? document.documentElement
}
