import {computed, type ComputedGetter, type DeepReadonly, type WritableComputedRef} from 'vue'
import {useRouteQuery} from '@vueuse/router'

import type {IProjectView} from '@/modelTypes/IProjectView'

export const INCLUDE_SUBPROJECTS_QUERY_PARAM = 'include_subprojects'

/**
 * Whether the current view should show tasks from subprojects.
 *
 * The flag saved on the view is only the default - toggling it from the filter popup
 * writes a query param instead, so it stays local to the url the user is on like every
 * other filter. Setting it back to the view default drops the param again.
 */
export function useIncludeSubprojects(
	viewGetter: ComputedGetter<IProjectView | DeepReadonly<IProjectView> | undefined>,
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
