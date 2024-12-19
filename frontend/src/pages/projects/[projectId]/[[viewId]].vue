<script setup lang="ts">
import {computed, watch, watchEffect} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useRoute, useRouter} from 'vue-router'
import {getProjectViewId, saveProjectView} from '@/helpers/projectView'

import ProjectList from '@/components/project/views/ProjectList.vue'
import ProjectGantt from '@/components/project/views/ProjectGantt.vue'
import ProjectTable from '@/components/project/views/ProjectTable.vue'
import ProjectKanban from '@/components/project/views/ProjectKanban.vue'
import {useAuthStore} from '@/stores/auth'
import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/modelTypes/IProjectView'

definePage({
	name: 'project',
	// beforeEnter(to) {
	// 	if (to.name !== 'project') {
	// 		throw new Error()
	// 	}

	// 	const projectId = Number(to.params?.projectId)
	// 	const viewIdFromRoute = Number(to.params?.viewId)

	// 	const newViewid = getProjectViewId(projectId) ?? 0

	// 	if (!viewIdFromRoute || viewIdFromRoute !== newViewid) {
	// 		return {
	// 			name: 'project',
	// 			replace: true,
	// 			params: {
	// 				projectId,
	// 				viewId: newViewid,
	// 			},
	// 		}
	// 	}
	// },
	props: route => ({ 
		projectId: parseInt(route.params.projectId as string),
		viewId: route.params.viewId ? parseInt(route.params.viewId as string): undefined,
	}),
})

const props = defineProps<{
	projectId: number,
	viewId: number,
}>()

const router = useRouter()
const projectStore = useProjectStore()
const authStore = useAuthStore()

const currentProject = computed(() => projectStore.projects[props.projectId])

const currentView = computed(() => {
	return currentProject.value?.views.find(v => v.id === props.viewId)
})

function redirectToDefaultViewIfNecessary() {
	if (props.viewId === 0 || !currentView.value) {
		// Ideally, we would do that in the router redirect, but the projects (and therefore, the views) 
		// are not always loaded then.

		let view
		const defaultView = authStore.settings.frontendSettings.defaultView
		if (defaultView !== DEFAULT_PROJECT_VIEW_SETTINGS.FIRST) {
			view = currentProject.value?.views.find(v => v.viewKind === defaultView)
		}

		// Use the first view as fallback if the default view is not available
		if (view === undefined && currentProject.value?.views?.length > 0) {
			view = currentProject.value?.views[0]
		}

		if (view) {
			router.replace({
				name: 'project',
				params: {
					projectId: props.projectId,
					viewId: view.id,
				},
			})
		}
	}
}

watch(
	() => props.viewId,
	redirectToDefaultViewIfNecessary,
	{immediate: true},
)

watch(
	() => projectStore.projects[props.projectId],
	redirectToDefaultViewIfNecessary,
)

// using a watcher instead of beforeEnter because beforeEnter is not called when only the viewId changes
watchEffect(() => saveProjectView(props.projectId, props.viewId))

const route = useRoute()
</script>

<template>
	<ProjectList
		v-if="currentView?.viewKind === 'list'"
		:project-id="projectId"
		:view-id
	/>
	<ProjectGantt
		v-if="currentView?.viewKind === 'gantt'"
		:route
		:view-id
	/>
	<ProjectTable
		v-if="currentView?.viewKind === 'table'"
		:project-id="projectId"
		:view-id
	/>
	<ProjectKanban
		v-if="currentView?.viewKind === 'kanban'"
		:project-id="projectId"
		:view-id
	/>
</template>
