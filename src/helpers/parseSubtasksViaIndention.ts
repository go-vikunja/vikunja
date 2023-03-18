import {getProjectFromPrefix} from '@/modules/parseTaskText'

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
export function parseSubtasksViaIndention(taskTitles: string): TaskWithParent[] {
	const titles = taskTitles.split(/[\r\n]+/)

	return titles.map((title, index) => {
		const task: TaskWithParent = {
			title: cleanupTitle(title),
			parent: null,
			project: null,
		}

		task.project = getProjectFromPrefix(task.title)

		if (index === 0) {
			return task
		}

		const matched = spaceRegex.exec(title)
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
			task.title = cleanupTitle(title.replace(spaceRegex, ''))
			task.parent = task.parent.replace(spaceRegex, '')
			if (task.project === null) {
				// This allows to specify a project once for the parent task and inherit it to all subtasks
				task.project = getProjectFromPrefix(task.parent)
			}
		}

		return task
	})
}
