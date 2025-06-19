import type {DateISO} from '@/types/DateISO'
import type {DateKebab} from '@/types/DateKebab'

// ✅ Format a date to YYYY-MM-DD (or any other format)
function padTo2Digits(num: number) {
	return num.toString().padStart(2, '0')
}

export function isoToKebabDate(isoDate: DateISO) {
	const date = new Date(isoDate)
	return [
		date.getFullYear(),
		padTo2Digits(date.getMonth() + 1), // January is 0, but we want it to be 1
		padTo2Digits(date.getDate()),
	].join('-') as DateKebab
}
