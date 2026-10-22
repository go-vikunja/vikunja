import {
	SECONDS_A_DAY,
	SECONDS_A_HOUR,
	SECONDS_A_MINUTE,
	SECONDS_A_WEEK,
} from '@/constants/date'

export type PeriodUnit = 'seconds' | 'minutes' | 'hours' | 'days' | 'weeks' | 'months' | 'years'

/**
 * Convert time period given as seconds to days, hour, minutes, seconds
 */
export function secondsToPeriod(seconds: number): { unit: PeriodUnit, amount: number } {
	if (seconds % SECONDS_A_DAY === 0) {
		if (seconds % SECONDS_A_WEEK === 0) {
			return {unit: 'weeks', amount: seconds / SECONDS_A_WEEK}
		} else {
			return {unit: 'days', amount: seconds / SECONDS_A_DAY}
		}
	}

	if (seconds % SECONDS_A_HOUR === 0) {
		return {
			unit: 'hours',
			amount: seconds / SECONDS_A_HOUR,
		}
	}
	
	return {
		unit: 'minutes',
		amount: seconds / SECONDS_A_MINUTE,
	}
}

/**
 * Convert time period of days, hour, minutes, seconds to duration in seconds
 */
export function periodToSeconds(period: number, unit: PeriodUnit): number {
	switch (unit) {
		case 'minutes':
			return period * SECONDS_A_MINUTE
		case 'hours':
			return period * SECONDS_A_HOUR
		case 'days':
			return period * SECONDS_A_DAY
		case 'weeks':
			return period * SECONDS_A_WEEK
	}

	return 0
}
