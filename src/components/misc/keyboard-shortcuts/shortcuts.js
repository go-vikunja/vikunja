import {isAppleDevice} from '@/helpers/isAppleDevice'

const ctrl = isAppleDevice() ? '⌘' : 'ctrl'

export const KEYBOARD_SHORTCUTS = [
	{
		title: 'keyboardShortcuts.general',
		available: () => null,
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
		available: (route) => route.name.startsWith('list.'),
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
		available: (route) => [
			'task.detail',
			'task.list.detail',
			'task.gantt.detail',
			'task.kanban.detail',
			'task.detail',
		].includes(route.name),
		shortcuts: [
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
		],
	},
]
