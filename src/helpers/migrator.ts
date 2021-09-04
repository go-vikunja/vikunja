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
			throw Error('Unknown migrator slug ' + slug)
	}
}