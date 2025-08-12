import type {RouteLocation} from 'vue-router'

import {isAppleDevice} from '@/helpers/isAppleDevice'

const ctrl = isAppleDevice() ? '⌘' : 'ctrl'

export interface Shortcut {
	title: string
	keys: string[]
	combination?: 'then'
}

export interface ShortcutGroup {
	title: string
	available?: (route: RouteLocation) => boolean
	shortcuts: Shortcut[]
}

export const KEYBOARD_SHORTCUTS: ShortcutGroup[] = [
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
				title: 'keyboardShortcuts.navigation.projects',
				keys: ['g', 'p'],
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
		title: 'keyboardShortcuts.list.title',
		available: (route) => route.name === 'project.view',
		shortcuts: [
			{
				title: 'keyboardShortcuts.list.navigateDown',
				keys: ['j'],
			},
			{
				title: 'keyboardShortcuts.list.navigateUp',
				keys: ['k'],
			},
			{
				title: 'keyboardShortcuts.list.open',
				keys: ['enter'],
			},
		],
	},
	{
		title: 'project.kanban.title',
		available: (route) => route.name === 'project.view',
		shortcuts: [
			{
				title: 'keyboardShortcuts.task.done',
				keys: [ctrl, 'click'],
			},
		],
	},
	{
		title: 'keyboardShortcuts.project.title',
		available: (route) => (route.name as string)?.startsWith('project.'),
		shortcuts: [
			{
				title: 'keyboardShortcuts.project.switchToListView',
				keys: ['g', 'l'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.project.switchToGanttView',
				keys: ['g', 'g'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.project.switchToTableView',
				keys: ['g', 't'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.project.switchToKanbanView',
				keys: ['g', 'k'],
				combination: 'then',
			},
		],
	},
	{
		title: 'keyboardShortcuts.gantt.title',
		available: (route) => route.name === 'project.view',
		shortcuts: [
			{
				title: 'keyboardShortcuts.gantt.moveTaskLeft',
				keys: ['←'],
			},
			{
				title: 'keyboardShortcuts.gantt.moveTaskRight',
				keys: ['→'],
			},
			{
				title: 'keyboardShortcuts.gantt.expandTaskLeft',
				keys: ['shift', '←'],
			},
			{
				title: 'keyboardShortcuts.gantt.expandTaskRight',
				keys: ['shift', '→'],
			},
			{
				title: 'keyboardShortcuts.gantt.shrinkTaskLeft',
				keys: [ctrl, '←'],
			},
			{
				title: 'keyboardShortcuts.gantt.shrinkTaskRight',
				keys: [ctrl, '→'],
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
			{
				title: 'keyboardShortcuts.task.reminder',
				keys: ['alt', 'r'],
			},
			{
				title: 'keyboardShortcuts.task.description',
				keys: ['e'],
			},
			{
				title: 'keyboardShortcuts.task.priority',
				keys: ['p'],
			},
			{
				title: 'keyboardShortcuts.task.delete',
				keys: ['shift', 'delete'],
			},
			{
				title: 'keyboardShortcuts.task.favorite',
				keys: ['s'],
			},
			{
				title: 'keyboardShortcuts.task.openProject',
				keys: ['u'],
			},
			{
				title: 'keyboardShortcuts.task.save',
				keys: [ctrl, 's'],
			},
		],
	},
] as const
