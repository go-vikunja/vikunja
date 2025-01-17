import { createRouter, createWebHistory, type RouteLocation } from 'vue-router'
import { routes, handleHotUpdate } from 'vue-router/auto-routes'

import {saveLastVisited} from '@/helpers/saveLastVisited'

import {getProjectViewId} from '@/helpers/projectView'
import {LINK_SHARE_HASH_PREFIX} from '@/constants/linkShareHash'

import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'

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
		return {left: 0, top: 0}
	},
	routes: [
		...routes,
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
	],
})

// Update routes at runtime without reloading the page
if (import.meta.hot) { 
	handleHotUpdate(router) 
}

export async function getAuthForRoute(to: RouteLocation, authStore) {
	if (authStore.authUser || authStore.authLinkShare) {
		return
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
		localStorage.getItem('passwordResetToken') === null &&
		localStorage.getItem('emailConfirmToken') === null &&
		!(to.name === 'home' && (typeof to.query.userPasswordReset !== 'undefined' || typeof to.query.userEmailConfirm !== 'undefined'))
	
	if (isValidUserAppRoute) {
		saveLastVisited(to.name as string, to.params, to.query)
	}
	
	const baseStore = useBaseStore()
	// When trying this before the current user was fully loaded we might get a flash of the login screen 
	// in the user shell. To make sure this does not happen we check if everything is ready before trying.
	if (!baseStore.ready) {
		return
	}

	if (isValidUserAppRoute) {
		return {name: 'user.login'}
	}
	
	if(localStorage.getItem('passwordResetToken') !== null && to.name !== 'user.password-reset.reset') {
		return {name: 'user.password-reset.reset'}
	}
	
	if(localStorage.getItem('emailConfirmToken') !== null && to.name !== 'user.login') {
		return {name: 'user.login'}
	}
}

router.beforeEach(async (to, from) => {
	const authStore = useAuthStore()

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