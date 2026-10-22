import {useRouter, type Router} from 'vue-router'
import {getLastVisited, clearLastVisited} from '@/helpers/saveLastVisited'

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

export function redirectToLastVisited(router: Router) {
	const lastRoute = getLastVisitedRoute()
	if (!lastRoute) {
		return router.push({name: 'home'})
	}

	return router.push(lastRoute)
}

export function useRedirectToLastVisited() {

	const router = useRouter()

	function redirectIfSaved() {
		return redirectToLastVisited(router)
	}

	return {
		redirectIfSaved,
		getLastVisitedRoute,
	}
}
