export const DATE_DISPLAY = {
	RELATIVE: 'relative',
	MM_DD_YYYY: 'mm-dd-yyyy',
	DD_MM_YYYY: 'dd-mm-yyyy',
	YYYY_MM_DD: 'yyyy-mm-dd',
	MM_SLASH_DD_YYYY: 'mm/dd/yyyy',
	DD_SLASH_MM_YYYY: 'dd/mm/yyyy',
	YYYY_SLASH_MM_DD: 'yyyy/mm/dd',
	DAY_MONTH_YEAR: 'dayMonthYear',
	WEEKDAY_DAY_MONTH_YEAR: 'weekdayDayMonthYear',
} as const

export type DateDisplay = typeof DATE_DISPLAY[keyof typeof DATE_DISPLAY]
