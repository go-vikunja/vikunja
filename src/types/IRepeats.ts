export const REPEAT_TYPES = {
	Hours: 'hours',
	Days: 'days',
	Weeks: 'weeks',
	Months: 'months',
	Years: 'years',
} as const

export type RepeatType = typeof REPEAT_TYPES[keyof typeof REPEAT_TYPES]

export interface IRepeats {
	type: RepeatType,
	amount: number,
}