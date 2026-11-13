const SIZES = [
	'B',
	'KB',
	'MB',
	'GB',
	'TB',
] as const

export function getHumanSize(inputSize: number) {
	let iterator = 0
	let size = inputSize
	// 1024 of a unit is one of the next ("1 KB", not "1024 B"). Stop at the largest unit.
	while (size >= 1024 && iterator < SIZES.length - 1) {
		size /= 1024
		iterator++
	}

	return Number(Math.round(Number(size + 'e2')) + 'e-2') + ' ' + SIZES[iterator]
}
