import { createRouter, createWebHistory } from 'vue-router'
import type { RouteLocation } from 'vue-router'
import {saveLastVisited} from '@/helpers/saveLastVisited'

import {getProjectViewId} from '@/helpers/projectView'
import {parseDateOrString} from '@/helpers/time/parseDateOrString'
import {getNextWeekDate} from '@/helpers/time/getNextWeekDate'
import {LINK_SHARE_HASH_PREFIX} from '@/constants/linkShareHash'

import {useAuthStore} from '@/stores/auth'

import Login from '@/views/user/Login.vue'
import Register from '@/views/user/Register.vue'
import LinkSharingAuth from '@/views/sharing/LinkSharingAuth.vue'
import OpenIdAuth from '@/views/user/OpenIdAuth.vue'
import UpcomingTasks from '@/views/tasks/ShowTasks.vue'

import NotFoundComponent from '@/views/404.vue'

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	scrollBehavior(to, from, savedPosition) {
		// If the user is using their forward/backward keys to navigate, we want to restore the scroll view
		if (savedPosition) {
			return savedPosition
		}

		// Scroll to anchor should still work
		if (to.hash && !to.hash.startsWith(LINK_SHARE_HASH_PREFIX)) {
			return {el: to.hash}
		}

		// Otherwise just scroll to the top
		return {
			'inset-inline-start': 0,
			'inset-block-start': 0,
		}
	},
	routes: [
		{
			path: '/',
			name: 'home',
			component: () => import('@/views/Home.vue'),
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
			component: Login,
			meta: {
				title: 'user.auth.login',
			},
		},
		{
			path: '/get-password-reset',
			name: 'user.password-reset.request',
			component: () => import('@/views/user/RequestPasswordReset.vue'),
			meta: {
				title: 'user.auth.resetPassword',
			},
		},
		{
			path: '/password-reset',
			name: 'user.password-reset.reset',
			component: () => import('@/views/user/PasswordReset.vue'),
			meta: {
				title: 'user.auth.resetPassword',
			},
		},
		{
			path: '/register',
			name: 'user.register',
			// FIXME: use dynamic imports
			// component: () => import('@/views/user/Register.vue'),
			component: Register,
			meta: {
				title: 'user.auth.createAccount',
			},
		},
		{
			path: '/user/settings',
			name: 'user.settings',
			component: () => import('@/views/user/Settings.vue'),
			redirect: {name: 'user.settings.general'},
			children: [
				{
					path: '/user/settings/avatar',
					name: 'user.settings.avatar',
					component: () => import('@/views/user/settings/Avatar.vue'),
				},
				{
					path: '/user/settings/caldav',
					name: 'user.settings.caldav',
					component: () => import('@/views/user/settings/Caldav.vue'),
				},
				{
					path: '/user/settings/data-export',
					name: 'user.settings.data-export',
					component: () => import('@/views/user/settings/DataExport.vue'),
				},
				{
					path: '/user/settings/deletion',
					name: 'user.settings.deletion',
					component: () => import('@/views/user/settings/Deletion.vue'),
				},
				{
					path: '/user/settings/email-update',
					name: 'user.settings.email-update',
					component: () => import('@/views/user/settings/EmailUpdate.vue'),
				},
				{
					path: '/user/settings/general',
					name: 'user.settings.general',
					component: () => import('@/views/user/settings/General.vue'),
				},
				{
					path: '/user/settings/password-update',
					name: 'user.settings.password-update',
					component: () => import('@/views/user/settings/PasswordUpdate.vue'),
				},
				{
					path: '/user/settings/totp',
					name: 'user.settings.totp',
					component: () => import('@/views/user/settings/TOTP.vue'),
				},
				{
					path: '/user/settings/api-tokens',
					name: 'user.settings.apiTokens',
					component: () => import('@/views/user/settings/ApiTokens.vue'),
				},
				{
					path: '/user/settings/migrate',
					name: 'migrate.start',
					component: () => import('@/views/migrate/Migration.vue'),
				},
				{
					path: '/migrate/:service',
					name: 'migrate.service',
					component: () => import('@/views/migrate/MigrationHandler.vue'),
					props: route => ({
						service: route.params.service as string,
						code: route.query.code as string,
					}),
				},
			],
		},
		{
			path: '/user/export/download',
			name: 'user.export.download',
			component: () => import('@/views/user/DataExportDownload.vue'),
		},
		{
			path: '/share/:share/auth',
			name: 'link-share.auth',
			// FIXME: use dynamic imports
			// component: () => import('@/views/sharing/LinkSharingAuth.vue'),
			component: LinkSharingAuth,
		},
		{
			path: '/tasks/:id',
			name: 'task.detail',
			component: () => import('@/views/tasks/TaskDetailView.vue'),
			props: route => ({ taskId: Number(route.params.id as string) }),
		},
		{
			path: '/tasks/by/upcoming',
			name: 'tasks.range',
			component: UpcomingTasks,
			props: route => ({
				dateFrom: parseDateOrString(route.query.from as string, new Date()),
				dateTo: parseDateOrString(route.query.to as string, getNextWeekDate()),
				showNulls: route.query.showNulls === 'true',
				showOverdue: route.query.showOverdue === 'true',
			}),
		},
		{
			// Redirect old list routes to the respective project routes
			// see: https://router.vuejs.org/guide/essentials/dynamic-matching.html#catch-all-404-not-found-route
			path: '/lists:pathMatch(.*)*',
			name: 'lists',
			redirect(to) {
				return {
					path: to.path.replace('/lists', '/projects'),
					query: to.query,
					hash: to.hash,
				}
			},
		},
		{
			path: '/projects',
			name: 'projects.index',
			component: () => import('@/views/project/ListProjects.vue'),
		},
		{
			path: '/projects/new',
			name: 'project.create',
			component: () => import('@/views/project/NewProject.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:parentProjectId/new',
			name: 'project.createFromParent',
			component: () => import('@/views/project/NewProject.vue'),
			props: route => ({ parentProjectId: Number(route.params.parentProjectId as string) }),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:projectId/settings/edit',
			name: 'project.settings.edit',
			component: () => import('@/views/project/settings/ProjectSettingsEdit.vue'),
			props: route => ({ projectId: Number(route.params.projectId as string) }),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:projectId/settings/background',
			name: 'project.settings.background',
			component: () => import('@/views/project/settings/ProjectSettingsBackground.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:projectId/settings/duplicate',
			name: 'project.settings.duplicate',
			component: () => import('@/views/project/settings/ProjectSettingsDuplicate.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:projectId/settings/share',
			name: 'project.settings.share',
			component: () => import('@/views/project/settings/ProjectSettingsShare.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:projectId/settings/webhooks',
			name: 'project.settings.webhooks',
			component: () => import('@/views/project/settings/ProjectSettingsWebhooks.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:projectId/settings/delete',
			name: 'project.settings.delete',
			component: () => import('@/views/project/settings/ProjectSettingsDelete.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:projectId/settings/archive',
			name: 'project.settings.archive',
			component: () => import('@/views/project/settings/ProjectSettingsArchive.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/projects/:projectId/settings/views',
			name: 'project.settings.views',
			component: () =>  import('@/views/project/settings/ProjectSettingsViews.vue'),
			meta: {
				showAsModal: true,
			},
			props: route => ({ projectId: Number(route.params.projectId as string) }),
		},
		{
			path: '/projects/:projectId/settings/edit',
			name: 'filter.settings.edit',
			component: () => import('@/views/filters/FilterEdit.vue'),
			meta: {
				showAsModal: true,
			},
			props: route => ({ projectId: Number(route.params.projectId as string) }),
		},
		{
			path: '/projects/:projectId/settings/delete',
			name: 'filter.settings.delete',
			component: () => import('@/views/filters/FilterDelete.vue'),
			meta: {
				showAsModal: true,
			},
			props: route => ({ projectId: Number(route.params.projectId as string) }),
		},
		{
			path: '/projects/:projectId/info',
			name: 'project.info',
			component: () => import('@/views/project/ProjectInfo.vue')			,
			meta: {
				showAsModal: true,
			},
			props: route => ({ projectId: Number(route.params.projectId as string) }),
		},
		{
			path: '/projects/:projectId',
			name: 'project.index',
			redirect(to) {
				const viewId = getProjectViewId(Number(to.params.projectId as string))

				if (viewId) {
					console.debug('Replaced list view with', viewId)
				}

				return {
					name: 'project.view',
					params: {
						projectId: parseInt(to.params.projectId as string),
						viewId: viewId ?? 0,
					},
				}
			},
		},
		{
			path: '/projects/:projectId/:viewId',
			name: 'project.view',
			component: () => import('@/views/project/ProjectView.vue'),
			props: route => ({ 
				projectId: parseInt(route.params.projectId as string),
				viewId: route.params.viewId ? parseInt(route.params.viewId as string): undefined,
			}),
		},
		{
			path: '/teams',
			name: 'teams.index',
			component: () => import('@/views/teams/ListTeams.vue'),
		},
		{
			path: '/teams/new',
			name: 'teams.create',
			component: () =>  import('@/views/teams/NewTeam.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/teams/:id/edit',
			name: 'teams.edit',
			component: () => import('@/views/teams/EditTeam.vue'),
		},
		{
			path: '/labels',
			name: 'labels.index',
			component: () => import('@/views/labels/ListLabels.vue'),
		},
		{
			path: '/labels/new',
			name: 'labels.create',
			component: () => import('@/views/labels/NewLabel.vue'),
			meta: {
				showAsModal: true,
			},
		},
		{
			path: '/filters/new',
			name: 'filters.create',
			component: () => import('@/views/filters/FilterNew.vue'),
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
			component: () => import('@/views/About.vue'),
		},
	],
})

export async function getAuthForRoute(to: RouteLocation, authStore) {
	if (authStore.authUser || authStore.authLinkShare) {
		return
	}
	
	// Check if password reset token is in query params
	const resetToken = to.query.userPasswordReset as string | undefined
	
	// Redirect to password reset page if we have a token stored
	if (resetToken && to.name !== 'user.password-reset.reset') {
		return {name: 'user.password-reset.reset', query: { userPasswordReset: resetToken }}
	}

	if (typeof resetToken === 'undefined' && to.name === 'user.password-reset.reset') {
		return {name: 'user.login'}
	}

	// Check if the route the user wants to go to is a route which needs authentication. We use this to 
	// redirect the user after successful login.
	const isValidUserAppRoute = ![
		'user.login',
		'user.password-reset.request',
		'user.password-reset.reset',
		'user.register',
		'link-share.auth',
		'openid.auth',
	].includes(to.name as string) &&
		localStorage.getItem('emailConfirmToken') === null
	
	if (isValidUserAppRoute) {
		saveLastVisited(to.name as string, to.params, to.query)
	}
	
	if (isValidUserAppRoute) {
		return {name: 'user.login'}
	}
	
	if(localStorage.getItem('emailConfirmToken') !== null && to.name !== 'user.login') {
		return {name: 'user.login'}
	}
}

router.beforeEach(async (to, from) => {
	const authStore = useAuthStore()

	await authStore.checkAuth()

	if(from.hash && from.hash.startsWith(LINK_SHARE_HASH_PREFIX)) {
		to.hash = from.hash
	}

	if (to.hash.startsWith(LINK_SHARE_HASH_PREFIX) && !authStore.authLinkShare) {
		saveLastVisited(to.name as string, to.params, to.query)
		return {
			name: 'link-share.auth',
			params: {
				share: to.hash.replace(LINK_SHARE_HASH_PREFIX, ''),
			},
		}
	}

	const newRoute = await getAuthForRoute(to, authStore)
	if(newRoute) {
		return {
			...newRoute,
			hash: to.hash,
		}
	}
	
	if(!to.fullPath.endsWith(to.hash)) {
		return to.fullPath + to.hash
	}
})

export default router
