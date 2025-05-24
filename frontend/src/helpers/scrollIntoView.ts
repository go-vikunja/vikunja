export function scrollIntoView(el: HTMLElement | null | undefined) {
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
			behavior: 'smooth',
			block: 'center',
			inline: 'nearest',
		})
	}
}
