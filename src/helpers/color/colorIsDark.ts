export const colorIsDark = color => {
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

	// luma will be a value 0..255 where 0 indicates the darkest, and 255 the brightest
	const luma = 0.2126 * r + 0.7152 * g + 0.0722 * b // per ITU-R BT.709
	return luma > 128
}