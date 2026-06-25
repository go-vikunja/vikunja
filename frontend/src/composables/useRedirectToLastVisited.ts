import {useRouter} from 'vue-router'

import {getDefaultPageRoute} from '@/helpers/getDefaultPageRoute'
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

	async function redirectIfSaved() {
		const lastRoute = getLastVisitedRoute()
		if (!lastRoute) {
			return router.push(await getDefaultPageRoute() ?? {name: 'home'})
		}

		return router.push(lastRoute)
	}

	return {
		redirectIfSaved,
		getLastVisitedRoute,
	}
}
