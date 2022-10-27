import {MILLISECONDS_A_WEEK} from '@/constants/date'

export function getNextWeekDate(): Date {
	return new Date((new Date()).getTime() + MILLISECONDS_A_WEEK)
}
