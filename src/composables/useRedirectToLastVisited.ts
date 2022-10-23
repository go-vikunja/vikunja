import {useRouter} from 'vue-router'
import {getLastVisited, clearLastVisited} from '@/helpers/saveLastVisited'

export function useRedirectToLastVisited() {

	const router = useRouter()

	function redirectIfSaved() {
		const last = getLastVisited()
		if (last !== null) {
			router.push({
				name: last.name,
				params: last.params,
				query: last.query,
			})
			clearLastVisited()
			return
		}

		router.push({name: 'home'})
	}

	return {
		redirectIfSaved,
	}
}