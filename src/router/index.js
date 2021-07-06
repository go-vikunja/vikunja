import Vue from 'vue'
import Router from 'vue-router'

import HomeComponent from '../views/Home'
import NotFoundComponent from '../views/404'
import LoadingComponent from '../components/misc/loading'
import ErrorComponent from '../components/misc/error'
import About from '../views/About'
// User Handling
import LoginComponent from '../views/user/Login'
import RegisterComponent from '../views/user/Register'
import OpenIdAuth from '@/views/user/OpenIdAuth'
// Tasks
import ShowTasksInRangeComponent from '../views/tasks/ShowTasksInRange'
import LinkShareAuthComponent from '../views/sharing/LinkSharingAuth'
import TaskDetailViewModal from '../views/tasks/TaskDetailViewModal'
import TaskDetailView from '../views/tasks/TaskDetailView'
import ListNamespaces from '../views/namespaces/ListNamespaces'
// Team Handling
import ListTeamsComponent from '../views/teams/ListTeams'
// Label Handling
import ListLabelsComponent from '../views/labels/ListLabels'
import NewLabelComponent from '../views/labels/NewLabel'
// Migration
import MigrationComponent from '../views/migrator/Migrate'
import MigrateServiceComponent from '../views/migrator/MigrateService'
// List Views
import ShowListComponent from '../views/list/ShowList'
import Kanban from '../views/list/views/Kanban'
import List from '../views/list/views/List'
import Gantt from '../views/list/views/Gantt'
import Table from '../views/list/views/Table'
// List Settings
import ListSettingEdit from '@/views/list/settings/edit'
import ListSettingBackground from '@/views/list/settings/background'
import ListSettingDuplicate from '@/views/list/settings/duplicate'
import ListSettingShare from '@/views/list/settings/share'
import ListSettingDelete from '@/views/list/settings/delete'
import ListSettingArchive from '@/views/list/settings/archive'
import FilterSettingEdit from '@/views/filters/settings/edit'
import FilterSettingDelete from '@/views/filters/settings/delete'
// Namespace Settings
import NamespaceSettingEdit from '@/views/namespaces/settings/edit'
import NamespaceSettingShare from '@/views/namespaces/settings/share'
import NamespaceSettingArchive from '@/views/namespaces/settings/archive'
import NamespaceSettingDelete from '@/views/namespaces/settings/delete'
// Saved Filters
import CreateSavedFilter from '@/views/filters/CreateSavedFilter'

