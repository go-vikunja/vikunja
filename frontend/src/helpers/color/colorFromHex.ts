/**
 * Returns the hex color code without the '#' if it has one.
 *
 * @param color
 * @returns {string}
 */
export function colorFromHex(color: string): string {
	if (color !== '' && color.substring(0, 1) === '#') {
		color = color.substring(1, 7)
	}

	return color
}
