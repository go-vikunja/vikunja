interface CheckboxStatistics {
	total: number
	checked: number
}

interface CheckboxInfo {
	index: number
	checked: boolean
	taskId: string | null
}

interface MatchedCheckboxes {
	checked: number[]
	unchecked: number[]
}

const getCheckboxesInText = (text: string): MatchedCheckboxes => {
	const regex = /data-checked="(true|false)"/g
	let match
	const checkboxes: MatchedCheckboxes = {
		checked: [],
		unchecked: [],
	}

	while ((match = regex.exec(text)) !== null) {
		if (match[1] === 'true') {
			checkboxes.checked.push(match.index)
		} else {
			checkboxes.unchecked.push(match.index)
		}
	}

	return checkboxes
}

/**
 * Returns the indices where checkboxes start and end in the given text.
 *
 * @param text
 */
export const findCheckboxesInText = (text: string): number[] => {
	const checkboxes = getCheckboxesInText(text)

	return [
		...checkboxes.checked,
		...checkboxes.unchecked,
	].sort((a, b) => a - b)
}

export const getChecklistStatistics = (text: string): CheckboxStatistics => {
	const checkboxes = getCheckboxesInText(text)

	return {
		total: checkboxes.checked.length + checkboxes.unchecked.length,
		checked: checkboxes.checked.length,
	}
}

/**
 * Returns detailed checkbox info including task IDs.
 */
export const getCheckboxesWithIds = (text: string): CheckboxInfo[] => {
	// Match <li> tags with data-checked attribute
	const liRegex = /<li[^>]*data-checked="(true|false)"[^>]*>/g
	const taskIdRegex = /data-task-id="([^"]*)"/
	const checkboxes: CheckboxInfo[] = []
	let match

	while ((match = liRegex.exec(text)) !== null) {
		const liTag = match[0]
		const taskIdMatch = taskIdRegex.exec(liTag)

		checkboxes.push({
			index: match.index,
			checked: match[1] === 'true',
			taskId: taskIdMatch ? taskIdMatch[1] : null,
		})
	}

	return checkboxes
}
