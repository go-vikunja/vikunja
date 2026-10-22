import type {Label} from '@/client/generated'
import {getTextColor} from '@/helpers/color/getTextColor'

export function getLabelColor(label: Label): string {
	const color = label.hex_color ?? ''
	if (color === '' || color.startsWith('#') || color.startsWith('var(')) {
		return color
	}

	return `#${color}`
}

export function useLabelStyles() {
	function getLabelStyles(label: Label) {
		const color = getLabelColor(label)
		return {
			'background': color || 'var(--grey-200)',
			'color': color ? getTextColor(color) : 'var(--grey-800)',
		}
	}

	return {
		getLabelStyles,
	}
}
