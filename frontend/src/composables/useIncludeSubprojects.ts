import {computed, type ComputedGetter, type DeepReadonly} from 'vue'
import {useRouteQuery} from '@vueuse/router'

import type {ProjectView} from '@/client/generated'
import {INCLUDE_SUBPROJECTS_QUERY_PARAM} from '@/constants/projectView'

// The view's own flag is only the default; toggling writes a query param instead.
export function useIncludeSubprojects(
	viewGetter: ComputedGetter<ProjectView | DeepReadonly<ProjectView> | undefined>,
) {
	// The api ors the view's own setting in, so the query param cannot turn it back off.
	const fromView = computed(() => viewGetter()?.filter?.include_subprojects ?? false)
	const query = useRouteQuery<string>(
		INCLUDE_SUBPROJECTS_QUERY_PARAM,
		() => String(fromView.value),
	)

	const includeSubprojects = computed({
		get: () => query.value === 'true',
		set: (value: boolean) => {
			query.value = String(value)
		},
	})

	const overridesView = computed(() => includeSubprojects.value && !fromView.value)

	return {includeSubprojects, fromView, overridesView}
}
