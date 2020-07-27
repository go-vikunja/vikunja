import Vue from 'vue'
import Router from 'vue-router'

import HomeComponent from '../views/Home'
import NotFoundComponent from '../views/404'
import LoadingComponent from '../components/misc/loading'
import ErrorComponent from '../components/misc/error'
// User Handling
import LoginComponent from '../views/user/Login'
import RegisterComponent from '../views/user/Register'

const PasswordResetComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "user-settings" */'../views/user/PasswordReset'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
const GetPasswordResetComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "user-settings" */'../views/user/RequestPasswordReset'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
const UserSettingsComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "user-settings" */'../views/user/Settings'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
// List Handling
const NewListComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "settings" */'../views/list/NewList'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
const EditListComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "settings" */'../views/list/EditList'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
// Tasks
import ShowTasksInRangeComponent from '../views/tasks/ShowTasksInRange'
import LinkShareAuthComponent from '../views/sharing/LinkSharingAuth'
import TaskDetailViewModal from '../views/tasks/TaskDetailViewModal'
import TaskDetailView from '../views/tasks/TaskDetailView'
// Namespace Handling
const NewNamespaceComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "settings" */'../views/namespaces/NewNamespace'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
const EditNamespaceComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "settings" */'../views/namespaces/EditNamespace'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
import ListNamespaces from '../views/namespaces/ListNamespaces'
// Team Handling
import ListTeamsComponent from '../views/teams/ListTeams'

const EditTeamComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "settings" */'../views/teams/EditTeam'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
const NewTeamComponent = () => ({
	component: import(/* webpackPrefetch: true *//* webpackChunkName: "settings" */'../views/teams/NewTeam'),
	loading: LoadingComponent,
	error: ErrorComponent,
	timeout: 60000,
})
// Label Handling
import ListLabelsComponent from '../views/labels/ListLabels'
// Migration
import MigrationComponent from '../views/migrator/Migrate'
import MigrateServiceComponent from '../views/migrator/MigrateService'
// List Views
import ShowListComponent from '../views/list/ShowList'
import Kanban from '../views/list/views/Kanban'
import List from '../views/list/views/List'
import Gantt from '../views/list/views/Gantt'
import Table from '../views/list/views/Table'

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
				selector: to.hash
			}
		}

		// Otherwise just scroll to the top
		return {x: 0, y: 0}
	},
	routes: [
		{
			path: '/',
			name: 'home',
			component: HomeComponent
		},
		{
			path: '*',
			name: '404',
			component: NotFoundComponent,
		},
		{
			path: '/login',
			name: 'user.login',
			component: LoginComponent
		},
		{
			path: '/get-password-reset',
			name: 'user.password-reset.request',
			component: GetPasswordResetComponent
		},
		{
			path: '/password-reset',
			name: 'user.password-reset.reset',
			component: PasswordResetComponent
		},
		{
			path: '/register',
			name: 'user.register',
			component: RegisterComponent
		},
		{
			path: '/user/settings',
			name: 'user.settings',
			component: UserSettingsComponent,
		},
		{
			path: '/share/:share/auth',
			name: 'link-share.auth',
			component: LinkShareAuthComponent
		},
		{
			path: '/namespaces',
			name: 'namespaces.index',
			component: ListNamespaces,
		},
		{
			path: '/namespaces/new',
			name: 'namespace.create',
			component: NewNamespaceComponent
		},
		{
			path: '/namespaces/:id/edit',
			name: 'namespace.edit',
			component: EditNamespaceComponent
		},
		{
			path: '/namespaces/:id/list',
			name: 'list.create',
			component: NewListComponent
		},
		{
			path: '/lists/:id/edit',
			name: 'list.edit',
			component: EditListComponent
		},
		{
			path: '/tasks/:id',
			name: 'task.detail',
			component: TaskDetailView,
		},
		{
			path: '/tasks/by/:type',
			name: 'tasks.range',
			component: ShowTasksInRangeComponent,
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
					],
				},
				{
					path: '/lists/:listId/table',
					name: 'list.table',
					component: Table,
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
					],
				},
			]
		},
		{
			path: '/teams',
			name: 'teams.index',
			component: ListTeamsComponent
		},
		{
			path: '/teams/new',
			name: 'teams.create',
			component: NewTeamComponent
		},
		{
			path: '/teams/:id/edit',
			name: 'teams.edit',
			component: EditTeamComponent
		},
		{
			path: '/labels',
			name: 'labels.index',
			component: ListLabelsComponent
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
	]
})