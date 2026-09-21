export const AUTH_TYPES = {
	UNKNOWN: 0,
	USER: 1,
	LINK_SHARE: 2,
} as const

export type AuthType = typeof AUTH_TYPES[keyof typeof AUTH_TYPES]

export const ERROR_CODE_TOTP_REQUIRED = 1017
export const ERROR_CODE_USER_DOES_NOT_EXIST = 1005
export const ERROR_CODE_ACCOUNT_DISABLED = 1020
export const ERROR_CODE_ACCOUNT_LOCKED = 1040
