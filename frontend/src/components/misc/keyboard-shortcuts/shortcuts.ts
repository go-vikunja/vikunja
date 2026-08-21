import type {RouteLocation} from 'vue-router'

import {PRIMARY_MODIFIER_KEY, SHORTCUTS} from '@/constants/shortcuts'
import {shortcutBindingToDisplay} from '@/helpers/shortcut'

const ctrl = PRIMARY_MODIFIER_KEY

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
				...shortcutBindingToDisplay(SHORTCUTS.toggleMenu),
			},
			{
				title: 'keyboardShortcuts.quickSearch',
				...shortcutBindingToDisplay(SHORTCUTS.quickSearch),
			},
		],
	},
	{
		title: 'keyboardShortcuts.navigation.title',
		shortcuts: [
			{
				title: 'keyboardShortcuts.navigation.overview',
				...shortcutBindingToDisplay(SHORTCUTS.navigation.overview),
			},
			{
				title: 'keyboardShortcuts.navigation.upcoming',
				...shortcutBindingToDisplay(SHORTCUTS.navigation.upcoming),
			},
			{
				title: 'keyboardShortcuts.navigation.projects',
				...shortcutBindingToDisplay(SHORTCUTS.navigation.projects),
			},
			{
				title: 'keyboardShortcuts.navigation.labels',
				...shortcutBindingToDisplay(SHORTCUTS.navigation.labels),
			},
			{
				title: 'keyboardShortcuts.navigation.teams',
				...shortcutBindingToDisplay(SHORTCUTS.navigation.teams),
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
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.done),
			},
			{
				title: 'keyboardShortcuts.task.assign',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.assignees),
			},
			{
				title: 'keyboardShortcuts.task.labels',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.labels),
			},
			{
				title: 'keyboardShortcuts.task.dueDate',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.dueDate),
			},
			{
				title: 'keyboardShortcuts.task.attachment',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.attachments),
			},
			{
				title: 'keyboardShortcuts.task.related',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.relatedTasks),
			},
			{
				title: 'keyboardShortcuts.task.move',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.moveProject),
			},
			{
				title: 'keyboardShortcuts.task.color',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.color),
			},
			{
				title: 'keyboardShortcuts.task.reminder',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.reminder),
			},
			{
				title: 'keyboardShortcuts.task.description',
				keys: ['e'],
			},
			{
				title: 'keyboardShortcuts.task.priority',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.priority),
			},
			{
				title: 'keyboardShortcuts.task.delete',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.delete),
			},
			{
				title: 'keyboardShortcuts.task.favorite',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.favorite),
			},
			{
				title: 'keyboardShortcuts.task.openProject',
				...shortcutBindingToDisplay(SHORTCUTS.taskDetail.openProject),
			},
			{
				title: 'keyboardShortcuts.task.save',
				keys: [ctrl, 's'],
			},
			{
				title: 'keyboardShortcuts.task.copyIdentifier',
				keys: ['.'],
			},
			{
				title: 'keyboardShortcuts.task.copyIdentifierAndTitle',
				keys: ['.', '.'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.task.copyIdentifierTitleAndUrl',
				keys: ['.', '.', '.'],
				combination: 'then',
			},
			{
				title: 'keyboardShortcuts.task.copyUrl',
				keys: [ctrl, '.'],
			},
		],
	},
] as const
