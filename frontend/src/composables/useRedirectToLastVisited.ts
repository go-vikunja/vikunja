import {useRouter} from 'vue-router'
import {getLastVisited, clearLastVisited, getLastVisitedPage} from '@/helpers/saveLastVisited'
import {useAuthStore} from '@/stores/auth'
import {DEFAULT_PAGE} from '@/constants/defaultPage'

export function useRedirectToLastVisited() {

	const router = useRouter()

	function getLastVisitedRoute() {
		const last = getLastVisited()
		if (last === null) {
			return null
		}

		clearLastVisited()
		return {
			name: last.name,
			params: last.params,
			query: last.query,
		}
	}

	function getDefaultPageRoute() {
		const authStore = useAuthStore()
		const defaultPage = authStore.settings?.frontendSettings?.defaultPage

		switch (defaultPage) {
			case DEFAULT_PAGE.UPCOMING:
				return {name: 'tasks.range'}
			case DEFAULT_PAGE.DEFAULT_PROJECT: {
				const projectId = authStore.settings?.defaultProjectId
				if (projectId) {
					return {name: 'project.index', params: {projectId}}
				}
				return {name: 'home'}
			}
			case DEFAULT_PAGE.LAST_VISITED: {
				const last = getLastVisitedPage()
				if (last) {
					return {name: last.name, params: last.params, query: last.query}
				}
				return {name: 'home'}
			}
			default:
				return {name: 'home'}
		}
	}

	function redirectIfSaved() {
		const lastRoute = getLastVisitedRoute()
		if (!lastRoute) {
			return router.push(getDefaultPageRoute())
		}

		return router.push(lastRoute)
	}

	return {
		redirectIfSaved,
		getLastVisitedRoute,
	}
}
