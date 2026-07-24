import {isAppleDevice} from '@/helpers/isAppleDevice'

export type ShortcutDefinition = {
	binding: string
	keys: string[]
	combination?: 'then'
}

export const PRIMARY_MODIFIER_KEY = isAppleDevice() ? '⌘' : 'ctrl'

export const SHORTCUTS = {
	toggleMenu: {
		binding: 'Mod+KeyE',
		keys: [PRIMARY_MODIFIER_KEY, 'e'],
	},
	quickSearch: {
		binding: 'Mod+KeyK',
		keys: [PRIMARY_MODIFIER_KEY, 'k'],
	},
	showKeyboardShortcuts: {
		binding: 'Shift+Slash',
		keys: ['shift', '/'],
	},
	navigation: {
		overview: {
			binding: 'KeyG KeyO',
			keys: ['g', 'o'],
			combination: 'then',
		},
		upcoming: {
			binding: 'KeyG KeyU',
			keys: ['g', 'u'],
			combination: 'then',
		},
		projects: {
			binding: 'KeyG KeyP',
			keys: ['g', 'p'],
			combination: 'then',
		},
		labels: {
			binding: 'KeyG KeyA',
			keys: ['g', 'a'],
			combination: 'then',
		},
		teams: {
			binding: 'KeyG KeyM',
			keys: ['g', 'm'],
			combination: 'then',
		},
	},
	taskDetail: {
		openProject: {
			binding: 'KeyU',
			keys: ['u'],
		},
		done: {
			binding: 'KeyT',
			keys: ['t'],
		},
		favorite: {
			binding: 'KeyS',
			keys: ['s'],
		},
		labels: {
			binding: 'KeyL',
			keys: ['l'],
		},
		priority: {
			binding: 'KeyP',
			keys: ['p'],
		},
		color: {
			binding: 'KeyC',
			keys: ['c'],
		},
		assignees: {
			binding: 'KeyA',
			keys: ['a'],
		},
		attachments: {
			binding: 'KeyF',
			keys: ['f'],
		},
		relatedTasks: {
			binding: 'KeyR',
			keys: ['r'],
		},
		moveProject: {
			binding: 'KeyM',
			keys: ['m'],
		},
		dueDate: {
			binding: 'KeyD',
			keys: ['d'],
		},
		reminder: {
			binding: isAppleDevice() ? 'Shift+KeyR' : 'Alt+KeyR',
			keys: [isAppleDevice() ? 'shift' : 'alt', 'r'],
		},
		delete: {
			binding: isAppleDevice() ? 'Backspace' : 'Delete',
			keys: [isAppleDevice() ? 'backspace' : 'delete'],
		},
	},
} as const satisfies Record<string, ShortcutDefinition | Record<string, ShortcutDefinition>>
