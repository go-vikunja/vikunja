import {REPEAT_TYPES, type IRepeatType} from '@/types/IRepeatAfter'
import {replaceAll} from '@/helpers/replaceAll'
import {getDefaultLocales, type QuickAddMagicLocale} from './locales'
import {findByForm, toPattern} from './locales/pattern'
import type {repeatParsedResult} from './types'

export const getRepeats = (text: string, locales: QuickAddMagicLocale[] = getDefaultLocales()): repeatParsedResult => {
	for (const locale of locales) {
		const result = matchStandardRepeat(text, locale)
		if (result !== null) {
			return result
		}
	}

	for (const locale of locales) {
		const result = matchWeekdayRepeat(text, locale)
		if (result !== null) {
			return result
		}
	}

	return {
		textWithoutMatched: text,
		repeats: null,
		matched: null,
	}
}

const matchStandardRepeat = (text: string, locale: QuickAddMagicLocale): repeatParsedResult | null => {
	const every = locale.everyKeywords.map(toPattern).join('|')
	const numbers = ['[0-9]+', ...locale.numberWords.flatMap(n => n.forms)].map(toPattern).join('|')
	const units = locale.repeatUnits.flatMap(u => u.forms).map(toPattern).join('|')
	const adverbs = locale.repeatAdverbs.flatMap(a => a.forms).map(toPattern).join('|')
	const regex = new RegExp(`(^| )(((${every}) ((${numbers}) )?(${units}))|(${adverbs}))($| )`, 'ig')
	const results = regex.exec(text)
	if (results === null) {
		return null
	}

	let amount = 1
	if (results[5] !== undefined) {
		const numberWord = results[5].trim()
		amount = findByForm(locale.numberWords, numberWord)?.value ?? parseInt(numberWord)
	}

	let type: IRepeatType = REPEAT_TYPES.Hours
	if (results[8] !== undefined) {
		const adverb = findByForm(locale.repeatAdverbs, results[8])
		if (adverb !== undefined) {
			type = adverb.type
			amount = adverb.amount
		}
	} else if (results[7] !== undefined) {
		const unit = findByForm(locale.repeatUnits, results[7])
		if (unit !== undefined) {
			type = unit.type
		}
	}

	let matchedText = results[0]
	if(matchedText.endsWith(' ')) {
		matchedText = matchedText.substring(0, matchedText.length - 1)
	}

	return {
		textWithoutMatched: text.replace(matchedText, ''),
		repeats: {
			amount,
			type,
		},
		matched: matchedText.trim(),
	}
}

/**
 * "every thursday" / "кожен четвер" — a weekly repeat anchored on the weekday.
 * Only the "every" keyword is stripped; the weekday stays in the text so the
 * date parser can anchor the due date on it.
 */
const matchWeekdayRepeat = (text: string, locale: QuickAddMagicLocale): repeatParsedResult | null => {
	const every = locale.everyKeywords.map(toPattern).join('|')
	const weekdays = locale.weekdays.flatMap(w => w.forms).map(toPattern).join('|')
	const regex = new RegExp(`(^| )(${every}) (${weekdays})($| )`, 'i')
	const results = regex.exec(text)
	if (results === null) {
		return null
	}

	return {
		textWithoutMatched: replaceAll(text, results[2], ''),
		repeats: {
			amount: 1,
			type: REPEAT_TYPES.Weeks,
		},
		matched: results[2],
	}
}
