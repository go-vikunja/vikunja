import {RouteLocation} from 'vue-router'

import {isAppleDevice} from '@/helpers/isAppleDevice'

const ctrl = isAppleDevice() ? '⌘' : 'ctrl'

interface Shortcut {
	title: string
	keys: string[]
	combination?: 'then'
}

interface ShortcutGroup {
	title: string
	available?: (route: RouteLocation) => boolean
	shortcuts: Shortcut[]
}

export const KEYBOARD_SHORTCUTS : ShortcutGroup[] = [
	{
		title: 'keyboardShortcuts.general',
		shortcuts: [
			{
				title: 'keyboardShortcuts.toggleMenu',
				keys: [ctrl, 'e'],
			},
			{
				title: 'keyboardShortcuts.quickSearch',
				keys: [ctrl, 'k'],
			},
		],
	},
	{
		title: 'keyboardShortcuts.navigation.title',
		shortcuts: [
			{
				title: 'keyboardShortcuts.navigation.overview',
				keys: ['g', 'o'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.navigation.upcoming',
				keys: ['g', 'u'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.navigation.namespaces',
				keys: ['g', 'n'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.navigation.labels',
				keys: ['g', 'a'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.navigation.teams',
				keys: ['g', 'm'],
				combination: 'then',
			},
		],
	},
	{
		title: 'list.kanban.title',
		available: (route) => route.name === 'list.kanban',
		shortcuts: [
			{
				title: 'keyboardShortcuts.task.done',
				keys: [ctrl, 'click'],
			},
		],
	},
	{
		title: 'keyboardShortcuts.list.title',
		available: (route) => (route.name as string)?.startsWith('list.'),
		shortcuts: [
			{
				title: 'keyboardShortcuts.list.switchToListView',
				keys: ['g', 'l'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.list.switchToGanttView',
				keys: ['g', 'g'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.list.switchToTableView',
				keys: ['g', 't'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.list.switchToKanbanView',
				keys: ['g', 'k'],
				combination: 'then',
			},
		],
	},
	{
		title: 'keyboardShortcuts.task.title',
		available: (route) => route.name === 'task.detail',
		shortcuts: [
			{
				title: 'keyboardShortcuts.task.done',
				keys: ['t'],
			},
			{
				title: 'keyboardShortcuts.task.assign',
				keys: ['a'],
			},
			{
				title: 'keyboardShortcuts.task.labels',
				keys: ['l'],
			},
			{
				title: 'keyboardShortcuts.task.dueDate',
				keys: ['d'],
			},
			{
				title: 'keyboardShortcuts.task.attachment',
				keys: ['f'],
			},
			{
				title: 'keyboardShortcuts.task.related',
				keys: ['r'],
			},
			{
				title: 'keyboardShortcuts.task.move',
				keys: ['m'],
			},
			{
				title: 'keyboardShortcuts.task.color',
				keys: ['c'],
			},
		],
	},
]
