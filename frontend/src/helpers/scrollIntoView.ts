export function scrollIntoView(el: HTMLElement | null | undefined, behavior: ScrollBehavior = 'smooth') {
	if (!el) {
		return
	}

	const boundingRect = el.getBoundingClientRect()
	const scrollY = window.scrollY

	if (
		boundingRect.top > (scrollY + window.innerHeight) ||
		boundingRect.top < scrollY
	) {
		el.scrollIntoView({
			behavior,
			block: 'center',
			inline: 'nearest',
		})
	}
}
