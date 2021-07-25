import {parseDate} from '../helpers/time/parseDate'
import _priorities from '../models/priorities.json'

const LABEL_PREFIX: string = '@'
const LIST_PREFIX: string = '#'
const PRIORITY_PREFIX: string = '!'
const ASSIGNEE_PREFIX: string = '+'

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

/**
 * Parses task text for dates, assignees, labels, lists, priorities and returns an object with all found intents.
 *
 * @param text
 */
export const parseTaskText = (text: string): ParsedTaskText => {
	const result: ParsedTaskText = {
		text: text,
		date: null,
		labels: [],
		list: null,
		priority: null,
		assignees: [],
	}

	result.labels = getItemsFromPrefix(text, LABEL_PREFIX)

	const lists: string[] = getItemsFromPrefix(text, LIST_PREFIX)
	result.list = lists.length > 0 ? lists[0] : null

	result.priority = getPriority(text)

	result.assignees = getItemsFromPrefix(text, ASSIGNEE_PREFIX)

	const {newText, date} = parseDate(text)
	result.text = newText
	result.date = date

	return cleanupResult(result)
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

const getPriority = (text: string): number | null => {
	const ps = getItemsFromPrefix(text, PRIORITY_PREFIX)
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

const cleanupResult = (result: ParsedTaskText): ParsedTaskText => {
	result.text = cleanupItemText(result.text, result.labels, LABEL_PREFIX)
	result.text = result.list !== null ? cleanupItemText(result.text, [result.list], LIST_PREFIX) : result.text
	result.text = result.priority !== null ? cleanupItemText(result.text, [String(result.priority)], PRIORITY_PREFIX) : result.text
	result.text = cleanupItemText(result.text, result.assignees, ASSIGNEE_PREFIX)
	result.text = result.text.trim()

	return result
}
