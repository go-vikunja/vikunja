import {getProjectFromPrefix, PrefixMode} from '@/modules/quickAddMagic'

export interface TaskWithParent {
	title: string,
	parentIndex: number | null,
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

	const results: TaskWithParent[] = []

	titles.forEach((title, index) => {
		const task: TaskWithParent = {
			title: cleanupTitle(title),
			parentIndex: null,
			project: null,
		}

		task.project = getProjectFromPrefix(task.title, prefixMode)

		if (index === 0) {
			results.push(task)
			return
		}

		const matched = spaceRegex.exec(task.title)
		const matchedSpaces = matched ? matched[0].length : 0

		if (matchedSpaces > 0) {
			// Go up the tree to find the first task with less indention than the current one
			let pi = 1
			let parentSpaces: number
			do {
				task.parentIndex = index - pi
				const parentTitle = cleanupTitle(titles[task.parentIndex])
				pi++
				const parentMatched = spaceRegex.exec(parentTitle)
				parentSpaces = parentMatched ? parentMatched[0].length : 0
			} while (parentSpaces >= matchedSpaces)
			task.title = cleanupTitle(task.title.replace(spaceRegex, ''))
			if (task.project === null) {
				// Inheriting the parent's resolved project instead of re-parsing its line carries
				// a project down through every level, not just the first.
				task.project = results[task.parentIndex].project
			}
		}

		results.push(task)
	})

	return results
}
