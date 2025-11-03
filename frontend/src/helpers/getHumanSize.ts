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
	while (size > 1024) {
		size /= 1024
		iterator++
	}

	return Number(Math.round(Number(size + 'e2')) + 'e-2') + ' ' + SIZES[iterator]
}
