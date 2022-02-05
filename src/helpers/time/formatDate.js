import {createDateFromString} from '@/helpers/time/createDateFromString'
import {format, formatDistanceToNow, formatISO as formatISOfns} from 'date-fns'
import {enGB, de, fr, ru} from 'date-fns/locale'

import {i18n} from '@/i18n'

const locales = {en: enGB, de, ch: de, fr, ru}

const dateIsValid = date => {
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
	return date ? formatISOfns(date) : ''
}
