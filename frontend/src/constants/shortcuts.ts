import {isAppleDevice} from '@/helpers/isAppleDevice'

export const PRIMARY_MODIFIER_KEY = isAppleDevice() ? '⌘' : 'ctrl'

export const SHORTCUTS = {
	toggleMenu: {
		binding: 'Mod+KeyE',
	},
	quickSearch: {
		binding: 'Mod+KeyK',
	},
	showKeyboardShortcuts: {
		binding: 'Shift+Slash',
	},
	navigation: {
		overview: {
			binding: 'KeyG KeyO',
		},
		upcoming: {
			binding: 'KeyG KeyU',
		},
		projects: {
			binding: 'KeyG KeyP',
		},
		labels: {
			binding: 'KeyG KeyA',
		},
		teams: {
			binding: 'KeyG KeyM',
		},
	},
	taskDetail: {
		openProject: {
			binding: 'KeyU',
		},
		done: {
			binding: 'KeyT',
		},
		favorite: {
			binding: 'KeyS',
		},
		labels: {
			binding: 'KeyL',
		},
		priority: {
			binding: 'KeyP',
		},
		color: {
			binding: 'KeyC',
		},
		assignees: {
			binding: 'KeyA',
		},
		attachments: {
			binding: 'KeyF',
		},
		relatedTasks: {
			binding: 'KeyR',
		},
		moveProject: {
			binding: 'KeyM',
		},
		dueDate: {
			binding: 'KeyD',
		},
		reminder: {
			binding: isAppleDevice() ? 'Shift+KeyR' : 'Alt+KeyR',
		},
		delete: {
			binding: isAppleDevice() ? 'Backspace' : 'Delete',
		},
	},
} as const
