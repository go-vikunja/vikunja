import {isAppleDevice} from '@/helpers/isAppleDevice'

export const PRIMARY_MODIFIER_KEY = isAppleDevice() ? '⌘' : 'ctrl'

export const SHORTCUTS = {
	toggleMenu: 'Mod+KeyE',
	quickSearch: 'Mod+KeyK',
	showKeyboardShortcuts: 'Shift+Slash',
	navigation: {
		overview: 'KeyG KeyO',
		upcoming: 'KeyG KeyU',
		projects: 'KeyG KeyP',
		labels: 'KeyG KeyA',
		teams: 'KeyG KeyM',
	},
	taskDetail: {
		openProject: 'KeyU',
		done: 'KeyT',
		favorite: 'KeyS',
		labels: 'KeyL',
		priority: 'KeyP',
		color: 'KeyC',
		assignees: 'KeyA',
		attachments: 'KeyF',
		relatedTasks: 'KeyR',
		moveProject: 'KeyM',
		dueDate: 'KeyD',
		reminder: isAppleDevice() ? 'Shift+KeyR' : 'Alt+KeyR',
		delete: isAppleDevice() ? 'Backspace' : 'Delete',
	},
} as const
