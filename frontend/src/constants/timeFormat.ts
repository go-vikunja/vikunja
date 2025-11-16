export const TIME_FORMAT = {
	HOURS_12: '12h',
	HOURS_24: '24h',
} as const

export type TimeFormat = typeof TIME_FORMAT[keyof typeof TIME_FORMAT]
