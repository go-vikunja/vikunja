import type {DateISO} from '@/types/DateISO'
import type {DateKebab} from '@/types/DateKebab'

export function parseDateProp(kebabDate: DateKebab | undefined): string | undefined {
	try {

		if (!kebabDate) {
			throw new Error('No value')
		}
		const dateValues = kebabDate.split('-')
		const [, monthString, dateString] = dateValues
		const [year, month, date] = dateValues.map(val => Number(val))
		const dateValuesAreValid = (
			!Number.isNaN(year) &&
			monthString && monthString.length >= 1 && monthString.length <= 2 &&
			!Number.isNaN(month) &&
			month && month >= 1 && month <= 12 &&
			dateString && dateString.length >= 1 && dateString.length <= 31 &&
			!Number.isNaN(date) &&
			date && date >= 1 && date <= 31
		)
		if (!dateValuesAreValid || !month || !date || year === undefined) {
			throw new Error('Invalid date values')
		}
		return new Date(year, month - 1, date).toISOString() as DateISO
	} catch(_) {
		// ignore nonsense route queries
		return
	}
}
