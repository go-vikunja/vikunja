export default {
	// When the bound element is inserted into the DOM...
	inserted: (el, {modifiers}) => {
		// Focus the element only if the viewport is big enough
		// auto focusing elements on mobile can be annoying since in these cases the
		// keyboard always pops up and takes half of the available space on the screen.
		// The threshhold is the same as the breakpoints in css.
		if (window.innerWidth > 769 || (typeof modifiers.always !== 'undefined' && modifiers.always)) {
			el.focus()
		}
	},
}
