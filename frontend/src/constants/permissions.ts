export const PERMISSIONS = {
	'READ': 0,
	'UPDATE': 1,
	'READ_WRITE': 2,
	'ADMIN': 3,
} as const

export type Permission = typeof PERMISSIONS[keyof typeof PERMISSIONS]
