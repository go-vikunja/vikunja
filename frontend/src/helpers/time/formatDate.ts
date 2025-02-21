import {createDateFromString} from '@/helpers/time/createDateFromString'
import dayjs from 'dayjs'

import {i18n} from '@/i18n'
import {createSharedComposable} from '@vueuse/core'
import {computed, toValue, type MaybeRefOrGetter} from 'vue'

export function dateIsValid(date: Date | null) {
	if (date === null) {
		return false
	}

	return date instanceof Date && !isNaN(date)
}

export const formatDate = (date: Date | string | null, f: string, locale = i18n.global.t('date.locale')) => {
	if (!dateIsValid(date)) {
		return ''
	}

	date = createDateFromString(date)

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

export const formatDateSince = (date: Date | string | null, locale = i18n.global.t('date.locale')) => {
	if (!dateIsValid(date)) {
		return ''
	}

	date = createDateFromString(date)

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