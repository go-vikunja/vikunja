export function colorIsDark(color: string | undefined) {
	if (typeof color === 'undefined') {
		return true // Defaults to dark
	}

	if (color === '#' || color === '') {
		return true // Defaults to dark
	}

	if (color.substring(0, 1) !== '#') {
		color = '#' + color
	}

	const rgb = parseInt(color.substring(1, 7), 16)   // convert rrggbb to decimal
	const r = (rgb >> 16) & 0xff  // extract red
	const g = (rgb >> 8) & 0xff  // extract green
	const b = (rgb >> 0) & 0xff  // extract blue

	const toLinear = (c: number) => {
		const v = c / 255
		return v <= 0.04045 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4)
	}
	const luminance = 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b)

	// sqrt(1.05 * 0.05) - 0.05: the luminance where contrast against #000 equals contrast against #fff,
	// guaranteeing >= 4.58:1 for whichever of black/white gets picked
	return luminance > 0.1791
}
