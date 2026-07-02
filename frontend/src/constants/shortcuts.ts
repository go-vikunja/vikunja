import {isAppleDevice} from '@/helpers/isAppleDevice'

type ShortcutDefinition = {
	binding: string
	keys: string[]
	combination?: 'then'
}

const ctrl = isAppleDevice() ? '⌘' : 'ctrl'

export const SHORTCUTS = {
	toggleMenu: {
		binding: 'Mod+KeyE',
		keys: [ctrl, 'e'],
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
} as const satisfies Record<string, ShortcutDefinition | Record<string, ShortcutDefinition>>
