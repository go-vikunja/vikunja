/**
 * Calls the close callback when a click happened outside of the rootElement.
 *
 * @param event The "click" event object.
 * @param rootElement
 * @param closeCallback A closure function to call when the click event happened outside of the rootElement.
 */
export const closeWhenClickedOutside = (event: MouseEvent, rootElement: HTMLElement, closeCallback: () => void) => {
	// We walk up the tree to see if any parent of the clicked element is the root element.
	// If it is not, we call the close callback. We're doing all this hassle to only call the
	// closing callback when a click happens outside of the rootElement.
	let parent = (event.target as HTMLElement)?.parentElement
	while (parent !== rootElement) {
		if (parent === null || parent.parentElement === null) {
			parent = null
			break
		}

		parent = parent.parentElement
	}

	if (parent === rootElement) {
		return
	}

	closeCallback()
}
