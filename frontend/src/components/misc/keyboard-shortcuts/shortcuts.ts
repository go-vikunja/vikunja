import type {RouteLocation} from 'vue-router'

import {PRIMARY_MODIFIER_KEY, SHORTCUTS} from '@/constants/shortcuts'

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
				keys: SHORTCUTS.toggleMenu.keys,
			},
			{
				title: 'keyboardShortcuts.quickSearch',
				keys: SHORTCUTS.quickSearch.keys,
			},
		],
	},
	{
		title: 'keyboardShortcuts.navigation.title',
		shortcuts: [
			{
				title: 'keyboardShortcuts.navigation.overview',
				keys: SHORTCUTS.navigation.overview.keys,
				combination: SHORTCUTS.navigation.overview.combination,
			},
			{
				title: 'keyboardShortcuts.navigation.upcoming',
				keys: SHORTCUTS.navigation.upcoming.keys,
				combination: SHORTCUTS.navigation.upcoming.combination,
			},
			{
				title: 'keyboardShortcuts.navigation.projects',
				keys: SHORTCUTS.navigation.projects.keys,
				combination: SHORTCUTS.navigation.projects.combination,
			},
			{
				title: 'keyboardShortcuts.navigation.labels',
				keys: SHORTCUTS.navigation.labels.keys,
				combination: SHORTCUTS.navigation.labels.combination,
			},
			{
				title: 'keyboardShortcuts.navigation.teams',
				keys: SHORTCUTS.navigation.teams.keys,
				combination: SHORTCUTS.navigation.teams.combination,
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
				keys: SHORTCUTS.taskDetail.done.keys,
			},
			{
				title: 'keyboardShortcuts.task.assign',
				keys: SHORTCUTS.taskDetail.assignees.keys,
			},
			{
				title: 'keyboardShortcuts.task.labels',
				keys: SHORTCUTS.taskDetail.labels.keys,
			},
			{
				title: 'keyboardShortcuts.task.dueDate',
				keys: SHORTCUTS.taskDetail.dueDate.keys,
			},
			{
				title: 'keyboardShortcuts.task.attachment',
				keys: SHORTCUTS.taskDetail.attachments.keys,
			},
			{
				title: 'keyboardShortcuts.task.related',
				keys: SHORTCUTS.taskDetail.relatedTasks.keys,
			},
			{
				title: 'keyboardShortcuts.task.move',
				keys: SHORTCUTS.taskDetail.moveProject.keys,
			},
			{
				title: 'keyboardShortcuts.task.color',
				keys: SHORTCUTS.taskDetail.color.keys,
			},
			{
				title: 'keyboardShortcuts.task.reminder',
				keys: SHORTCUTS.taskDetail.reminder.keys,
			},
			{
				title: 'keyboardShortcuts.task.description',
				keys: ['e'],
			},
			{
				title: 'keyboardShortcuts.task.priority',
				keys: SHORTCUTS.taskDetail.priority.keys,
			},
			{
				title: 'keyboardShortcuts.task.delete',
				keys: SHORTCUTS.taskDetail.delete.keys,
			},
			{
				title: 'keyboardShortcuts.task.favorite',
				keys: SHORTCUTS.taskDetail.favorite.keys,
			},
			{
				title: 'keyboardShortcuts.task.openProject',
				keys: SHORTCUTS.taskDetail.openProject.keys,
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
