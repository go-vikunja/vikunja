export const RIGHTS = {
	'READ': 0,
	'READ_WRITE': 1,
	'ADMIN': 2,
} as const

export type Right = typeof RIGHTS[keyof typeof RIGHTS]
