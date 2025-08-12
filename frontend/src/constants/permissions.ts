export const PERMISSIONS = {
	'READ': 0,
	'READ_WRITE': 1,
	'ADMIN': 2,
} as const

export type Permission = typeof PERMISSIONS[keyof typeof PERMISSIONS]
