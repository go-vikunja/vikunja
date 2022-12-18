import { createRouter, createWebHistory } from 'vue-router'
import type { RouteLocation } from 'vue-router'
import {saveLastVisited} from '@/helpers/saveLastVisited'

import {saveListView, getListView} from '@/helpers/saveListView'
import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {getNextWeekDate} from '@/helpers/time/getNextWeekDate'
import {setTitle} from '@/helpers/setTitle'

import {useListStore} from '@/stores/lists'
import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'

import HomeComponent from '@/views/Home.vue'
import NotFoundComponent from '@/views/404.vue'
const About = () => import('@/views/About.vue')
// User Handling
import LoginComponent from '@/views/user/Login.vue'
import RegisterComponent from '@/views/user/Register.vue'
import OpenIdAuth from '@/views/user/OpenIdAuth.vue'
const DataExportDownload = () => import('@/views/user/DataExportDownload.vue')
// Tasks
import UpcomingTasksComponent from '@/views/tasks/ShowTasks.vue'
import LinkShareAuthComponent from '@/views/sharing/LinkSharingAuth.vue'
const ListNamespaces = () => import('@/views/namespaces/ListNamespaces.vue')
const TaskDetailView = () => import('@/views/tasks/TaskDetailView.vue')

// Team Handling
const ListTeamsComponent = () => import('@/views/teams/ListTeams.vue')
// Label Handling
const ListLabelsComponent = () => import('@/views/labels/ListLabels.vue')
const NewLabelComponent = () => import('@/views/labels/NewLabel.vue')
// Migration
const MigrationComponent = () => import('@/views/migrate/Migration.vue')
const MigrationHandlerComponent = () => import('@/views/migrate/MigrationHandler.vue')
// List Views
const ListList = () => import('@/views/list/ListList.vue')
const ListGantt = () => import('@/views/list/ListGantt.vue')
const ListTable = () => import('@/views/list/ListTable.vue')
const ListKanban = () => import('@/views/list/ListKanban.vue')
const ListInfo = () => import('@/views/list/ListInfo.vue')

// List Settings
const ListSettingEdit = () => import('@/views/list/settings/edit.vue')
const ListSettingBackground = () => import('@/views/list/settings/background.vue')
const ListSettingDuplicate = () => import('@/views/list/settings/duplicate.vue')
const ListSettingShare = () => import('@/views/list/settings/share.vue')
const ListSettingDelete = () => import('@/views/list/settings/delete.vue')
const ListSettingArchive = () => import('@/views/list/settings/archive.vue')

// Namespace Settings
const NamespaceSettingEdit = () => import('@/views/namespaces/settings/edit.vue')
const NamespaceSettingShare = () => import('@/views/namespaces/settings/share.vue')
const NamespaceSettingArchive = () => import('@/views/namespaces/settings/archive.vue')
const NamespaceSettingDelete = () => import('@/views/namespaces/settings/delete.vue')

// Saved Filters
const FilterNew = () => import('@/views/filters/FilterNew.vue')
const FilterEdit = () => import('@/views/filters/FilterEdit.vue')
const FilterDelete = () => import('@/views/filters/FilterDelete.vue')

const PasswordResetComponent = () => import('@/views/user/PasswordReset.vue')
const GetPasswordResetComponent = () => import('@/views/user/RequestPasswordReset.vue')
const UserSettingsComponent = () => import('@/views/user/Settings.vue')
const UserSettingsAvatarComponent = () => import('@/views/user/settings/Avatar.vue')
const UserSettingsCaldavComponent = () => import('@/views/user/settings/Caldav.vue')
const UserSettingsDataExportComponent = () => import('@/views/user/settings/DataExport.vue')
const UserSettingsDeletionComponent = () => import('@/views/user/settings/Deletion.vue')
const UserSettingsEmailUpdateComponent = () => import('@/views/user/settings/EmailUpdate.vue')
const UserSettingsGeneralComponent = () => import('@/views/user/settings/General.vue')
const UserSettingsPasswordUpdateComponent = () => import('@/views/user/settings/PasswordUpdate.vue')
const UserSettingsTOTPComponent = () => import('@/views/user/settings/TOTP.vue')

// List Handling
const NewListComponent = () => import('@/views/list/NewList.vue')

// Namespace Handling
const NewNamespaceComponent = () => import('@/views/namespaces/NewNamespace.vue')

