import {preprocess, date} from 'zod'

export const DateSchema = preprocess((arg) => {
	if (
		// FIXME: Add comment why we check for `0001`
		typeof arg == 'string' && !arg.startsWith('0001') ||
		arg instanceof Date
	) {
		return new Date(arg)
	}
}, date())