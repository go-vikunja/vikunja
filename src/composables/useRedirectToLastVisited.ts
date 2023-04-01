import {useRouter} from 'vue-router'
import {getLastVisited, clearLastVisited} from '@/helpers/saveLastVisited'

export function useRedirectToLastVisited() {

	const router = useRouter()

	function getRedirectRoute() {
		const last = getLastVisited()
		if (last !== null) {
			clearLastVisited()
			return {
				name: last.name,
				params: last.params,
				query: last.query,
			}
		}

		return null
	}

	function redirectIfSaved() {
		const lastRoute = getRedirectRoute()
		if (lastRoute) {
			router.push(lastRoute)
		}

		router.push({name: 'home'})
	}

	return {
		redirectIfSaved,
		getRedirectRoute,
	}
}