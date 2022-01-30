import { createRouter, createWebHistory, RouteLocation } from 'vue-router'
import {saveLastVisited} from '@/helpers/saveLastVisited'
import {store} from '@/store'

import {saveListView, getListView} from '@/helpers/saveListView'

import HomeComponent from '../views/Home.vue'
import NotFoundComponent from '../views/404.vue'
import About from '../views/About.vue'
// User Handling
import LoginComponent from '../views/user/Login.vue'
import RegisterComponent from '../views/user/Register.vue'
import OpenIdAuth from '../views/user/OpenIdAuth.vue'
import DataExportDownload from '../views/user/DataExportDownload.vue'
// Tasks
import ShowTasksInRangeComponent from '../views/tasks/ShowTasksInRange.vue'
import LinkShareAuthComponent from '../views/sharing/LinkSharingAuth.vue'
import ListNamespaces from '../views/namespaces/ListNamespaces.vue'
import TaskDetailViewModal from '../views/tasks/TaskDetailViewModal.vue'
// Team Handling
import ListTeamsComponent from '../views/teams/ListTeams.vue'
// Label Handling
import ListLabelsComponent from '../views/labels/ListLabels.vue'
import NewLabelComponent from '../views/labels/NewLabel.vue'
// Migration
import MigrationComponent from '../views/migrator/Migrate.vue'
import MigrateServiceComponent from '../views/migrator/MigrateService.vue'
// List Views
import ListList from '../views/list/ListList.vue'
import ListGantt from '../views/list/ListGantt.vue'
import ListTable from '../views/list/ListTable.vue'
import ListKanban from '../views/list/ListKanban.vue'

// List Settings
import ListSettingEdit from '../views/list/settings/edit.vue'
import ListSettingBackground from '../views/list/settings/background.vue'
import ListSettingDuplicate from '../views/list/settings/duplicate.vue'
import ListSettingShare from '../views/list/settings/share.vue'
import ListSettingDelete from '../views/list/settings/delete.vue'
import ListSettingArchive from '../views/list/settings/archive.vue'

// Namespace Settings
import NamespaceSettingEdit from '../views/namespaces/settings/edit.vue'
import NamespaceSettingShare from '../views/namespaces/settings/share.vue'
import NamespaceSettingArchive from '../views/namespaces/settings/archive.vue'
import NamespaceSettingDelete from '../views/namespaces/settings/delete.vue'

// Saved Filters
import FilterNew from '@/views/filters/FilterNew.vue'
import FilterEdit from '@/views/filters/FilterEdit.vue'
import FilterDelete from '@/views/filters/FilterDelete.vue'

const PasswordResetComponent = () => import('../views/user/PasswordReset.vue')
const GetPasswordResetComponent = () => import('../views/user/RequestPasswordReset.vue')
const UserSettingsComponent = () => import('../views/user/Settings.vue')
const UserSettingsAvatarComponent = () => import('../views/user/settings/Avatar.vue')
const UserSettingsCaldavComponent = () => import('../views/user/settings/Caldav.vue')
const UserSettingsDataExportComponent = () => import('../views/user/settings/DataExport.vue')
const UserSettingsDeletionComponent = () => import('../views/user/settings/Deletion.vue')
const UserSettingsEmailUpdateComponent = () => import('../views/user/settings/EmailUpdate.vue')
const UserSettingsGeneralComponent = () => import('../views/user/settings/General.vue')
const UserSettingsPasswordUpdateComponent = () => import('../views/user/settings/PasswordUpdate.vue')
const UserSettingsTOTPComponent = () => import('../views/user/settings/TOTP.vue')

// List Handling
const NewListComponent = () => import('../views/list/NewList.vue')

// Namespace Handling
const NewNamespaceComponent = () => import('../views/namespaces/NewNamespace.vue')

const EditTeamComponent = () => import('../views/teams/EditTeam.vue')
const NewTeamComponent = () =>  import('../views/teams/NewTeam.vue')

