interface ListHistory {
	id: number;
}

export function getHistory(): ListHistory[] {
	const savedHistory = localStorage.getItem('listHistory')
	if (savedHistory === null) {
		return []
	}

	return JSON.parse(savedHistory)
}

export function saveListToHistory(list: ListHistory) {
	const history = getHistory()

	// list.id = parseInt(list.id)

	// Remove the element if it already exists in history, preventing duplicates and essentially moving it to the beginning
	history.forEach((l, i) => {
		if (l.id === list.id) {
			history.splice(i, 1)
		}
	})

	// Add the new list to the beginning of the list
	history.unshift(list)

	if (history.length > 5) {
		history.pop()
	}
	localStorage.setItem('listHistory', JSON.stringify(history))
}
