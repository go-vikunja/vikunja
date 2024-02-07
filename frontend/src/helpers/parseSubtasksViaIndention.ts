import {getProjectFromPrefix, PrefixMode} from '@/modules/parseTaskText'

export interface TaskWithParent {
	title: string,
	parent: string | null,
	project: string | null,
}

function cleanupTitle(title: string) {
	return title.replace(/^((\* |\+ |- )(\[ \] )?)/g, '')
}

const spaceRegex = /^ */

/**
 * @param taskTitles should be multiple lines of task tiles with indention to declare their parent/subtask
 * relation between each other.
 */
export function parseSubtasksViaIndention(taskTitles: string, prefixMode: PrefixMode): TaskWithParent[] {
	let titles = taskTitles
		.split(/[\r\n]+/)
		.filter(t => t.replace(/\s/g, '').length > 0) // Remove titles which are empty or only contain spaces / tabs
	
	if (titles.length == 0) {
		return []
	}
	
	const spaceOnFirstLine = /^(\t| )+/
	const spaces = spaceOnFirstLine.exec(titles[0])
	if (spaces !== null) {
		let spacesToCut = spaces[0].length
		titles = titles.map(title => {
			const spacesOnThisLine = spaceOnFirstLine.exec(title)
			if (spacesOnThisLine === null) {
				// This means the current task title does not start with indention, but the very first one did
				// To prevent cutting actual task data we now need to update the number of spaces to cut
				spacesToCut = 0
			}
			if (spacesOnThisLine !== null && spacesOnThisLine[0].length < spacesToCut) {
				spacesToCut = spacesOnThisLine[0].length
			}
			return title.substring(spacesToCut)
		})
	}

	return titles.map((title, index) => {
		const task: TaskWithParent = {
			title: cleanupTitle(title),
			parent: null,
			project: null,
		}

		task.project = getProjectFromPrefix(task.title, prefixMode)

		if (index === 0) {
			return task
		}

		const matched = spaceRegex.exec(task.title)
		const matchedSpaces = matched ? matched[0].length : 0

		if (matchedSpaces > 0) {
			// Go up the tree to find the first task with less indention than the current one
			let pi = 1
			let parentSpaces = 0
			do {
				task.parent = cleanupTitle(titles[index - pi])
				pi++
				const parentMatched = spaceRegex.exec(task.parent)
				parentSpaces = parentMatched ? parentMatched[0].length : 0
			} while (parentSpaces >= matchedSpaces)
			task.title = cleanupTitle(task.title.replace(spaceRegex, ''))
			task.parent = task.parent.replace(spaceRegex, '')
			if (task.project === null) {
				// This allows to specify a project once for the parent task and inherit it to all subtasks
				task.project = getProjectFromPrefix(task.parent, prefixMode)
			}
		}

		return task
	})
}
