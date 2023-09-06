
const COLORS = [
	'#ffbe0b',
	'#fd8a09',
	'#fb5607',
	'#ff006e',
	'#efbdeb',
	'#8338ec',
	'#5f5ff6',
	'#3a86ff',
	'#4c91ff',
	'#0ead69',
	'#25be8b',
	'#073b4c',
	'#373f47',
]

export function getRandomColorHex(): string {
	return COLORS[Math.floor(Math.random() * COLORS.length)]
}
