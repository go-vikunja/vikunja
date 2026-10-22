import type {MigrationKind, MigrationProvider} from '@/client/queries/migration'
import todoistIcon from './icons/todoist.svg?url'
import trelloIcon from './icons/trello.svg?url'
import microsoftTodoIcon from './icons/microsoft-todo.svg?url'
import vikunjaFileIcon from './icons/vikunja-file.png?url'
import tickTickIcon from './icons/ticktick.svg?url'
import wekanIcon from './icons/wekan.png?url'
import csvIcon from './icons/csv.svg?url'
import plankaIcon from './icons/planka.png?url'

export interface Migrator {
	id: MigrationProvider
	name: string
	kind: MigrationKind
	icon: string
}

export const MIGRATORS = {
	todoist: {
		id: 'todoist',
		name: 'Todoist',
		kind: 'oauth',
		icon: todoistIcon as string,
	},
	trello: {
		id: 'trello',
		name: 'Trello',
		kind: 'oauth',
		icon: trelloIcon as string,
	},
	'microsoft-todo': {
		id: 'microsoft-todo',
		name: 'Microsoft Todo',
		kind: 'oauth',
		icon: microsoftTodoIcon as string,
	},
	'vikunja-file': {
		id: 'vikunja-file',
		name: 'Vikunja Export',
		kind: 'file',
		icon: vikunjaFileIcon,
	},
	ticktick: {
		id: 'ticktick',
		name: 'TickTick',
		kind: 'file',
		icon: tickTickIcon as string,
	},
	wekan: {
		id: 'wekan',
		name: 'WeKan ®',
		kind: 'file',
		icon: wekanIcon,
	},
	csv: {
		id: 'csv',
		name: 'CSV',
		kind: 'csv',
		icon: csvIcon as string,
	},
	planka: {
		id: 'planka',
		name: 'Planka',
		kind: 'credentials',
		icon: plankaIcon,
	},
} as const satisfies Record<MigrationProvider, Migrator>

export type MigrationProviderOfKind<K extends MigrationKind> = {
	[P in MigrationProvider]: typeof MIGRATORS[P]['kind'] extends K ? P : never
}[MigrationProvider]

export function isMigratorOfKind<K extends MigrationKind>(
	provider: MigrationProvider,
	kind: K,
): provider is MigrationProviderOfKind<K> {
	return MIGRATORS[provider].kind === kind
}
