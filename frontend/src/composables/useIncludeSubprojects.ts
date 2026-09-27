import {computed, type ComputedGetter, type DeepReadonly, type WritableComputedRef} from 'vue'
import {useRouteQuery} from '@vueuse/router'

import type {ProjectView} from '@/client/generated'

export const INCLUDE_SUBPROJECTS_QUERY_PARAM = 'include_subprojects'

// The view's own flag is only the default; toggling writes a query param instead.
export function useIncludeSubprojects(
	viewGetter: ComputedGetter<ProjectView | DeepReadonly<ProjectView> | undefined>,
): WritableComputedRef<boolean> {
	const viewDefault = computed(() => viewGetter()?.filter?.include_subprojects ?? false)
	const query = useRouteQuery<string>(
		INCLUDE_SUBPROJECTS_QUERY_PARAM,
		() => String(viewDefault.value),
	)

	return computed({
		get: () => query.value === 'true',
		set: (value: boolean) => {
			query.value = String(value)
		},
	})
}
