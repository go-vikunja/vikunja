import wunderlistIcon from './icons/wunderlist.jpg'
import todoistIcon from './icons/todoist.svg?url'
import trelloIcon from './icons/trello.svg?url'
import microsoftTodoIcon from './icons/microsoft-todo.svg?url'
import vikunjaFileIcon from './icons/vikunja-file.png?url'
import tickTickIcon from './icons/ticktick.svg?url'

export interface Migrator {
	id: string
	name: string
	isFileMigrator?: boolean
	icon: string
}

interface IMigratorRecord {
	[key: Migrator['id']]: Migrator
 }

export const MIGRATORS = {
	wunderlist: {
		id: 'wunderlist',
		name: 'Wunderlist',
		icon: wunderlistIcon,
	},
	todoist: {
		id: 'todoist',
		name: 'Todoist',
		icon: todoistIcon as string,
	},
	trello: {
		id: 'trello',
		name: 'Trello',
		icon: trelloIcon as string,
	},
	'microsoft-todo': {
		id: 'microsoft-todo',
		name: 'Microsoft Todo',
		icon: microsoftTodoIcon as string,
	},
	'vikunja-file': {
		id: 'vikunja-file',
		name: 'Vikunja Export',
		icon: vikunjaFileIcon,
		isFileMigrator: true,
	},
	ticktick: {
		id: 'ticktick',
		name: 'TickTick',
		icon: tickTickIcon as string,
		isFileMigrator: true,
	},
} as const satisfies IMigratorRecord
