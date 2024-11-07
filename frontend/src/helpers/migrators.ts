import wunderlistIcon from '@/assets/migrators/wunderlist.jpg'
import todoistIcon from '@/assets/migrators/todoist.svg?url'
import trelloIcon from '@/assets/migrators/trello.svg?url'
import microsoftTodoIcon from '@/assets/migrators/microsoft-todo.svg?url'
import vikunjaFileIcon from '@/assets/migrators/vikunja-file.png?url'
import tickTickIcon from '@/assets/migrators/ticktick.svg?url'

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
