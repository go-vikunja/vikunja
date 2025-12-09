/**
 * Calls the close callback when a click happened outside of the rootElement.
 *
 * @param event The "click" event object.
 * @param rootElement
 * @param closeCallback A closure function to call when the click event happened outside of the rootElement.
 */
export const closeWhenClickedOutside = (event: MouseEvent, rootElement: HTMLElement, closeCallback: () => void) => {
	// Use composedPath() to get the full event path including elements inside Shadow DOM.
	// This ensures clicks inside shadow roots (like emoji-picker-element) are detected correctly.
	const path = event.composedPath()

	if (path.includes(rootElement)) {
		return
	}

	closeCallback()
}
