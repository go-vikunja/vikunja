export const REPEAT_TYPES = {
	Seconds: 'seconds',
	Minutes: 'minutes',
	Hours: 'hours',
	Days: 'days',
	Weeks: 'weeks',
	Months: 'months',
	Years: 'years',
} as const

export type IRepeatType = typeof REPEAT_TYPES[keyof typeof REPEAT_TYPES]

export interface IRepeatAfter {
	type: IRepeatType,
	amount: number,
}
