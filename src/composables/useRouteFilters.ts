import {computed, ref, watch, type Ref} from 'vue'
import {useRouter, type RouteLocationNormalized, type RouteLocationRaw} from 'vue-router'
import cloneDeep from 'lodash.clonedeep'

export type Filters = Record<string, any>

export function useRouteFilters<CurrentFilters extends Filters>(
		route: Ref<RouteLocationNormalized>,
		routeToFilters: (route: RouteLocationNormalized) => CurrentFilters,
		filtersToRoute: (filters: CurrentFilters) => RouteLocationRaw,
	) {
	const router = useRouter()

	const filters = ref<CurrentFilters>(routeToFilters(route.value))

	const routeFromFiltersFullPath = computed(() => router.resolve(filtersToRoute(filters.value)).fullPath)

  watch(() => cloneDeep(route.value), (route, oldRoute) => {
    if (
			route.name !== oldRoute.name ||
			routeFromFiltersFullPath.value === route.fullPath
		) {
      return
    }
  
    filters.value = routeToFilters(route)
  })

  watch(
    filters,
    async () => {
      if (routeFromFiltersFullPath.value !== route.value.fullPath) {
        await router.push(routeFromFiltersFullPath.value)
      }
    },
    // only apply new route after all filters have changed in component cycle
    {flush: 'post'},
  )

  return {
    filters,
  }
}