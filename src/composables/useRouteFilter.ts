import {reactive, watch, type Ref} from 'vue'
import {useRouter, type RouteLocationNormalized, type RouteLocationRaw} from 'vue-router'
import cloneDeep from 'lodash.clonedeep'

export type Filter = Record<string, any>

export function useRouteFilter<F extends Filter = Filter>(
		route: Ref<RouteLocationNormalized>,
		routeToFilter: (route: RouteLocationNormalized) => F,
		filterToRoute: (filter: F) => RouteLocationRaw,
	) {
	const router = useRouter()

	const filters: F = reactive(routeToFilter(route.value))

  watch(() => cloneDeep(route.value), (route, oldRoute) => {
    if (route.name !== oldRoute.name) {
      return
    }
    const filterFullPath = router.resolve(filterToRoute(filters)).fullPath
    if (filterFullPath === route.fullPath) {
      return
    }
  
    Object.assign(filters, routeToFilter(route))
  })

  watch(
    filters,
    async () => {
      const newRouteFullPath = router.resolve(filterToRoute(filters)).fullPath
      if (newRouteFullPath !== route.value.fullPath) {
        await router.push(newRouteFullPath)
      }
    },
    // only apply new route after all filters have changed in component cycle
    {flush: 'post'},
  )

  return {
    filters,
  }
}