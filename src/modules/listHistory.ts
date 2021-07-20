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

function saveHistory(history: ListHistory[]) {
	if (history.length === 0) {
		localStorage.removeItem('listHistory')
		return
	}

	localStorage.setItem('listHistory', JSON.stringify(history))
}

export function saveListToHistory(list: ListHistory) {
	const history: ListHistory[] = getHistory()

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
	saveHistory(history)
}

export function removeListFromHistory(list: ListHistory) {
	const history: ListHistory[] = getHistory()

	history.forEach((l, i) => {
		if (l.id === list.id) {
			history.splice(i, 1)
		}
	})
	saveHistory(history)
}
