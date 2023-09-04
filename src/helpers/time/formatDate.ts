import {createDateFromString} from '@/helpers/time/createDateFromString'
import {format, formatDistanceToNow} from 'date-fns'

// FIXME: support all locales and load dynamically
import {enGB, de, fr, ru} from 'date-fns/locale'

import {i18n} from '@/i18n'
import {createSharedComposable, type MaybeRef} from '@vueuse/core'
import {computed, unref} from 'vue'

const locales = {en: enGB, de, ch: de, fr, ru}

export function dateIsValid(date: Date | null) {
	if (date === null) {
		return false
	}

	return date instanceof Date && !isNaN(date)
}

export const formatDate = (date, f, locale = i18n.global.t('date.locale')) => {
	if (!dateIsValid(date)) {
		return ''
	}

	date = createDateFromString(date)

	return date ? format(date, f, {locale: locales[locale]}) : ''
}

export function formatDateLong(date) {
	return formatDate(date, 'PPPPpppp')
}

export function formatDateShort(date) {
	return formatDate(date, 'PPpp')
}

export const formatDateSince = (date) => {
	if (!dateIsValid(date)) {
		return ''
	}

	date = createDateFromString(date)

	return formatDistanceToNow(date, {
		locale: locales[i18n.global.t('date.locale')],
		addSuffix: true,
	})
}

export function formatISO(date) {
	return date ? new Date(date).toISOString() : ''
}

/**
 * Because `Intl.DateTimeFormat` is expensive to instatiate we try to reuse it as often as possible,
 * by creating a shared composable.
 */
export const useDateTimeFormatter = createSharedComposable((options?: MaybeRef<Intl.DateTimeFormatOptions>) => {
	return computed(() => new Intl.DateTimeFormat(i18n.global.locale.value, unref(options)))
})

export function useWeekDayFromDate() {
	const dateTimeFormatter = useDateTimeFormatter({weekday: 'short'})

	return computed(() => (date: Date) => dateTimeFormatter.value.format(date))
}