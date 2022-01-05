import { createRouter, createWebHistory, RouteLocation } from 'vue-router'
import {saveLastVisited} from '@/helpers/saveLastVisited'
import {store} from '@/store'

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
import TaskDetailViewModal from '../views/tasks/TaskDetailViewModal.vue'
import TaskDetailView from '../views/tasks/TaskDetailView.vue'
import ListNamespaces from '../views/namespaces/ListNamespaces.vue'
// Team Handling
import ListTeamsComponent from '../views/teams/ListTeams.vue'
// Label Handling
import ListLabelsComponent from '../views/labels/ListLabels.vue'
import NewLabelComponent from '../views/labels/NewLabel.vue'
// Migration
import MigrationComponent from '../views/migrator/Migrate.vue'
import MigrateServiceComponent from '../views/migrator/MigrateService.vue'
// List Views
import ShowListComponent from '../views/list/ShowList.vue'
import Kanban from '../views/list/views/Kanban.vue'
import List from '../views/list/views/List.vue'
import Gantt from '../views/list/views/Gantt.vue'
import Table from '../views/list/views/Table.vue'
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
			components: {
				popup: NewNamespaceComponent,
			},
		},
		{
			path: '/namespaces/:id/list',
			name: 'list.create',
			components: {
				popup: NewListComponent,
			},
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
				popup: FilterEdit,
			},
		},
		{
			path: '/lists/:listId/settings/delete',
			name: 'filter.settings.delete',
			components: {
				popup: FilterDelete,
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
							component: FilterEdit,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'filter.list.settings.delete',
							component: FilterDelete,
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
							component: FilterEdit,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'filter.gantt.settings.delete',
							component: FilterDelete,
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
							component: FilterEdit,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'filter.table.settings.delete',
							component: FilterDelete,
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
							component: FilterEdit,
						},
						{
							path: '/lists/:listId/settings/delete',
							name: 'filter.kanban.settings.delete',
							component: FilterDelete,
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
				popup: FilterNew,
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