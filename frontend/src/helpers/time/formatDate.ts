import {createDateFromString} from '@/helpers/time/createDateFromString'
import dayjs from 'dayjs'

import {i18n} from '@/i18n'
import {createSharedComposable} from '@vueuse/core'
import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useDateDisplay} from '@/composables/useDateDisplay'
import {DATE_DISPLAY, type DateDisplay} from '@/constants/dateDisplay'
import {DAYJS_LOCALE_MAPPING} from '@/i18n/useDayjsLanguageSync.ts'

export function dateIsValid(date: Date | null) {
	if (date === null) {
		return false
	}

	return date instanceof Date && !isNaN(date)
}

export const formatDate = (date: Date | string | null, f: string) => {
	if (!dateIsValid(date)) {
		return ''
	}

	date = createDateFromString(date)
	
	const locale = DAYJS_LOCALE_MAPPING[i18n.global.locale.value.toLowerCase()] ?? 'en'

	return date 
		? dayjs(date).locale(locale).format(f) 
		: ''
}

export function formatDateLong(date) {
	return formatDate(date, 'LLLL')
}

export function formatDateShort(date) {
	return formatDate(date, 'lll')
}

export const formatDateSince = (date: Date | string | null) => {
	if (!dateIsValid(date)) {
		return ''
	}

	date = createDateFromString(date)

	const locale = DAYJS_LOCALE_MAPPING[i18n.global.locale.value.toLowerCase()] ?? 'en'

	return date
		? dayjs(date).locale(locale).fromNow()
		: ''
}

export function formatISO(date) {
	return date ? new Date(date).toISOString() : ''
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

	return computed(() => (date: Date) => dateTimeFormatter.value.format(date))
}

export function formatDisplayDate(date: Date | string | null) {
	const {store} = useDateDisplay()
	const current = store.value

	return formatDisplayDateFormat(date, current)	
}

export function formatDisplayDateFormat(date: Date | string | null, format: DateDisplay) {
	if (typeof date === 'string') {
		date = createDateFromString(date)
	}
	
	if (date === null || !dateIsValid(date)) {
		return ''
	}

	switch (format) {
		case DATE_DISPLAY.MM_DD_YYYY:
			return formatDate(date, 'MM-DD-YYYY')
		case DATE_DISPLAY.DD_MM_YYYY:
			return formatDate(date, 'DD-MM-YYYY')
		case DATE_DISPLAY.YYYY_MM_DD:
			return formatDate(date, 'YYYY-MM-DD')
		case DATE_DISPLAY.MM_SLASH_DD_YYYY:
			return formatDate(date, 'MM/DD/YYYY')
		case DATE_DISPLAY.DD_SLASH_MM_YYYY:
			return formatDate(date, 'DD/MM/YYYY')
		case DATE_DISPLAY.YYYY_SLASH_MM_DD:
			return formatDate(date, 'YYYY/MM/DD')
		case DATE_DISPLAY.DAY_MONTH_YEAR: {
			return new Intl.DateTimeFormat(i18n.global.locale.value, {day: 'numeric', month: 'long', year: 'numeric'}).format(date)
		}
		case DATE_DISPLAY.WEEKDAY_DAY_MONTH_YEAR: {
			return new Intl.DateTimeFormat(i18n.global.locale.value, {weekday: 'long', day: 'numeric', month: 'long', year: 'numeric'}).format(date)
		}
		case DATE_DISPLAY.RELATIVE:
		default:
			return formatDateSince(date)
	}
}
