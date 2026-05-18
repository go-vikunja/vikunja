import {repeatFromSettings, type RepeatFrequency} from '@/helpers/rrule'
import type {repeatParsedResult} from './types'

export const getRepeats = (text: string): repeatParsedResult => {
	const regex = /(^| )(((every|each) (([0-9]+|one|two|three|four|five|six|seven|eight|nine|ten) )?(hours?|days?|weeks?|months?|years?))|(annually|biannually|semiannually|biennially|daily|hourly|monthly|weekly|yearly))($| )/ig
	const results = regex.exec(text)
	if (results === null) {
		return {
			textWithoutMatched: text,
			repeat: null,
		}
	}

	let amount = 1
	switch (results[5] ? results[5].trim() : undefined) {
		case 'one':
			amount = 1
			break
		case 'two':
			amount = 2
			break
		case 'three':
			amount = 3
			break
		case 'four':
			amount = 4
			break
		case 'five':
			amount = 5
			break
		case 'six':
			amount = 6
			break
		case 'seven':
			amount = 7
			break
		case 'eight':
			amount = 8
			break
		case 'nine':
			amount = 9
			break
		case 'ten':
			amount = 10
			break
		default:
			amount = results[5] ? parseInt(results[5]) : 1
	}
	let freq: RepeatFrequency = 'hours'

	switch (results[2]) {
		case 'biennially':
			freq = 'years'
			amount = 2
			break
		case 'biannually':
		case 'semiannually':
			freq = 'months'
			amount = 6
			break
		case 'yearly':
		case 'annually':
			freq = 'years'
			break
		case 'daily':
			freq = 'days'
			break
		case 'hourly':
			freq = 'hours'
			break
		case 'monthly':
			freq = 'months'
			break
		case 'weekly':
			freq = 'weeks'
			break
		default:
			switch (results[7]) {
				case 'hour':
				case 'hours':
					freq = 'hours'
					break
				case 'day':
				case 'days':
					freq = 'days'
					break
				case 'week':
				case 'weeks':
					freq = 'weeks'
					break
				case 'month':
				case 'months':
					freq = 'months'
					break
				case 'year':
				case 'years':
					freq = 'years'
					break
			}
	}

	let matchedText = results[0]
	if(matchedText.endsWith(' ')) {
		matchedText = matchedText.substring(0, matchedText.length - 1)
	}

	return {
		textWithoutMatched: text.replace(matchedText, ''),
		repeat: repeatFromSettings(amount, freq),
	}
}
