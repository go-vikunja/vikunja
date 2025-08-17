import {parseDate} from '../helpers/time/parseDate'
import {PRIORITIES} from '@/constants/priorities'
import {REPEAT_TYPES, type IRepeatAfter, type IRepeatType} from '@/types/IRepeatAfter'

const VIKUNJA_PREFIXES: Prefixes = {
	label: '*',
	project: '+',
	priority: '!',
	assignee: '@',
}

const TODOIST_PREFIXES: Prefixes = {
	label: '@',
	project: '#',
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

interface repeatParsedResult {
	textWithoutMatched: string,
	repeats: IRepeatAfter | null,
}

export interface ParsedTaskText {
	text: string,
	date: Date | null,
	labels: string[],
	project: string | null,
	priority: number | null,
	assignees: string[],
	repeats: IRepeatAfter | null,
}

interface Prefixes {
	label: string,
	project: string,
	priority: string,
	assignee: string,
}

/**
 * Parses task text for dates, assignees, labels, projects, priorities and returns an object with all found intents.
 *
 * @param text
 */
export const parseTaskText = (text: string, prefixesMode: PrefixMode = PrefixMode.Default, now: Date = new Date()): ParsedTaskText => {
	const result: ParsedTaskText = {
		text: text,
		date: null,
		labels: [],
		project: null,
		priority: null,
		assignees: [],
		repeats: null,
	}

	const prefixes = PREFIXES[prefixesMode]
	if (prefixes === undefined) {
		return result
	}

	result.labels = getLabelsFromPrefix(text, prefixesMode) ?? []
	result.text = cleanupItemText(result.text, result.labels, prefixes.label)

	result.project = getProjectFromPrefix(result.text, prefixesMode)
	result.text = result.project !== null ? cleanupItemText(result.text, [result.project], prefixes.project) : result.text

	result.priority = getPriority(result.text, prefixes.priority)
	result.text = result.priority !== null ? cleanupItemText(result.text, [String(result.priority)], prefixes.priority) : result.text

	result.assignees = getItemsFromPrefix(result.text, prefixes.assignee)

	const {textWithoutMatched, repeats} = getRepeats(result.text)
	result.text = textWithoutMatched
	result.repeats = repeats

	const {newText, date} = parseDate(result.text, now)
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

		if (p.startsWith(prefix)) {
			p = p.substring(1)
		}

		let itemText
		if (p.charAt(0) === '\'') {
			itemText = p.split('\'')[1]
		} else if (p.charAt(0) === '"') {
			itemText = p.split('"')[1]
		} else {
			// Only until the next space
			itemText = p.split(' ')[0]
		}

		if (itemText !== '') {
			items.push(itemText)
		}
	})

	return Array.from(new Set(items))
}

export const getProjectFromPrefix = (text: string, prefixMode: PrefixMode): string | null => {
	const projectPrefix = PREFIXES[prefixMode]?.project
	if(typeof projectPrefix === 'undefined') {
		return null
	}
	const projects: string[] = getItemsFromPrefix(text, projectPrefix)
	return projects.length > 0 ? projects[0] : null
}

export const getLabelsFromPrefix = (text: string, prefixMode: PrefixMode): string[] | null => {
	const labelsPrefix = PREFIXES[prefixMode]?.label
	if(typeof labelsPrefix === 'undefined') {
		return null
	}
	return getItemsFromPrefix(text, labelsPrefix)
}

const getPriority = (text: string, prefix: string): number | null => {
	const ps = getItemsFromPrefix(text, prefix)
	if (ps.length === 0) {
		return null
	}

	for (const p of ps) {
		for (const pi of Object.values(PRIORITIES)) {
			if (pi === parseInt(p)) {
				return parseInt(p)
			}
		}
	}

	return null
}

const getRepeats = (text: string): repeatParsedResult => {
	const regex = /(^| )(((every|each) (([0-9]+|one|two|three|four|five|six|seven|eight|nine|ten) )?(hours?|days?|weeks?|months?|years?))|(annually|biannually|semiannually|biennially|daily|hourly|monthly|weekly|yearly))($| )/ig
	const results = regex.exec(text)
	if (results === null) {
		return {
			textWithoutMatched: text,
			repeats: null,
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

const escapeRegExp = (s: string): string => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

export const cleanupItemText = (text: string, items: string[], prefix: string): string => {
	items.forEach(l => {
		if (l === '') {
			return
		}
		const escaped = escapeRegExp(l)
		text = text
			.replace(new RegExp(`\\${prefix}'${escaped}' `, 'ig'), '')
			.replace(new RegExp(`\\${prefix}'${escaped}'`, 'ig'), '')
			.replace(new RegExp(`\\${prefix}"${escaped}" `, 'ig'), '')
			.replace(new RegExp(`\\${prefix}"${escaped}"`, 'ig'), '')
			.replace(new RegExp(`\\${prefix}${escaped} `, 'ig'), '')
			.replace(new RegExp(`\\${prefix}${escaped}`, 'ig'), '')
	})
	return text
}

const cleanupResult = (result: ParsedTaskText, prefixes: Prefixes): ParsedTaskText => {
	result.text = cleanupItemText(result.text, result.labels, prefixes.label)
	result.text = result.project !== null ? cleanupItemText(result.text, [result.project], prefixes.project) : result.text
	result.text = result.priority !== null ? cleanupItemText(result.text, [String(result.priority)], prefixes.priority) : result.text
	// Not removing assignees to avoid removing @text where the user does not exist
	result.text = result.text.trim()

	return result
}
