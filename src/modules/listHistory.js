export const getHistory = () => {
	const savedHistory = localStorage.getItem('listHistory')
	if (savedHistory === null) {
		return []
	}

	return JSON.parse(savedHistory)
}

export function saveListToHistory(list) {
	const history = getHistory()

	// Remove the element if it already exists in history, preventing duplicates and essentially moving it to the beginning
	for (const i in history) {
		if (history[i].id === list.id) {
			history.splice(i, 1)
		}
	}

	history.unshift(list)

	if (history.length > 5) {
		history.pop()
	}
	localStorage.setItem('listHistory', JSON.stringify(history))
}
