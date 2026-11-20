import {computed, type Ref} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useRouteQuery} from '@vueuse/router'

type RouteQueryValue = string | string[] | undefined

/**
 * Reads a query parameter from the view rendered behind a task modal.
 *
 * While a task is open, the current route is the task route, but the content
 * behind the modal is still the project view from `history.state.backdropView`.
 * Reading the query with `useRouteQuery` would see the task route and drop the
 * project view's filter query. Writes are ignored while the backdrop is active:
 * that view is not interactive, and writing would navigate away from the task.
 */
export function useBackdropRouteQuery<T extends RouteQueryValue = string>(name: string): Ref<T | undefined> {
	const router = useRouter()
	const route = useRoute()
	const routeQuery = useRouteQuery(name)

	const backdropView = computed(() => route.fullPath ? window.history.state?.backdropView : undefined)
	const backdropRoute = computed(() => {
		return typeof backdropView.value === 'string' ? router.resolve(backdropView.value) : null
	})

	return computed<T | undefined>({
		get() {
			if (backdropRoute.value) {
				return (backdropRoute.value.query[name] ?? undefined) as T | undefined
			}

			return (routeQuery.value ?? undefined) as T | undefined
		},
		set(value) {
			if (backdropRoute.value) {
				return
			}

			routeQuery.value = value
		},
	})
}
