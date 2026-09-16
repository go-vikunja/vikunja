import {createDateFromString} from '@/helpers/time/createDateFromString'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {toISOStringOrNull} from '@/helpers/time/toISOStringOrNull'
import dayjs from 'dayjs'

import {i18n} from '@/i18n'
import {createSharedComposable} from '@vueuse/core'
import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useDateDisplay} from '@/composables/useDateDisplay'
import {useGlobalNow} from '@/composables/useGlobalNow'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {DATE_DISPLAY, type DateDisplay} from '@/constants/dateDisplay'
import {TIME_FORMAT, type TimeFormat} from '@/constants/timeFormat'
import {DAYJS_LOCALE_MAPPING} from '@/i18n/useDayjsLanguageSync.ts'

export function dateIsValid(date: Date | string | null | undefined): boolean {
	return toDate(date) !== null
}

function toDate(date: Date | string | null | undefined): Date | null {
	if (date === null || typeof date === 'undefined') {
		return null
	}

	return parseDateOrNull(createDateFromString(date))
}

export const formatDate = (date: Date | string | null | undefined, f: string) => {
	const parsed = toDate(date)
	if (parsed === null) {
		return ''
	}

	const locale = DAYJS_LOCALE_MAPPING[i18n.global.locale.value.toLowerCase()] ?? 'en'

	return dayjs(parsed).locale(locale).format(f)
}

export function formatDateLong(date) {
	return formatDate(date, 'LLLL')
}

export function formatDateShort(date) {
	return formatDate(date, 'lll')
}

export const formatDateSince = (date: Date | string | null | undefined) => {
	const parsed = toDate(date)
	if (parsed === null) {
		return ''
	}

	const locale = DAYJS_LOCALE_MAPPING[i18n.global.locale.value.toLowerCase()] ?? 'en'

	// Computing the relative string against the shared, ticking `now` (instead of fromNow's
	// internal Date.now()) makes every reactive caller re-render on the 60s tick, so open views
	// don't keep showing a stale "x minutes ago".
	const {now} = useGlobalNow()

	return dayjs(parsed).locale(locale).from(now.value)
}

export function formatISO(date: Date | string | null | undefined) {
	return toISOStringOrNull(date) ?? ''
}

/**
 * Because `Intl.DateTimeFormat` is expensive to instatiate we try to reuse it as often as possible,
 * by creating a shared composable.
 */
export const useDateTimeFormatter = createSharedComposable((options?: MaybeRefOrGetter<Intl.DateTimeFormatOptions>) => {
	return computed(() => new Intl.DateTimeFormat(i18n.global.locale.value, toValue(options)))
})

export function useWeekDayFromDate() {
	const dateTimeFormatter = useDateTimeFormatter({weekday: 'short'})

	return computed(() => (date: Date) => dateIsValid(date) ? dateTimeFormatter.value.format(date) : '')
}

export function formatDisplayDate(date: Date | string | null | undefined) {
	const {store: dateDisplay} = useDateDisplay()
	const {store: timeFormat} = useTimeFormat()

	return formatDisplayDateFormat(date, dateDisplay.value, timeFormat.value)	
}

export function formatDisplayDateFormat(date: Date | string | null | undefined, format: DateDisplay, timeFormat?: TimeFormat) {
	const parsed = toDate(date)
	if (parsed === null) {
		return ''
	}

	// Determine the time format string to use
	// For 24-hour: HH:mm (24-hour format)
	// For 12-hour: hh:mm A (explicit 12-hour format with AM/PM, ignoring locale default)
	const timeFormatString = timeFormat === TIME_FORMAT.HOURS_24 ? 'HH:mm' : 'hh:mm A'

	switch (format) {
		case DATE_DISPLAY.MM_DD_YYYY:
			return formatDate(parsed, `MM-DD-YYYY ${timeFormatString}`)
		case DATE_DISPLAY.DD_MM_YYYY:
			return formatDate(parsed, `DD-MM-YYYY ${timeFormatString}`)
		case DATE_DISPLAY.YYYY_MM_DD:
			return formatDate(parsed, `YYYY-MM-DD ${timeFormatString}`)
		case DATE_DISPLAY.MM_SLASH_DD_YYYY:
			return formatDate(parsed, `MM/DD/YYYY ${timeFormatString}`)
		case DATE_DISPLAY.DD_SLASH_MM_YYYY:
			return formatDate(parsed, `DD/MM/YYYY ${timeFormatString}`)
		case DATE_DISPLAY.YYYY_SLASH_MM_DD:
			return formatDate(parsed, `YYYY/MM/DD ${timeFormatString}`)
		case DATE_DISPLAY.DAY_MONTH_YEAR: {
			const hour12 = timeFormat !== TIME_FORMAT.HOURS_24
			return new Intl.DateTimeFormat(i18n.global.locale.value, {day: 'numeric', month: 'long', year: 'numeric', hour: 'numeric', minute: 'numeric', hour12}).format(parsed)
		}
		case DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR: {
			const hour12 = timeFormat !== TIME_FORMAT.HOURS_24
			return new Intl.DateTimeFormat(i18n.global.locale.value, {weekday: 'long', day: 'numeric', month: 'long', year: 'numeric', hour: 'numeric', minute: 'numeric', hour12}).format(parsed)
		}
		case DATE_DISPLAY.RELATIVE:
		default:
			return formatDateSince(parsed)
	}
}
