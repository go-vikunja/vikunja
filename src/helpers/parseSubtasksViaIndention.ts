export interface TaskWithParent {
	title: string,
	parent: string | null,
}

function cleanupTitle(title: string) {
	return title.replace(/^((\* |\+ |- )(\[ \] )?)/g, '')
}

const spaceRegex = /^ */

// taskTitles should be multiple lines of task tiles with indention to declare their parent/subtask 
// relation between each other.
export function parseSubtasksViaIndention(taskTitles: string): TaskWithParent[] {
	const titles = taskTitles.split(/[\r\n]+/)

	return titles.map((t, i) => {
		const task: TaskWithParent = {
			title: cleanupTitle(t),
			parent: null,
		}

		const matched = spaceRegex.exec(t)
		const matchedSpaces = matched ? matched[0].length : 0

		if (matchedSpaces > 0 && i > 0) {
			// Go up the tree to find the first task with less indention than the current one
			let pi = 1
			let parentSpaces = 0
			do {
				task.parent = cleanupTitle(titles[i - pi])
				pi++
				const parentMatched = spaceRegex.exec(task.parent)
				parentSpaces = parentMatched ? parentMatched[0].length : 0
			} while (parentSpaces >= matchedSpaces)
			task.title = cleanupTitle(t.replace(spaceRegex, ''))
			task.parent = task.parent.replace(spaceRegex, '')
		}

		return task
	})
}
