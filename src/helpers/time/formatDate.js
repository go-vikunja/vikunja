import {createDateFromString} from '@/helpers/time/createDateFromString'
import {format, formatDistanceToNow} from 'date-fns'
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

	return formatDistanceToNow(date, {
		locale: locales[$t('date.locale')],
		addSuffix: true,
	})
}
