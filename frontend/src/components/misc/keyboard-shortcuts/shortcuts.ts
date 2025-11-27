import type {RouteLocation} from 'vue-router'

import {isAppleDevice} from '@/helpers/isAppleDevice'

const ctrl = isAppleDevice() ? '⌘' : 'ctrl'
const reminderModifier = isAppleDevice() ? 'shift' : 'alt'

export enum ShortcutCategory {
	GENERAL = 'general',
	NAVIGATION = 'navigation',
	TASK_ACTIONS = 'taskActions',
	PROJECT_VIEWS = 'projectViews',
	LIST_VIEW = 'listView',
	GANTT_VIEW = 'ganttView',
}

export interface ShortcutAction {
	actionId: string           // Unique ID like "general.toggleMenu"
	title: string             // i18n key for display
	keys: string[]            // Default keys
	customizable: boolean     // Can user customize this?
	contexts?: string[]       // Which routes/contexts apply
	category: ShortcutCategory
	combination?: 'then'      // For multi-key sequences
}

export interface Shortcut {
	title: string
	keys: string[]
	combination?: 'then'
}

export interface ShortcutGroup {
	title: string
	category: ShortcutCategory
	available?: (route: RouteLocation) => boolean
	shortcuts: ShortcutAction[]
}

export const KEYBOARD_SHORTCUTS: ShortcutGroup[] = [
	{
		title: 'keyboardShortcuts.general',
		category: ShortcutCategory.GENERAL,
		shortcuts: [
			{
				actionId: 'general.toggleMenu',
				title: 'keyboardShortcuts.toggleMenu',
				keys: [ctrl, 'e'],
				customizable: true,
				contexts: ['*'],
				category: ShortcutCategory.GENERAL,
			},
			{
				actionId: 'general.quickSearch',
				title: 'keyboardShortcuts.quickSearch',
				keys: [ctrl, 'k'],
				customizable: true,
				contexts: ['*'],
				category: ShortcutCategory.GENERAL,
			},
			{
				actionId: 'general.showHelp',
				title: 'keyboardShortcuts.showHelp',
				keys: ['shift', '?'],
				customizable: true,
				contexts: ['*'],
				category: ShortcutCategory.GENERAL,
			},
		],
	},
	{
		title: 'keyboardShortcuts.navigation.title',
		category: ShortcutCategory.NAVIGATION,
		shortcuts: [
			{
				actionId: 'navigation.goToOverview',
				title: 'keyboardShortcuts.navigation.overview',
				keys: ['g', 'o'],
				combination: 'then',
				customizable: false,  // Navigation shortcuts are fixed
				contexts: ['*'],
				category: ShortcutCategory.NAVIGATION,
			},
			{
				actionId: 'navigation.goToUpcoming',
				title: 'keyboardShortcuts.navigation.upcoming',
				keys: ['g', 'u'],
				combination: 'then',
				customizable: false,
				contexts: ['*'],
				category: ShortcutCategory.NAVIGATION,
			},
			{
				actionId: 'navigation.goToProjects',
				title: 'keyboardShortcuts.navigation.projects',
				keys: ['g', 'p'],
				combination: 'then',
				customizable: false,
				contexts: ['*'],
				category: ShortcutCategory.NAVIGATION,
			},
			{
				actionId: 'navigation.goToLabels',
				title: 'keyboardShortcuts.navigation.labels',
				keys: ['g', 'a'],
				combination: 'then',
				customizable: false,
				contexts: ['*'],
				category: ShortcutCategory.NAVIGATION,
			},
			{
				actionId: 'navigation.goToTeams',
				title: 'keyboardShortcuts.navigation.teams',
				keys: ['g', 'm'],
				combination: 'then',
				customizable: false,
				contexts: ['*'],
				category: ShortcutCategory.NAVIGATION,
			},
		],
	},
	{
		title: 'keyboardShortcuts.list.title',
		category: ShortcutCategory.LIST_VIEW,
		available: (route) => route.name === 'project.view',
		shortcuts: [
			{
				actionId: 'listView.nextTask',
				title: 'keyboardShortcuts.list.navigateDown',
				keys: ['j'],
				customizable: false,  // List navigation is fixed
				contexts: ['/projects/:id/list'],
				category: ShortcutCategory.LIST_VIEW,
			},
			{
				actionId: 'listView.previousTask',
				title: 'keyboardShortcuts.list.navigateUp',
				keys: ['k'],
				customizable: false,
				contexts: ['/projects/:id/list'],
				category: ShortcutCategory.LIST_VIEW,
			},
			{
				actionId: 'listView.openTask',
				title: 'keyboardShortcuts.list.open',
				keys: ['enter'],
				customizable: false,
				contexts: ['/projects/:id/list'],
				category: ShortcutCategory.LIST_VIEW,
			},
		],
	},
	{
		title: 'project.kanban.title',
		category: ShortcutCategory.PROJECT_VIEWS,
		available: (route) => route.name === 'project.view',
		shortcuts: [
			{
				actionId: 'kanban.markTaskDone',
				title: 'keyboardShortcuts.task.done',
				keys: [ctrl, 'click'],
				customizable: false,  // Mouse combinations are not customizable
				contexts: ['/projects/:id/kanban'],
				category: ShortcutCategory.PROJECT_VIEWS,
			},
		],
	},
	{
		title: 'keyboardShortcuts.project.title',
		category: ShortcutCategory.PROJECT_VIEWS,
		available: (route) => (route.name as string)?.startsWith('project.'),
		shortcuts: [
			{
				actionId: 'projectViews.switchToList',
				title: 'keyboardShortcuts.project.switchToListView',
				keys: ['g', 'l'],
				combination: 'then',
				customizable: false,  // Navigation shortcuts are fixed
				contexts: ['/projects/:id/*'],
				category: ShortcutCategory.PROJECT_VIEWS,
			},
			{
				actionId: 'projectViews.switchToGantt',
				title: 'keyboardShortcuts.project.switchToGanttView',
				keys: ['g', 'g'],
				combination: 'then',
				customizable: false,
				contexts: ['/projects/:id/*'],
				category: ShortcutCategory.PROJECT_VIEWS,
			},
			{
				actionId: 'projectViews.switchToTable',
				title: 'keyboardShortcuts.project.switchToTableView',
				keys: ['g', 't'],
				combination: 'then',
				customizable: false,
				contexts: ['/projects/:id/*'],
				category: ShortcutCategory.PROJECT_VIEWS,
			},
			{
				actionId: 'projectViews.switchToKanban',
				title: 'keyboardShortcuts.project.switchToKanbanView',
				keys: ['g', 'k'],
				combination: 'then',
				customizable: false,
				contexts: ['/projects/:id/*'],
				category: ShortcutCategory.PROJECT_VIEWS,
			},
		],
	},
	{
		title: 'keyboardShortcuts.gantt.title',
		category: ShortcutCategory.GANTT_VIEW,
		available: (route) => route.name === 'project.view',
		shortcuts: [
			{
				actionId: 'gantt.moveTaskLeft',
				title: 'keyboardShortcuts.gantt.moveTaskLeft',
				keys: ['←'],
				customizable: true,
				contexts: ['/projects/:id/gantt'],
				category: ShortcutCategory.GANTT_VIEW,
			},
			{
				actionId: 'gantt.moveTaskRight',
				title: 'keyboardShortcuts.gantt.moveTaskRight',
				keys: ['→'],
				customizable: true,
				contexts: ['/projects/:id/gantt'],
				category: ShortcutCategory.GANTT_VIEW,
			},
			{
				actionId: 'gantt.expandTaskLeft',
				title: 'keyboardShortcuts.gantt.expandTaskLeft',
				keys: ['shift', '←'],
				customizable: true,
				contexts: ['/projects/:id/gantt'],
				category: ShortcutCategory.GANTT_VIEW,
			},
			{
				actionId: 'gantt.expandTaskRight',
				title: 'keyboardShortcuts.gantt.expandTaskRight',
				keys: ['shift', '→'],
				customizable: true,
				contexts: ['/projects/:id/gantt'],
				category: ShortcutCategory.GANTT_VIEW,
			},
			{
				actionId: 'gantt.shrinkTaskLeft',
				title: 'keyboardShortcuts.gantt.shrinkTaskLeft',
				keys: [ctrl, '←'],
				customizable: true,
				contexts: ['/projects/:id/gantt'],
				category: ShortcutCategory.GANTT_VIEW,
			},
			{
				actionId: 'gantt.shrinkTaskRight',
				title: 'keyboardShortcuts.gantt.shrinkTaskRight',
				keys: [ctrl, '→'],
				customizable: true,
				contexts: ['/projects/:id/gantt'],
				category: ShortcutCategory.GANTT_VIEW,
			},
		],
	},
	{
		title: 'keyboardShortcuts.task.title',
		category: ShortcutCategory.TASK_ACTIONS,
		available: (route) => route.name === 'task.detail',
		shortcuts: [
			{
				actionId: 'task.markDone',
				title: 'keyboardShortcuts.task.done',
				keys: ['t'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.assign',
				title: 'keyboardShortcuts.task.assign',
				keys: ['a'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.labels',
				title: 'keyboardShortcuts.task.labels',
				keys: ['l'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.dueDate',
				title: 'keyboardShortcuts.task.dueDate',
				keys: ['d'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.attachment',
				title: 'keyboardShortcuts.task.attachment',
				keys: ['f'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.related',
				title: 'keyboardShortcuts.task.related',
				keys: ['r'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.move',
				title: 'keyboardShortcuts.task.move',
				keys: ['m'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.color',
				title: 'keyboardShortcuts.task.color',
				keys: ['c'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.reminder',
				title: 'keyboardShortcuts.task.reminder',
				keys: [reminderModifier, 'r'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.description',
				title: 'keyboardShortcuts.task.description',
				keys: ['e'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.priority',
				title: 'keyboardShortcuts.task.priority',
				keys: ['p'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.delete',
				title: 'keyboardShortcuts.task.delete',
				keys: ['shift', 'delete'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.toggleFavorite',
				title: 'keyboardShortcuts.task.favorite',
				keys: ['s'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.openProject',
				title: 'keyboardShortcuts.task.openProject',
				keys: ['u'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
			{
				actionId: 'task.save',
				title: 'keyboardShortcuts.task.save',
				keys: [ctrl, 's'],
				customizable: true,
				contexts: ['/tasks/:id'],
				category: ShortcutCategory.TASK_ACTIONS,
			},
		],
	},
] as const
