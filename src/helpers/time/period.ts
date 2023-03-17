import {SECONDS_A_DAY, SECONDS_A_HOUR, SECONDS_A_MINUTE} from '@/constants/date'

/**
 * Convert time period given as seconds to days, hour, minutes, seconds
 */
export function secondsToPeriod(seconds: number): {days: number, hours: number, minutes: number, seconds: number} {
	const d = Math.floor(seconds / SECONDS_A_DAY)
	const h = Math.floor(seconds % SECONDS_A_DAY / 3600)
	const m = Math.floor(seconds % SECONDS_A_HOUR / 60)
	const s = Math.floor(seconds % 60)
	return {days: d, hours: h, minutes: m, seconds: s}
}

/**
 * Convert time period of days, hour, minutes, seconds to duration in seconds
 */
export function periodToSeconds(days: number, hours: number, minutes: number, seconds: number): number {
	return days * SECONDS_A_DAY + hours * SECONDS_A_HOUR + minutes * SECONDS_A_MINUTE + seconds
}
