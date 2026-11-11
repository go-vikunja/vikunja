import {REPEAT_TYPES, type IRepeatType} from '@/types/IRepeatAfter'
import type {repeatParsedResult} from './types'
import {numberWordsRegexGroup, parseAmount} from './numberWords'

export const getRepeats = (text: string): repeatParsedResult => {
	const regex = new RegExp(`(^| )(((every|each) (([0-9]+|${numberWordsRegexGroup}) )?(hours?|days?|weeks?|months?|years?))|(annually|biannually|semiannually|biennially|daily|hourly|monthly|weekly|yearly))($| )`, 'ig')
	const results = regex.exec(text)
	if (results === null) {
		return {
			textWithoutMatched: text,
			repeats: null,
		}
	}

	let amount = results[5] ? parseAmount(results[5].trim()) : 1
	let type: IRepeatType = REPEAT_TYPES.Hours

	switch (results[2]) {
		case 'biennially':
			type = REPEAT_TYPES.Years
			amount = 2
			break
		case 'biannually':
		case 'semiannually':
			type = REPEAT_TYPES.Months
			amount = 6
			break
		case 'yearly':
		case 'annually':
			type = REPEAT_TYPES.Years
			break
		case 'daily':
			type = REPEAT_TYPES.Days
			break
		case 'hourly':
			type = REPEAT_TYPES.Hours
			break
		case 'monthly':
			type = REPEAT_TYPES.Months
			break
		case 'weekly':
			type = REPEAT_TYPES.Weeks
			break
		default:
			switch (results[7]) {
				case 'hour':
				case 'hours':
					type = REPEAT_TYPES.Hours
					break
				case 'day':
				case 'days':
					type = REPEAT_TYPES.Days
					break
				case 'week':
				case 'weeks':
					type = REPEAT_TYPES.Weeks
					break
				case 'month':
				case 'months':
					type = REPEAT_TYPES.Months
					break
				case 'year':
				case 'years':
					type = REPEAT_TYPES.Years
					break
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
	}
}
