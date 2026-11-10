
// Okabe-Ito based: stays distinguishable for red-green colorblind users.
const COLORS = [
	'#E69F00',
	'#D55E00',
	'#F0E442',
	'#009E73',
	'#56B4E9',
	'#0072B2',
	'#CC79A7',
	'#7F23FF',
	'#373f47',
]

export function getRandomColorHex(): string {
	return COLORS[Math.floor(Math.random() * COLORS.length)]
}
