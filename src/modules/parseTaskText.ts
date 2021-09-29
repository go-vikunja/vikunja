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

interface ParsedTaskText {
	text: string,
	date: Date | null,
	labels: string[],
	list: string | null,
	priority: number | null,
	assignees: string[],
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

	const {newText, date} = parseDate(text)
	result.text = newText
	result.date = date

	return cleanupResult(result, prefixes)
}

const getItemsFromPrefix = (text: string, prefix: string): string[] => {
	const items: string[] = []

	const itemParts = text.split(prefix)
	itemParts.forEach((p, index) => {
		// First part contains the rest
		if (index < 1) {
			return
		}

		let labelText
		if (p.charAt(0) === '\'') {
			labelText = p.split('\'')[1]
		} else if (p.charAt(0) === '"') {
			labelText = p.split('"')[1]
		} else {
			// Only until the next space
			labelText = p.split(' ')[0]
		}
		items.push(labelText)
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
