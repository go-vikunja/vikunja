export interface ProjectHistory {
	id: number;
}

export function getHistory(): ProjectHistory[] {
	const savedHistory = localStorage.getItem('projectHistory')
	if (savedHistory === null) {
		return []
	}

	return JSON.parse(savedHistory)
}

function saveHistory(history: ProjectHistory[]) {
	if (history.length === 0) {
		localStorage.removeItem('projectHistory')
		return
	}

	localStorage.setItem('projectHistory', JSON.stringify(history))
}

export function saveProjectToHistory(project: ProjectHistory) {
	const history: ProjectHistory[] = getHistory()

	// Remove the element if it already exists in history, preventing duplicates and essentially moving it to the beginning
	history.forEach((l, i) => {
		if (l.id === project.id) {
			history.splice(i, 1)
		}
	})

	// Add the new project to the beginning of the project
	history.unshift(project)

	if (history.length > 5) {
		history.pop()
	}
	saveHistory(history)
}

export function removeProjectFromHistory(project: ProjectHistory) {
	const history: ProjectHistory[] = getHistory()

	history.forEach((l, i) => {
		if (l.id === project.id) {
			history.splice(i, 1)
		}
	})
	saveHistory(history)
}
