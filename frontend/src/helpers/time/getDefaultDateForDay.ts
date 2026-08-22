import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {useAuthStore} from '@/stores/auth'

function parseUserDefaultDueTime(defaultDueTime?: string): {hours: number, minutes: number} | null {
	if (!defaultDueTime) {
		return null
	}

	const match = /^(\d{2}):(\d{2})$/.exec(defaultDueTime)
	if (!match) {
		return null
	}

	const hours = Number(match[1])
	const minutes = Number(match[2])
	if (hours > 23 || minutes > 59) {
		return null
	}

	return {hours, minutes}
}

export function getDefaultDueTimeParts(date: Date): {hours: number, minutes: number} {
	const defaultDueTime = useAuthStore().settings.frontendSettings.defaultDueTime
	const parsedTime = parseUserDefaultDueTime(defaultDueTime)

	if (parsedTime !== null) {
		return parsedTime
	}

	return {
		hours: calculateNearestHours(date),
		minutes: 0,
	}
}

export function getDefaultDateForDay(date: Date): Date {
	const newDate = new Date(date)
	const defaultDueTime = getDefaultDueTimeParts(newDate)
	newDate.setHours(defaultDueTime.hours, defaultDueTime.minutes, 0, 0)
	return newDate
}

export {parseUserDefaultDueTime}
