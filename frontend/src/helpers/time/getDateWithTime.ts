import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {useAuthStore} from '@/stores/auth'

export function parseUserDefaultTime(default_due_time?: string): {hours: number, minutes: number} | null {
	if (!default_due_time) {
		return null
	}

	const match = /^(\d{2}):(\d{2})$/.exec(default_due_time)
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

export function getDefaultTimeParts(date: Date): {hours: number, minutes: number} {
	const default_due_time = useAuthStore().settings.frontend_settings.default_due_time
	const parsedTime = parseUserDefaultTime(default_due_time)

	if (parsedTime !== null) {
		return parsedTime
	}

	return {
		hours: calculateNearestHours(date),
		minutes: 0,
	}
}

export function getDateWithTime(date: Date): Date {
	const newDate = new Date(date)
	const default_due_time = getDefaultTimeParts(newDate)
	newDate.setHours(default_due_time.hours, default_due_time.minutes, 0, 0)
	return newDate
}
