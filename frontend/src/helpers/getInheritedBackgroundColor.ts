function getDefaultBackground() {
	const div = document.createElement('div')
	document.head.appendChild(div)
	const bg = window.getComputedStyle(div).backgroundColor
	document.head.removeChild(div)
	return bg
}

// get default style for current browser
const defaultStyle = getDefaultBackground() // typically "rgba(0, 0, 0, 0)"

// based on https://stackoverflow.com/a/62630563/15522256
export function getInheritedBackgroundColor(el: HTMLElement): string {  
	const backgroundColor = window.getComputedStyle(el).backgroundColor

	if (backgroundColor !== defaultStyle) return backgroundColor

	if (!el.parentElement) {
		// we reached the top parent el without getting an explicit color
		return defaultStyle
	}
  
	return getInheritedBackgroundColor(el.parentElement)
}
