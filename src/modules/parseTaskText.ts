import {parseDate} from '../helpers/time/parseDate'
import _priorities from '../models/constants/priorities.json'

const VIKUNJA_PREFIXES: Prefixes = {
	label: '*',
	list: '+',
	priority: '!',
	assignee: '@',
}

const TODOIST_PREFIXES: Prefixes = {
	label: '@',
	list: '#',
	priority: '!',
	assignee: '+',
}

export enum PrefixMode {
	Disabled = 'disabled',
	Default = 'vikunja',
	Todoist = 'todoist',
}

export const PREFIXES = {
	[PrefixMode.Disabled]: undefined,
	[PrefixMode.Default]: VIKUNJA_PREFIXES,
	[PrefixMode.Todoist]: TODOIST_PREFIXES,
}

const priorities: Priorites = _priorities

interface Priorites {
	UNSET: number,
	LOW: number,
	MEDIUM: number,
	HIGH: number,
	URGENT: number,
	DO_NOW: number,
}

enum RepeatType {
	Hours = 'hours',
	Days = 'days',
	Weeks = 'weeks',
	Months = 'months',
	Years = 'years',
}

interface Repeats {
	type: RepeatType,
	amount: number,
}

interface repeatParsedResult {
	textWithoutMatched: string,
	repeats: Repeats | null,
}

interface ParsedTaskText {
	text: string,
	date: Date | null,
	labels: string[],
	list: string | null,
	priority: number | null,
	assignees: string[],
	repeats: Repeats | null,
}

interface Prefixes {
	label: string,
	list: string,
	priority: string,
	assignee: string,
}

/**
 * Parses task text for dates, assignees, labels, lists, priorities and returns an object with all found intents.
 *
 * @param text
 */
export const parseTaskText = (text: string, prefixesMode: PrefixMode = PrefixMode.Default): ParsedTaskText => {
	const result: ParsedTaskText = {
		text: text,
		date: null,
		labels: [],
		list: null,
		priority: null,
		assignees: [],
		repeats: null,
	}

	const prefixes = PREFIXES[prefixesMode]
	if (prefixes === undefined) {
		return result
	}

	result.labels = getItemsFromPrefix(text, prefixes.label)

	const lists: string[] = getItemsFromPrefix(text, prefixes.list)
	result.list = lists.length > 0 ? lists[0] : null

	result.priority = getPriority(text, prefixes.priority)

	result.assignees = getItemsFromPrefix(text, prefixes.assignee)

	const {textWithoutMatched, repeats} = getRepeats(text)
	result.text = textWithoutMatched
	result.repeats = repeats

	const {newText, date} = parseDate(result.text)
	result.text = newText
	result.date = date

	return cleanupResult(result, prefixes)
}

const getItemsFromPrefix = (text: string, prefix: string): string[] => {
	const items: string[] = []

	const itemParts = text.split(' ' + prefix)
	if (text.startsWith(prefix)) {
		const firstItem = text.split(prefix)[1]
		itemParts.unshift(firstItem)
	}

	itemParts.forEach((p, index) => {
		// First part contains the rest
		if (index < 1) {
			return
		}

		p = p.replace(prefix, '')

		let itemText
		if (p.charAt(0) === '\'') {
			itemText = p.split('\'')[1]
		} else if (p.charAt(0) === '"') {
			itemText = p.split('"')[1]
		} else {
			// Only until the next space
			itemText = p.split(' ')[0]
		}
		items.push(itemText)
	})

	return Array.from(new Set(items))
}

const getPriority = (text: string, prefix: string): number | null => {
	const ps = getItemsFromPrefix(text, prefix)
	if (ps.length === 0) {
		return null
	}

	for (const p of ps) {
		for (const pi of Object.values(priorities)) {
			if (pi === parseInt(p)) {
				return parseInt(p)
			}
		}
	}

	return null
}

const getRepeats = (text: string): repeatParsedResult => {
	const regex = /((every|each) (([0-9]+|one|two|three|four|five|six|seven|eight|nine|ten) )?(hours?|days?|weeks?|months?|years?))|anually|bianually|semiannually|biennially|daily|hourly|monthly|weekly|yearly/ig
	const results = regex.exec(text)
	if (results === null) {
		return {
			textWithoutMatched: text,
			repeats: null,
		}
	}

	let amount = 1
	switch (results[3] ? results[3].trim() : undefined) {
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
			amount = results[3] ? parseInt(results[3]) : 1
	}
	let type: RepeatType = RepeatType.Hours

	switch (results[0]) {
		case 'biennially':
			type = RepeatType.Years
			amount = 2
			break
		case 'bianually':
		case 'semiannually':
			type = RepeatType.Months
			amount = 6
			break
		case 'yearly':
		case 'anually':
			type = RepeatType.Years
			break
		case 'daily':
			type = RepeatType.Days
			break
		case 'hourly':
			type = RepeatType.Hours
			break
		case 'monthly':
			type = RepeatType.Months
			break
		case 'weekly':
			type = RepeatType.Weeks
			break
		default:
			switch (results[5]) {
				case 'hour':
				case 'hours':
					type = RepeatType.Hours
					break
				case 'day':
				case 'days':
					type = RepeatType.Days
					break
				case 'week':
				case 'weeks':
					type = RepeatType.Weeks
					break
				case 'month':
				case 'months':
					type = RepeatType.Months
					break
				case 'year':
				case 'years':
					type = RepeatType.Years
					break
			}
	}

	return {
		textWithoutMatched: text.replace(results[0], ''),
		repeats: {
			amount,
			type,
		},
	}
}

const cleanupItemText = (text: string, items: string[], prefix: string): string => {
	items.forEach(l => {
		text = text
			.replace(`${prefix}'${l}' `, '')
			.replace(`${prefix}'${l}'`, '')
			.replace(`${prefix}"${l}" `, '')
			.replace(`${prefix}"${l}"`, '')
			.replace(`${prefix}${l} `, '')
			.replace(`${prefix}${l}`, '')
	})
	return text
}

const cleanupResult = (result: ParsedTaskText, prefixes: Prefixes): ParsedTaskText => {
	result.text = cleanupItemText(result.text, result.labels, prefixes.label)
	result.text = result.list !== null ? cleanupItemText(result.text, [result.list], prefixes.list) : result.text
	result.text = result.priority !== null ? cleanupItemText(result.text, [String(result.priority)], prefixes.priority) : result.text
	result.text = cleanupItemText(result.text, result.assignees, prefixes.assignee)
	result.text = result.text.trim()

	return result
}
