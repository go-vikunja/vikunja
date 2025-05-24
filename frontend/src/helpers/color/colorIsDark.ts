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

	// this is a quick and dirty implementation of the WCAG 3.0 APCA color contrast formula
	// see: https://gist.github.com/Myndex/e1025706436736166561d339fd667493#andys-shortcut-to-luminance--lightness
	const Ys = Math.pow(r/255.0,2.2) * 0.2126 +
		Math.pow(g/255.0,2.2) * 0.7152 +
		Math.pow(b/255.0,2.2) * 0.0722

	return Math.pow(Ys,0.678) >= 0.5
}
