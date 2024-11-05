<script setup lang="ts">
import {computed, watch} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useRoute, useRouter} from 'vue-router'
import {saveProjectView} from '@/helpers/projectView'

import ProjectList from '@/components/project/views/ProjectList.vue'
import ProjectGantt from '@/components/project/views/ProjectGantt.vue'
import ProjectTable from '@/components/project/views/ProjectTable.vue'
import ProjectKanban from '@/components/project/views/ProjectKanban.vue'
import {useAuthStore} from '@/stores/auth'
import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/modelTypes/IProjectView'

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
	if (props.viewId === 0 || !projectStore.projects[props.projectId]?.views.find(v => v.id === props.viewId)) {
		// Ideally, we would do that in the router redirect, but the projects (and therefore, the views) 
		// are not always loaded then.

		let view
		if (authStore.settings.frontendSettings.defaultView !== DEFAULT_PROJECT_VIEW_SETTINGS.FIRST) {
			view = projectStore.projects[props.projectId]?.views.find(v => v.viewKind === authStore.settings.frontendSettings.defaultView)
		}

		// Use the first view as fallback if the default view is not available
		if (view === undefined && projectStore.projects[props.projectId]?.views?.length > 0) {
			view = projectStore.projects[props.projectId]?.views[0]
		}

		if (view) {
			router.replace({
				name: 'project.view',
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
watch(
	() => [props.projectId, props.viewId],
	() => saveProjectView(props.projectId, props.viewId),
	{immediate: true},
)

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
