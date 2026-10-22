import { colorIsDark } from './colorIsDark'

// Fixed colors (not grey-200/grey-800) to guarantee 4.5:1 contrast at the luminance flip point
export const LIGHT = '#fff'
export const DARK = '#000'

export function getTextColor(backgroundColor: string) {
	return colorIsDark(backgroundColor)
		? DARK
		: LIGHT
}
