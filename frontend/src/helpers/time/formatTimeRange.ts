import {i18n} from '@/i18n'
import {formatDate} from '@/helpers/time/formatDate'
import {TIME_FORMAT, type TimeFormat} from '@/constants/timeFormat'

function formatDay(date: Date, now: Date): string {
	return new Intl.DateTimeFormat(i18n.global.locale.value, {
		day: 'numeric',
		month: 'short',
		year: date.getFullYear() === now.getFullYear() ? undefined : 'numeric',
	}).format(date)
}

function formatTime(date: Date, timeFormat: TimeFormat): string {
	return formatDate(date, timeFormat === TIME_FORMAT.HOURS_24 ? 'HH:mm' : 'hh:mm A')
}

export function formatDayTime(date: Date, timeFormat: TimeFormat, now = new Date()): string {
	return `${formatDay(date, now)} ${formatTime(date, timeFormat)}`
}

// A null end means the range is still open (e.g. a running timer).
export function formatTimeRange(start: Date, end: Date | null, timeFormat: TimeFormat, now = new Date()): string {
	const startDay = formatDay(start, now)
	const startLabel = formatDayTime(start, timeFormat, now)
	if (end === null) {
		return `${startLabel} – …`
	}
	const endDay = formatDay(end, now)
	const endTime = formatTime(end, timeFormat)
	return `${startLabel} – ${endDay === startDay ? endTime : `${endDay} ${endTime}`}`
}
