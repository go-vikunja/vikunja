export const colorIsDark = color => {
	if (color === '#' || color === '') {
		return true // Defaults to dark
	}

	if (color.substring(0, 1) !== '#') {
		color = '#' + color
	}

	let rgb = parseInt(color.substring(1, 7), 16)   // convert rrggbb to decimal
	let r = (rgb >> 16) & 0xff  // extract red
	let g = (rgb >> 8) & 0xff  // extract green
	let b = (rgb >> 0) & 0xff  // extract blue

	// luma will be a value 0..255 where 0 indicates the darkest, and 255 the brightest
	let luma = 0.2126 * r + 0.7152 * g + 0.0722 * b // per ITU-R BT.709
	return luma > 128
}