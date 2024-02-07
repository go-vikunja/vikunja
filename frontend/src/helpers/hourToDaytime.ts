import type { Daytime } from '@/composables/useDaytimeSalutation'

export function hourToDaytime(now: Date): Daytime {
	const hours = now.getHours()

	const daytimeMap = {
		night: hours < 5 || hours > 23,
		morning: hours < 11,
		day: hours < 18,
		evening: hours < 23,
	} as Record<Daytime, boolean>

	return (Object.keys(daytimeMap) as Daytime[]).find((daytime) => daytimeMap[daytime]) || 'night'
}
