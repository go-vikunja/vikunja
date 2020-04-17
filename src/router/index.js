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
import ShowListComponent from '@/components/lists/ShowList'
import NewListComponent from '@/components/lists/NewList'
import EditListComponent from '@/components/lists/EditList'
import ShowTasksInRangeComponent from '@/components/tasks/ShowTasksInRange'
import LinkShareAuthComponent from '@/components/sharing/linkSharingAuth'
import TaskDetailViewComponent from '@/components/tasks/TaskDetailView'
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
import WunderlistMigrationComponent from '../components/migrator/wunderlist'

Vue.use(Router)

export default new Router({
	mode: 'history',
	scrollBehavior (to, from, savedPosition) {
		// If the user is using their forward/backward keys to navigate, we want to restore the scroll view
		if(savedPosition) {
			return savedPosition
		}

		// Scroll to anchor should still work
		if(to.hash) {
			return {
				selector: to.hash
			}
		}

		// Otherwise just scroll to the top
		return { x: 0, y: 0 }
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
			path: '/lists/:id',
			name: 'showList',
			component: ShowListComponent
		},
		{
			path: '/lists/:id/edit',
			name: 'editList',
			component: EditListComponent
		},
		{
			path: '/lists/:id/:type',
			name: 'showListWithType',
			component: ShowListComponent,
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
			component: ShowTasksInRangeComponent
		},
		{
			path: '/tasks/:id',
			name: 'taskDetailView',
			component: TaskDetailViewComponent,
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
			path: '/migrate/wunderlist',
			name: 'migrateWunderlist',
			component: WunderlistMigrationComponent,
		},
		{
			path: '/user/settings',
			name: 'userSettings',
			component: UserSettingsComponent,
		},
	]
})