import {createDateFromString} from '@/helpers/time/createDateFromString'
import {format, formatDistance} from 'date-fns'
import {enGB, de, fr, ru} from 'date-fns/locale'

const locales = {en: enGB, de, ch: de, fr, ru}

const dateIsValid = date => {
	if (date === null) {
		return false
	}

	return date instanceof Date && !isNaN(date)
}

export const formatDate = (date, f, locale) => {
	if (!dateIsValid(date)) {
		return ''
	}

	date = createDateFromString(date)

	return date ? format(date, f, {locale: locales[locale]}) : ''
}

export const formatDateSince = (date, $t) => {
	if (!dateIsValid(date)) {
		return ''
	}

	date = createDateFromString(date)

	const currentDate = new Date()
	const distance = formatDistance(date, currentDate, {locale: locales[$t('date.locale')]})

	if (date > currentDate) {
		return $t('date.in', {date: distance})
	}

	return $t('date.ago', {date: distance})
}