const router = createRouter({
	history: createWebHistory(),
	scrollBehavior(to, from, savedPosition) {
		// If the user is using their forward/backward keys to navigate, we want to restore the scroll view
		if (savedPosition) {
			return savedPosition
		}

		// Scroll to anchor should still work
		if (to.hash) {
			return {el: document.getElementById(to.hash.slice(1))}
		}

		// Otherwise just scroll to the top
		return {left: 0, top: 0}
	},
	routes: [
		{
			path: '/',
			name: 'home',
			component: HomeComponent,
		},
		{
			path: '/:pathMatch(.*)*',
			name: 'not-found',
			component: NotFoundComponent,
		},
		// if you omit the last `*`, the `/` character in params will be encoded when resolving or pushing
		{
			path: '/:pathMatch(.*)',
			name: 'bad-not-found',
			component: NotFoundComponent,
		},
		{
			path: '/login',
			name: 'user.login',
			component: LoginComponent,
			meta: {
				title: 'user.auth.login',
			},
		},
		{
			path: '/get-password-reset',
			name: 'user.password-reset.request',
			component: GetPasswordResetComponent,
			meta: {
				title: 'user.auth.resetPassword',
			},
		},
		{
			path: '/password-reset',
			name: 'user.password-reset.reset',
			component: PasswordResetComponent,
			meta: {
				title: 'user.auth.resetPassword',
			},
		},
		{
			path: '/register',
			name: 'user.register',
			component: RegisterComponent,
			meta: {
				title: 'user.auth.register',
			},
		},
		{
			path: '/user/settings',
			name: 'user.settings',
			component: UserSettingsComponent,
			redirect: {name: 'user.settings.general'},
			children: [
				{
					path: '/user/settings/avatar',
					name: 'user.settings.avatar',
					component: UserSettingsAvatarComponent,
				},
				{
					path: '/user/settings/caldav',
					name: 'user.settings.caldav',
					component: UserSettingsCaldavComponent,
				},
				{
					path: '/user/settings/data-export',
					name: 'user.settings.data-export',
					component: UserSettingsDataExportComponent,
				},
				{
					path: '/user/settings/deletion',
					name: 'user.settings.deletion',
					component: UserSettingsDeletionComponent,
				},
				{
					path: '/user/settings/email-update',
					name: 'user.settings.email-update',
					component: UserSettingsEmailUpdateComponent,
				},
				{
					path: '/user/settings/general',
					name: 'user.settings.general',
					component: UserSettingsGeneralComponent,
				},
				{
					path: '/user/settings/password-update',
					name: 'user.settings.password-update',
					component: UserSettingsPasswordUpdateComponent,
				},
				{
					path: '/user/settings/totp',
					name: 'user.settings.totp',
					component: UserSettingsTOTPComponent,
				},
			],
		},
		{
			path: '/user/export/download',
			name: 'user.export.download',
			component: DataExportDownload,
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
			component: NewNamespaceComponent,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/namespaces/:id/settings/edit',
			name: 'namespace.settings.edit',
			component: NamespaceSettingEdit,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/namespaces/:namespaceId/settings/share',
			name: 'namespace.settings.share',
			component: NamespaceSettingShare,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/namespaces/:id/settings/archive',
			name: 'namespace.settings.archive',
			component: NamespaceSettingArchive,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/namespaces/:id/settings/delete',
			name: 'namespace.settings.delete',
			component: NamespaceSettingDelete,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/tasks/:id',
			name: 'task.detail',
			component: TaskDetailViewModal,
			props: route => ({ taskId: parseInt(route.params.id as string) }),
		},
		{
			path: '/tasks/by/upcoming',
			name: 'tasks.range',
			component: ShowTasksInRangeComponent,
		},
		{
			path: '/lists/new/:namespaceId/',
			name: 'list.create',
			component: NewListComponent,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId/settings/edit',
			name: 'list.settings.edit',
			component: ListSettingEdit,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId/settings/background',
			name: 'list.settings.background',
			component: ListSettingBackground,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId/settings/duplicate',
			name: 'list.settings.duplicate',
			component: ListSettingDuplicate,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId/settings/share',
			name: 'list.settings.share',
			component: ListSettingShare,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId/settings/delete',
			name: 'list.settings.delete',
			component: ListSettingDelete,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId/settings/archive',
			name: 'list.settings.archive',
			component: ListSettingArchive,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId/settings/edit',
			name: 'filter.settings.edit',
			component: FilterEdit,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId/settings/delete',
			name: 'filter.settings.delete',
			component: FilterDelete,
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/lists/:listId',
			name: 'list.index',
			redirect(to) {
				// Redirect the user to list view by default

				const savedListView = getListView(to.params.listId)
				console.debug('Replaced list view with', savedListView)

				return {
					name:  router.hasRoute(savedListView)
						? savedListView
						: 'list.list',
					params: {listId: to.params.listId},
				}
			},
		},
		{
			path: '/lists/:listId/list',
			name: 'list.list',
			component: ListList,
			beforeEnter: (to) => saveListView(to.params.listId, to.name),
			props: route => ({ listId: parseInt(route.params.listId as string) }),
		},
		{
			path: '/lists/:listId/gantt',
			name: 'list.gantt',
			component: ListGantt,
			beforeEnter: (to) => saveListView(to.params.listId, to.name),
			props: route => ({ listId: parseInt(route.params.listId as string) }),
		},
		{
			path: '/lists/:listId/table',
			name: 'list.table',
			component: ListTable,
			beforeEnter: (to) => saveListView(to.params.listId, to.name),
			props: route => ({ listId: parseInt(route.params.listId as string) }),
		},
		{
			path: '/lists/:listId/kanban',
			name: 'list.kanban',
			component: ListKanban,
			beforeEnter: (to) => saveListView(to.params.listId, to.name),
			props: route => ({ listId: parseInt(route.params.listId as string) }),
		},
		{
			path: '/teams',
			name: 'teams.index',
			component: ListTeamsComponent,
		},
		{
			path: '/teams/new',
			name: 'teams.create',
			component: NewTeamComponent,
			meta: {
				showAsModal: true,
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
			component: NewLabelComponent,
			meta: {
				showAsModal: true,
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
			component: FilterNew,
			meta: {
				showAsModal: true,
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

router.beforeEach((to) => {
	return checkAuth(to)
})

function checkAuth(route: RouteLocation) {
	const authUser = store.getters['auth/authUser']
	const authLinkShare = store.getters['auth/authLinkShare']

	if (authUser || authLinkShare) {
		return
	}

	// Check if the user is already logged in and redirect them to the home page if not
	if (
		![
			'user.login',
			'user.password-reset.request',
			'user.password-reset.reset',
			'user.register',
			'link-share.auth',
			'openid.auth',
		].includes(route.name as string) &&
		localStorage.getItem('passwordResetToken') === null &&
		localStorage.getItem('emailConfirmToken') === null
	) {
		saveLastVisited(route.name as string, route.params)
		return {name: 'user.login'}
	}
}

export default router