const PasswordResetComponent = () => ({
	component: import(/* webpackChunkName: "user-settings" */'../views/user/PasswordReset'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
const GetPasswordResetComponent = () => ({
	component: import(/* webpackChunkName: "user-settings" */'../views/user/RequestPasswordReset'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
const UserSettingsComponent = () => ({
	component: import(/* webpackChunkName: "user-settings" */'../views/user/Settings'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
// List Handling
const NewListComponent = () => ({
	component: import(/* webpackChunkName: "settings" */'../views/list/NewList'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
// Namespace Handling
const NewNamespaceComponent = () => ({
	component: import(/* webpackChunkName: "settings" */'../views/namespaces/NewNamespace'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})

const EditTeamComponent = () => ({
	component: import(/* webpackChunkName: "settings" */'../views/teams/EditTeam'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
const NewTeamComponent = () => ({
	component: import(/* webpackChunkName: "settings" */'../views/teams/NewTeam'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})

Vue.use(Router)

export default new Router({
	mode: 'history',
	scrollBehavior(to, from, savedPosition) {
		// If the user is using their forward/backward keys to navigate, we want to restore the scroll view
		if (savedPosition) {
			return savedPosition
		}

		// Scroll to anchor should still work
		if (to.hash) {
			return {
				selector: to.hash,
			}
		}

		// Otherwise just scroll to the top
		return {x: 0, y: 0}
	},
	routes: [
		{
			path: '/',
			name: 'home',
			component: HomeComponent,
		},
		{
			path: '*',
			name: '404',
			component: NotFoundComponent,
		},
		{
			path: '/login',
			name: 'user.login',
			component: LoginComponent,
		},
		{
			path: '/get-password-reset',
			name: 'user.password-reset.request',
			component: GetPasswordResetComponent,
		},
		{
			path: '/password-reset',
			name: 'user.password-reset.reset',
			component: PasswordResetComponent,
		},
		{
			path: '/register',
			name: 'user.register',
			component: RegisterComponent,
		},
		{
			path: '/user/settings',
			name: 'user.settings',
			component: UserSettingsComponent,
		},
		{
			path: '/share/:share/auth',
			name: 'link-share.auth',
			component: LinkShareAuthComponent,
		},
		{
			path: '/namespaces',
			name: 'namespaces.index',
			component: ListNamespaces,
		},
		{
			path: '/namespaces/new',
			name: 'namespace.create',
			components: {
				popup: NewNamespaceComponent,
			},
		},
		{
			path: '/namespaces/:id/list',
			name: 'list.create',
			components: {
				popup: NewListComponent,
			}
		},
		{
			path: '/namespaces/:id/settings/edit',
			name: 'namespace.settings.edit',
			components: {
				popup: NamespaceSettingEdit,
			},
		},
		{
			path: '/namespaces/:id/settings/share',
			name: 'namespace.settings.share',
			components: {
				popup: NamespaceSettingShare,
			},
		},
		{
			path: '/namespaces/:id/settings/archive',
			name: 'namespace.settings.archive',
			components: {
				popup: NamespaceSettingArchive,
			},
		},
		{
			path: '/namespaces/:id/settings/delete',
			name: 'namespace.settings.delete',
			components: {
				popup: NamespaceSettingDelete,
			},
		},
		{
			path: '/tasks/:id',
			name: 'task.detail',
			component: TaskDetailView,
		},
		{
			path: '/tasks/by/upcoming',
			name: 'tasks.range',
			component: ShowTasksInRangeComponent,
		},
		{
			path: '/lists/:listId/settings/edit',
			name: 'list.settings.edit',
			components: {
				popup: ListSettingEdit,
			},
		},
		{
			path: '/lists/:listId/settings/background',
			name: 'list.settings.background',
			components: {
				popup: ListSettingBackground,
			},
		},
		{
			path: '/lists/:listId/settings/duplicate',
			name: 'list.settings.duplicate',
			components: {
				popup: ListSettingDuplicate,
			},
		},
		{
			path: '/lists/:listId/settings/share',
			name: 'list.settings.share',
			components: {
				popup: ListSettingShare,
			},
		},
		{
			path: '/lists/:listId/settings/delete',
			name: 'list.settings.delete',
			components: {
				popup: ListSettingDelete,
			},
		},
		{
			path: '/lists/:listId/settings/archive',
			name: 'list.settings.archive',
			components: {
				popup: ListSettingArchive,
			},
		},
		{
			path: '/lists/:listId/settings/edit',
			name: 'filter.settings.edit',
			components: {
				popup: FilterSettingEdit,
			},
		},
		{
			path: '/lists/:listId/settings/delete',
			name: 'filter.settings.delete',
			components: {
				popup: FilterSettingDelete,
			},
		},
		{
			path: '/lists/:listId',
			name: 'list.index',
			component: ShowListComponent,
			children: [
				{
					path: '/lists/:listId/list',
					name: 'list.list',
					component: List,
					children: [
						{
							path: '/tasks/:id',
							name: 'task.list.detail',
							component: TaskDetailViewModal,
						},
						{
							path: '/lists/:listId/settings/edit',
							name: 'list.list.settings.edit',
							component: ListSettingEdit,
						},
						{
							path: '/lists/:listId/settings/background',
							name: 'list.list.settings.background',
							component: ListSettingBackground,
						},
						{
							path: '/lists/:listId/settings/duplicate',
							name: 'list.list.settings.duplicate',
							component: ListSettingDuplicate,
						},
						{
							path: '/lists/:listId/settings/share',
							name: 'list.list.settings.share',
							component: ListSettingShare,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'list.list.settings.delete',
							component: ListSettingDelete,
						},
						{
							path: '/lists/:listId/settings/archive',
							name: 'list.list.settings.archive',
							component: ListSettingArchive,
						},
						{
							path: '/lists/:listId/settings/edit',
							name: 'filter.list.settings.edit',
							component: FilterSettingEdit,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'filter.list.settings.delete',
							component: FilterSettingDelete,
						},
					],
				},
				{
					path: '/lists/:listId/gantt',
					name: 'list.gantt',
					component: Gantt,
					children: [
						{
							path: '/tasks/:id',
							name: 'task.gantt.detail',
							component: TaskDetailViewModal,
						},
						{
							path: '/lists/:listId/settings/edit',
							name: 'list.gantt.settings.edit',
							component: ListSettingEdit,
						},
						{
							path: '/lists/:listId/settings/background',
							name: 'list.gantt.settings.background',
							component: ListSettingBackground,
						},
						{
							path: '/lists/:listId/settings/duplicate',
							name: 'list.gantt.settings.duplicate',
							component: ListSettingDuplicate,
						},
						{
							path: '/lists/:listId/settings/share',
							name: 'list.gantt.settings.share',
							component: ListSettingShare,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'list.gantt.settings.delete',
							component: ListSettingDelete,
						},
						{
							path: '/lists/:listId/settings/archive',
							name: 'list.gantt.settings.archive',
							component: ListSettingArchive,
						},
						{
							path: '/lists/:listId/settings/edit',
							name: 'filter.gantt.settings.edit',
							component: FilterSettingEdit,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'filter.gantt.settings.delete',
							component: FilterSettingDelete,
						},
					],
				},
				{
					path: '/lists/:listId/table',
					name: 'list.table',
					component: Table,
					children: [
						{
							path: '/lists/:listId/settings/edit',
							name: 'list.table.settings.edit',
							component: ListSettingEdit,
						},
						{
							path: '/lists/:listId/settings/background',
							name: 'list.table.settings.background',
							component: ListSettingBackground,
						},
						{
							path: '/lists/:listId/settings/duplicate',
							name: 'list.table.settings.duplicate',
							component: ListSettingDuplicate,
						},
						{
							path: '/lists/:listId/settings/share',
							name: 'list.table.settings.share',
							component: ListSettingShare,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'list.table.settings.delete',
							component: ListSettingDelete,
						},
						{
							path: '/lists/:listId/settings/archive',
							name: 'list.table.settings.archive',
							component: ListSettingArchive,
						},
						{
							path: '/lists/:listId/settings/edit',
							name: 'filter.table.settings.edit',
							component: FilterSettingEdit,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'filter.table.settings.delete',
							component: FilterSettingDelete,
						},
					],
				},
				{
					path: '/lists/:listId/kanban',
					name: 'list.kanban',
					component: Kanban,
					children: [
						{
							path: '/tasks/:id',
							name: 'task.kanban.detail',
							component: TaskDetailViewModal,
						},
						{
							path: '/lists/:listId/settings/edit',
							name: 'list.kanban.settings.edit',
							component: ListSettingEdit,
						},
						{
							path: '/lists/:listId/settings/background',
							name: 'list.kanban.settings.background',
							component: ListSettingBackground,
						},
						{
							path: '/lists/:listId/settings/duplicate',
							name: 'list.kanban.settings.duplicate',
							component: ListSettingDuplicate,
						},
						{
							path: '/lists/:listId/settings/share',
							name: 'list.kanban.settings.share',
							component: ListSettingShare,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'list.kanban.settings.delete',
							component: ListSettingDelete,
						},
						{
							path: '/lists/:listId/settings/archive',
							name: 'list.kanban.settings.archive',
							component: ListSettingArchive,
						},
						{
							path: '/lists/:listId/settings/edit',
							name: 'filter.kanban.settings.edit',
							component: FilterSettingEdit,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'filter.kanban.settings.delete',
							component: FilterSettingDelete,
						},
					],
				},
			],
		},
		{
			path: '/teams',
			name: 'teams.index',
			component: ListTeamsComponent,
		},
		{
			path: '/teams/new',
			name: 'teams.create',
			components: {
				popup: NewTeamComponent,
			},
		},
		{
			path: '/teams/:id/edit',
			name: 'teams.edit',
			component: EditTeamComponent,
		},
		{
			path: '/labels',
			name: 'labels.index',
			component: ListLabelsComponent,
		},
		{
			path: '/labels/new',
			name: 'labels.create',
			components: {
				popup: NewLabelComponent,
			},
		},
		{
			path: '/migrate',
			name: 'migrate.start',
			component: MigrationComponent,
		},
		{
			path: '/migrate/:service',
			name: 'migrate.service',
			component: MigrateServiceComponent,
		},
		{
			path: '/filters/new',
			name: 'filters.create',
			components: {
				popup: CreateSavedFilter,
			},
		},
		{
			path: '/auth/openid/:provider',
			name: 'openid.auth',
			component: OpenIdAuth,
		},
		{
			path: '/about',
			name: 'about',
			component: About,
		},
	],
})