const EditTeamComponent = () => import('@/views/teams/EditTeam.vue')
const NewTeamComponent = () =>  import('@/views/teams/NewTeam.vue')

const router = createRouter({
	history: createWebHistory(),
	scrollBehavior(to, from, savedPosition) {
		// If the user is using their forward/backward keys to navigate, we want to restore the scroll view
		if (savedPosition) {
			return savedPosition
		}

		// Scroll to anchor should still work
		if (to.hash) {
			return {el: to.hash}
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
				title: 'user.auth.createAccount',
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
			props: route => ({ namespaceId: Number(route.params.id as string) }),
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
			props: route => ({ namespaceId: parseInt(route.params.id as string) }),
		},
		{
			path: '/namespaces/:id/settings/delete',
			name: 'namespace.settings.delete',
			component: NamespaceSettingDelete,
			meta: {
				showAsModal: true,
			},
			props: route => ({ namespaceId: Number(route.params.id as string) }),
		},
		{
			path: '/tasks/:id',
			name: 'task.detail',
			component: TaskDetailView,
			props: route => ({ taskId: Number(route.params.id as string) }),
		},
		{
			path: '/tasks/by/upcoming',
			name: 'tasks.range',
			component: UpcomingTasksComponent,
			props: route => ({
				dateFrom: parseDateOrString(route.query.from as string, new Date()),
				dateTo: parseDateOrString(route.query.to as string, getNextWeekDate()),
				showNulls: route.query.showNulls === 'true',
				showOverdue: route.query.showOverdue === 'true',
			}),
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
			props: route => ({ listId: Number(route.params.listId as string) }),
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
			props: route => ({ listId: Number(route.params.listId as string) }),
		},
		{
			path: '/lists/:listId/settings/delete',
			name: 'filter.settings.delete',
			component: FilterDelete,
			meta: {
				showAsModal: true,
			},
			props: route => ({ listId: Number(route.params.listId as string) }),
		},
		{
			path: '/lists/:listId/info',
			name: 'list.info',
			component: ListInfo,
			meta: {
				showAsModal: true,
			},
			props: route => ({ listId: Number(route.params.listId as string) }),
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
			props: route => ({ listId: Number(route.params.listId as string) }),
		},
		{
			path: '/lists/:listId/gantt',
			name: 'list.gantt',
			component: ListGantt,
			beforeEnter: (to) => saveListView(to.params.listId, to.name),
			// FIXME: test if `useRoute` would be the same. If it would use it instead.
			props: route => ({route}),
		},
		{
			path: '/lists/:listId/table',
			name: 'list.table',
			component: ListTable,
			beforeEnter: (to) => saveListView(to.params.listId, to.name),
			props: route => ({ listId: Number(route.params.listId as string) }),
		},
		{
			path: '/lists/:listId/kanban',
			name: 'list.kanban',
			component: ListKanban,
			beforeEnter: (to) => {
				saveListView(to.params.listId, to.name)
				// Properly set the page title when a task popup is closed
				const listStore = useListStore()
				const listFromStore = listStore.getListById(Number(to.params.listId))
				if(listFromStore) {
					setTitle(listFromStore.title)
				}
			},
			props: route => ({ listId: Number(route.params.listId as string) }),
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
			component: MigrationHandlerComponent,
			props: route => ({
				service: route.params.service as string,
				code: route.query.code as string,
			}),
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

export async function getAuthForRoute(route: RouteLocation) {
	const authStore = useAuthStore()
	if (authStore.authUser || authStore.authLinkShare) {
		return
	}
	
	const baseStore = useBaseStore()
	// When trying this before the current user was fully loaded we might get a flash of the login screen 
	// in the user shell. To make shure this does not happen we check if everything is ready before trying.
	if (!baseStore.ready) {
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
		localStorage.getItem('emailConfirmToken') === null &&
		!(route.name === 'home' && (typeof route.query.userPasswordReset !== 'undefined' || typeof route.query.userEmailConfirm !== 'undefined'))
	) {
		saveLastVisited(route.name as string, route.params, route.query)
		return {name: 'user.login'}
	}
	
	if(localStorage.getItem('passwordResetToken') !== null && route.name !== 'user.password-reset.reset') {
		return {name: 'user.password-reset.reset'}
	}
	
	if(localStorage.getItem('emailConfirmToken') !== null && route.name !== 'user.login') {
		return {name: 'user.login'}
	}
}

router.beforeEach(async (to) => {
	return getAuthForRoute(to)
})

export default router