import {computed, ref, watch, type Ref} from 'vue'
import {useRouter, type RouteLocationNormalized, type RouteLocationRaw} from 'vue-router'
import equal from 'fast-deep-equal/es6'

export type Filters = Record<string, any>

export function useRouteFilters<CurrentFilters extends Filters>(
	route: Ref<RouteLocationNormalized>,
	getDefaultFilters: (route: RouteLocationNormalized) => CurrentFilters,
	routeToFilters: (route: RouteLocationNormalized) => CurrentFilters,
	filtersToRoute: (filters: CurrentFilters) => RouteLocationRaw,
) {
	const router = useRouter()

	const filters = ref<CurrentFilters>(routeToFilters(route.value))

	const routeFromFiltersFullPath = computed(() => router.resolve(filtersToRoute(filters.value)).fullPath)

	watch(
		route,
		(route, oldRoute) => {
			if (
				route?.name !== oldRoute?.name ||
				routeFromFiltersFullPath.value === route.fullPath
			) {
				return
			}

			filters.value = routeToFilters(route)
		},
		{
			deep: true,
			immediate: true, // set the filter from the initial route
		},
	)

	watch(
		filters,
		async () => {
			if (routeFromFiltersFullPath.value !== route.value.fullPath) {
				await router.push(routeFromFiltersFullPath.value)
			}
		},
		// only apply new route after all filters have changed in component cycle
		{
			deep: true,
			flush: 'post',
		},
	)

	const hasDefaultFilters = ref(false)
	watch(
		[filters, route],
		([filters, route]) => {
			hasDefaultFilters.value = equal(filters, getDefaultFilters(route))
		},
		{
			deep: true,
			immediate: true,
		},
	)

	function setDefaultFilters() {
		filters.value = getDefaultFilters(route.value)
	}

	return {
		filters,
		hasDefaultFilters,
		setDefaultFilters,
	}
}