import {useRouter} from 'vue-router'
import {getLastVisited, clearLastVisited} from '@/helpers/saveLastVisited'

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

	function redirectIfSaved() {
		const lastRoute = getLastVisitedRoute()
		if (!lastRoute) {
			return router.push({name: 'home'})
		}

		return router.push(lastRoute)
	}

	return {
		redirectIfSaved,
		getLastVisitedRoute,
	}
}
