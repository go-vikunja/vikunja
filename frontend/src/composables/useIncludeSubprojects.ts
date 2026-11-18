import {computed, type ComputedGetter, type DeepReadonly} from 'vue'
import {useRouteQuery} from '@vueuse/router'

import type {ProjectView} from '@/client/generated'
import {INCLUDE_SUBPROJECTS_QUERY_PARAM} from '@/constants/projectView'
import {useProjects} from '@/composables/useProjects'

// The view's own flag is only the default; toggling writes a query param instead.
export function useIncludeSubprojects(
	viewGetter: ComputedGetter<ProjectView | DeepReadonly<ProjectView> | undefined>,
) {
	// The api ors the view's own setting in, so the query param cannot turn it back off.
	const projectList = useProjects()
	const fromView = computed(() => {
		const view = viewGetter()

		if (view === undefined) {
			return false
		}

		const project = projectList.projects[view.project_id ?? 0]
		const viewIncludesSubprojects = view.filter?.include_subprojects === true
		const projectIncludesSubprojects = project?.include_subprojects === true

		return viewIncludesSubprojects || projectIncludesSubprojects
	})
	const param = useRouteQuery<string | undefined>(INCLUDE_SUBPROJECTS_QUERY_PARAM)

	const includeSubprojects = computed({
		get: () => param.value === undefined ? fromView.value : param.value === 'true',
		set: (value: boolean) => {
			param.value = value === fromView.value ? undefined : String(value)
		},
	})

	const overridesView = computed(() => includeSubprojects.value && !fromView.value)

	return {
		includeSubprojects,
		fromView,
		overridesView,
		queryParam: computed(() => param.value),
	}
}
