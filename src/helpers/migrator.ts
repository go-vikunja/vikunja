export interface Migrator {
	name: string
	identifier: string
	isFileMigrator?: boolean
}

export const getMigratorFromSlug = (slug: string): Migrator => {
	switch (slug) {
		case 'wunderlist':
			return {
				name: 'Wunderlist',
				identifier: 'wunderlist',
			}
		case 'todoist':
			return {
				name: 'Todoist',
				identifier: 'todoist',
			}
		case 'trello':
			return {
				name: 'Trello',
				identifier: 'trello',
			}
		case 'microsoft-todo':
			return {
				name: 'Microsoft Todo',
				identifier: 'microsoft-todo',
			}
		case 'vikunja-file':
			return {
				name: 'Vikunja Export',
				identifier: 'vikunja-file',
				isFileMigrator: true,
			}
		default:
			throw new Error('Unknown migrator slug ' + slug)
	}
}


// NOTE: we list the imports individually for better build time optimisation
export const SERVICE_ICONS = {
	'vikunja-file': () => import('@/assets/migration/vikunja-file.png'),
	'microsoft-todo': () => import('@/assets/migration/microsoft-todo.svg'),
	'todoist': () => import('@/assets/migration/todoist.svg'),
	'trello': () => import('@/assets/migration/trello.svg'),
	'wunderlist': () => import('@/assets/migration/wunderlist.jpg'),
}