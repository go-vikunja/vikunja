import Vue from 'vue'
import Router from 'vue-router'

import HomeComponent from '@/components/Home'
import NotFoundComponent from '@/components/404'
// User Handling
import LoginComponent from '@/components/user/Login'
import RegisterComponent from '@/components/user/Register'
import PasswordResetComponent from '@/components/user/PasswordReset'
import GetPasswordResetComponent from '@/components/user/RequestPasswordReset'
import UserSettingsComponent from '@/components/user/Settings'
// List Handling
import NewListComponent from '@/components/lists/NewList'
import EditListComponent from '@/components/lists/EditList'
import ShowTasksInRangeComponent from '@/components/tasks/ShowTasksInRange'
import LinkShareAuthComponent from '../components/sharing/linkSharingAuth'
import TaskDetailViewModal from '../components/tasks/TaskDetailViewModal'
import TaskDetailView from '../components/tasks/TaskDetailView'
// Namespace Handling
import NewNamespaceComponent from '@/components/namespaces/NewNamespace'
import EditNamespaceComponent from '@/components/namespaces/EditNamespace'
// Team Handling
import ListTeamsComponent from '@/components/teams/ListTeams'
import EditTeamComponent from '@/components/teams/EditTeam'
import NewTeamComponent from '@/components/teams/NewTeam'
// Label Handling
import ListLabelsComponent from '@/components/labels/ListLabels'
// Migration
import MigrationComponent from '../components/migrator/migrate'
import MigrateServiceComponent from '../components/migrator/migrate-service'
// List Views
import ShowListComponent from '../components/lists/ShowList'
import Kanban from '../components/lists/views/Kanban'
import List from '../components/lists/views/List'
import Gantt from '../components/lists/views/Gantt'
import Table from '../components/lists/views/Table'

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
			name: 'login',
			component: LoginComponent
		},
		{
			path: '/get-password-reset',
			name: 'getPasswordReset',
			component: GetPasswordResetComponent
		},
		{
			path: '/password-reset',
			name: 'passwordReset',
			component: PasswordResetComponent
		},
		{
			path: '/register',
			name: 'register',
			component: RegisterComponent
		},
		{
			path: '/lists/:id/edit',
			name: 'editList',
			component: EditListComponent
		},
		{
			path: '/tasks/:id',
			name: 'task.detail',
			component: TaskDetailView,
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
			path: '/namespaces/:id/list',
			name: 'newList',
			component: NewListComponent
		},
		{
			path: '/namespaces/new',
			name: 'newNamespace',
			component: NewNamespaceComponent
		},
		{
			path: '/namespaces/:id/edit',
			name: 'editNamespace',
			component: EditNamespaceComponent
		},
		{
			path: '/teams',
			name: 'listTeams',
			component: ListTeamsComponent
		},
		{
			path: '/teams/new',
			name: 'newTeam',
			component: NewTeamComponent
		},
		{
			path: '/teams/:id/edit',
			name: 'editTeam',
			component: EditTeamComponent
		},
		{
			path: '/tasks/by/:type',
			name: 'showTasksInRange',
			component: ShowTasksInRangeComponent,
		},
		{
			path: '/labels',
			name: 'listLabels',
			component: ListLabelsComponent
		},
		{
			path: '/share/:share/auth',
			name: 'linkShareAuth',
			component: LinkShareAuthComponent
		},
		{
			path: '/migrate',
			name: 'migrateStart',
			component: MigrationComponent,
		},
		{
			path: '/migrate/:service',
			name: 'migrate',
			component: MigrateServiceComponent,
		},
		{
			path: '/user/settings',
			name: 'userSettings',
			component: UserSettingsComponent,
		},
	]
})