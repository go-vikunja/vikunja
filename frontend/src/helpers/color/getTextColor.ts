import { colorIsDark } from './colorIsDark'

export const LIGHT = 'hsl(220, 13%, 91%)' // grey-200
export const DARK = 'hsl(215, 27.9%, 16.9%)' // grey-800

export function getTextColor(backgroundColor: string) {
	return colorIsDark(backgroundColor)
		// Fixed colors to avoid flipping in dark mode
		? DARK
		: LIGHT
}
