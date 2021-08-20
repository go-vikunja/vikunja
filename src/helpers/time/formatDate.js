import {createDateFromString} from '@/helpers/time/createDateFromString'
import {format, formatDistance} from 'date-fns'
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

	const currentDate = new Date()
	const distance = formatDistance(date, currentDate, {locale: locales[i18n.global.t('date.locale')]})

	if (date > currentDate) {
		return i18n.global.t('date.in', {date: distance})
	}

	return i18n.global.t('date.ago', {date: distance})
}
