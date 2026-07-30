import wunderlistIcon from './icons/wunderlist.jpg'
import todoistIcon from './icons/todoist.svg?url'
import trelloIcon from './icons/trello.svg?url'
import microsoftTodoIcon from './icons/microsoft-todo.svg?url'
import vikunjaFileIcon from './icons/vikunja-file.png?url'
import tickTickIcon from './icons/ticktick.svg?url'
import wekanIcon from './icons/wekan.png?url'
import csvIcon from './icons/csv.svg?url'
import clickupIcon from './icons/clickup.svg?url'

export interface Migrator {
	id: string
	name: string
	isFileMigrator?: boolean
	isCSVMigrator?: boolean
	// True for migrators that authenticate with a pasted personal API token
	// instead of an OAuth redirect (no AuthURL to redirect to) or a file
	// upload. MigrationHandler.vue renders a plain text input for these.
	isTokenMigrator?: boolean
	icon: string
}

interface IMigratorRecord {
	[key: Migrator['id']]: Migrator
 }

export const MIGRATORS = {
	clickup: {
		id: 'clickup',
		name: 'ClickUp',
		icon: clickupIcon as string,
		isTokenMigrator: true,
	},
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
	wekan: {
		id: 'wekan',
		name: 'WeKan ®',
		icon: wekanIcon,
		isFileMigrator: true,
	},
	csv: {
		id: 'csv',
		name: 'CSV',
		icon: csvIcon as string,
		isFileMigrator: true,
		isCSVMigrator: true,
	},
} as const satisfies IMigratorRecord
