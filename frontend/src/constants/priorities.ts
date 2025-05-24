export const PRIORITIES = {
	'UNSET': 0,
	'LOW': 1,
	'MEDIUM': 2,
	'HIGH': 3,
	'URGENT': 4,
	'DO_NOW': 5,
} as const

export type Priority = typeof PRIORITIES[keyof typeof PRIORITIES]
