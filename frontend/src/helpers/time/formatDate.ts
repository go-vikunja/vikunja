import {createDateFromString} from '@/helpers/time/createDateFromString'
import dayjs from 'dayjs'

import {i18n} from '@/i18n'
import {createSharedComposable} from '@vueuse/core'
import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {DAYJS_LOCALE_MAPPING} from '@/i18n/useDayjsLanguageSync.ts'

export function dateIsValid(date: Date | null) {
	if (date === null) {
		return false
	}

	return date instanceof Date && !isNaN(date.getTime())
}

export const formatDate = (date: Date | string | null, f: string) => {
	if (!dateIsValid(date as Date)) {
		return ''
	}

	date = createDateFromString(date!)
	
	// @ts-expect-error - Vue i18n locale type issue
	const locale = (DAYJS_LOCALE_MAPPING as Record<string, string>)[i18n.global.locale.value.toLowerCase()] ?? 'en'

	return date 
		? dayjs(date).locale(locale).format(f) 
		: ''
}

export function formatDateLong(date: Date | string | null) {
	return formatDate(date, 'LLLL')
}

export function formatDateShort(date: Date | string | null) {
	return formatDate(date, 'lll')
}

export const formatDateSince = (date: Date | string | null) => {
	if (!dateIsValid(date as Date)) {
		return ''
	}

	date = createDateFromString(date!)

	// @ts-expect-error - Vue i18n locale type issue
	const locale = (DAYJS_LOCALE_MAPPING as Record<string, string>)[i18n.global.locale.value.toLowerCase()] ?? 'en'

	return date
		? dayjs(date).locale(locale).fromNow()
		: ''
}

export function formatISO(date: Date | string | null) {
	return date ? new Date(date).toISOString() : ''
}

/**
 * Because `Intl.DateTimeFormat` is expensive to instatiate we try to reuse it as often as possible,
 * by creating a shared composable.
 */
export const useDateTimeFormatter = createSharedComposable((options?: MaybeRefOrGetter<Intl.DateTimeFormatOptions>) => {
	// @ts-expect-error - Vue i18n locale type issue
	return computed(() => new Intl.DateTimeFormat(i18n.global.locale.value, toValue(options)))
})

export function useWeekDayFromDate() {
	const dateTimeFormatter = useDateTimeFormatter({weekday: 'short'})

	return computed(() => (date: Date) => dateTimeFormatter.value.format(date))
}
