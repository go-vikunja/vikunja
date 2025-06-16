import {computed, ref, watch, type Ref} from 'vue'
import {useRouter, type RouteLocationNormalized, type RouteLocationRaw, type RouteRecordName} from 'vue-router'
import equal from 'fast-deep-equal/es6'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type Filters = Record<string, any>

export interface UseRouteFiltersReturn<CurrentFilters extends Filters> {
	filters: Ref<CurrentFilters>
	hasDefaultFilters: Ref<boolean>
	setDefaultFilters: () => void
}

export function useRouteFilters<CurrentFilters extends Filters>(
	route: Ref<RouteLocationNormalized>,
	getDefaultFilters: (route: RouteLocationNormalized) => CurrentFilters,
	routeToFilters: (route: RouteLocationNormalized) => CurrentFilters,
	filtersToRoute: (filters: CurrentFilters) => RouteLocationRaw,
	routeAllowList: RouteRecordName[] = [],
) : UseRouteFiltersReturn<CurrentFilters> {
	const router = useRouter()

	const filters = ref<CurrentFilters>(routeToFilters(route.value))

	const routeFromFiltersFullPath = computed(() => router.resolve(filtersToRoute(filters.value)).fullPath)

	watch(
		route.value,
		(route, oldRoute) => {
			if (
				route?.name !== oldRoute?.name ||
				routeFromFiltersFullPath.value === route.fullPath ||
				!routeAllowList.includes(route.name ?? '')
			) {
				return
			}

			filters.value = routeToFilters(route)
		},
		{
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
