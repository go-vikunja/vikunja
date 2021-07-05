import {parseDate} from './time/parseDate'
import priorities from '../models/priorities.json'

const LABEL_PREFIX = '~'
const LIST_PREFIX = '*'
const PRIORITY_PREFIX = '!'
const ASSIGNEE_PREFIX = '@'

/**
 * Parses task text for dates, assignees, labels, lists, priorities and returns an object with all found intents.
 *
 * @param text
 */
export const parseTaskText = text => {
	const result = {
		text: text,
		date: null,
		labels: [],
		list: null,
		priority: null,
		assignees: [],
	}

	result.labels = getItemsFromPrefix(text, LABEL_PREFIX)

	const lists = getItemsFromPrefix(text, LIST_PREFIX)
	result.list = lists.length > 0 ? lists[0] : null

	result.priority = getPriority(text)

	result.assignees = getItemsFromPrefix(text, ASSIGNEE_PREFIX)

	const {newText, date} = parseDate(text)
	result.text = newText
	result.date = date

	return cleanupResult(result)
}

const getItemsFromPrefix = (text, prefix) => {
	const items = []

	const itemParts = text.split(prefix)
	itemParts.forEach((p, index) => {
		// First part contains the rest
		if (index < 1) {
			return
		}

		let labelText
		if (p.charAt(0) === `'`) {
			labelText = p.split(`'`)[1]
		} else if (p.charAt(0) === `"`) {
			labelText = p.split(`"`)[1]
		} else {
			// Only until the next space
			labelText = p.split(' ')[0]
		}
		items.push(labelText)
	})

	return Array.from(new Set(items))
}

const getPriority = text => {
	const ps = getItemsFromPrefix(text, PRIORITY_PREFIX)
	if (ps.length === 0) {
		return null
	}

	for (const p of ps) {
		for (const pi in priorities) {
			if (priorities[pi] === parseInt(p)) {
				return parseInt(p)
			}
		}
	}

	return null
}

const cleanupItemText = (text, items, prefix) => {
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

const cleanupResult = result => {
	result.text = cleanupItemText(result.text, result.labels, LABEL_PREFIX)
	result.text = cleanupItemText(result.text, [result.list], LIST_PREFIX)
	result.text = cleanupItemText(result.text, [result.priority], PRIORITY_PREFIX)
	result.text = cleanupItemText(result.text, result.assignees, ASSIGNEE_PREFIX)
	result.text = result.text.trim()

	return result
}
