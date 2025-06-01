interface CheckboxStatistics {
	total: number
	checked: number
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
