import {computed, ref, watch, type Ref} from 'vue'
import {useRouter, type RouteLocationNormalized, type RouteLocationRaw} from 'vue-router'
import cloneDeep from 'lodash.clonedeep'

export type Filter = Record<string, any>

export function useRouteFilter<F extends Filter = Filter>(
		route: Ref<RouteLocationNormalized>,
		routeToFilter: (route: RouteLocationNormalized) => F,
		filterToRoute: (filter: F) => RouteLocationRaw,
	) {
	const router = useRouter()

	const filters = ref<F>(routeToFilter(route.value))

	const routeFromFiltersFullPath = computed(() => router.resolve(filterToRoute(filters.value)).fullPath)

  watch(() => cloneDeep(route.value), (route, oldRoute) => {
    if (
			route.name !== oldRoute.name ||
			routeFromFiltersFullPath.value === route.fullPath
		) {
      return
    }
  
    filters.value = routeToFilter(route)
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