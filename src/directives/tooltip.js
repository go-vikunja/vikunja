const calculateTop = (coords, tooltip) => {
	// Bottom tooltip use the exact inverse calculation compared to the default.
	if (tooltip.classList.contains('bottom')) {
		return coords.top + tooltip.offsetHeight + 5
	}

	// The top position of the tooltip is the coordinates of the bound element - the height of the tooltip -
	// 5px spacing for the arrow (which is exactly 5px high)
	return coords.top - tooltip.offsetHeight - 5
}

const calculateArrowTop = (top, tooltip) => {
	if (tooltip.classList.contains('bottom')) {
		return `${top - 5}px` // 5px arrow height
	}
	return `${top + tooltip.offsetHeight}px`
}

// This global object holds all created tooltip elements (and their arrows) using the element they were created for as
// key. This allows us to find the tooltip elements if the element the tooltip was created for is unbound so that
// we can remove the tooltip element.
const createdTooltips = {}

export default {
	inserted: (el, {value, modifiers}) => {
		// First, we create the tooltip and arrow elements
		const tooltip = document.createElement('div')
		tooltip.style.position = 'fixed'
		tooltip.innerText = value
		tooltip.classList.add('tooltip')
		const arrow = document.createElement('div')
		arrow.classList.add('tooltip-arrow')
		arrow.style.position = 'fixed'

		if (typeof modifiers.bottom !== 'undefined') {
			tooltip.classList.add('bottom')
			arrow.classList.add('bottom')
		}

		// We don't append the element until hovering over it because that's the most reliable way to determine
		// where the parent elemtent is located at the time the user hovers over it.
		el.addEventListener('mouseover', () => {
			// Appending the element right away because we can only calculate the height of the element if it is
			// already in the DOM.
			document.body.appendChild(tooltip)
			document.body.appendChild(arrow)

			const coords = el.getBoundingClientRect()
			const top = calculateTop(coords, tooltip)
			// The left position of the tooltip is calculated so that the middle point of the tooltip
			// (where the arrow will be) is the middle of the bound element
			const left = coords.left - (tooltip.offsetWidth / 2) + (el.offsetWidth / 2)
			// Now setting all the values
			tooltip.style.top = `${top}px`
			tooltip.style.left = `${coords.left}px`
			tooltip.style.left = `${left}px`

			arrow.style.left = `${left + (tooltip.offsetWidth / 2) - (arrow.offsetWidth / 2)}px`
			arrow.style.top = calculateArrowTop(top, tooltip)

			// And finally make it visible to the user. This will also trigger a nice fade-in animation through
			// css transitions
			tooltip.classList.add('visible')
			arrow.classList.add('visible')
		})

		el.addEventListener('mouseout', () => {
			tooltip.classList.remove('visible')
			arrow.classList.remove('visible')
		})

		createdTooltips[el] = {
			tooltip: tooltip,
			arrow: arrow,
		}
	},
	unbind: el => {
		if (typeof createdTooltips[el] !== 'undefined') {
			createdTooltips[el].tooltip.remove()
			createdTooltips[el].arrow.remove()
		}
	},
}
