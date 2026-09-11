<script setup lang="ts">
import {computed, watch, watchEffect} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useQuery} from '@tanstack/vue-query'

import {useBaseStore} from '@/stores/base'
import {useAuthStore} from '@/stores/auth'

import {saveProjectView} from '@/helpers/projectView'
import {projectQuery} from '@/client/queries/projects'

import ProjectList from '@/components/project/views/ProjectList.vue'
import ProjectGantt from '@/components/project/views/ProjectGantt.vue'
import ProjectTable from '@/components/project/views/ProjectTable.vue'
import ProjectKanban from '@/components/project/views/ProjectKanban.vue'

import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/constants/projectView'
import {saveProjectToHistory} from '@/modules/projectHistory'

const props = defineProps<{
	projectId: number,
	viewId: number,
}>()

const router = useRouter()
const route = useRoute()
const baseStore = useBaseStore()
const authStore = useAuthStore()
const project = useQuery(computed(() => projectQuery(props.projectId)))
const currentProject = computed(() => project.data.value)

const currentView = computed(() => {
	return currentProject.value?.views.find(v => v.id === props.viewId)
})

const isLoadingProject = project.isPending

watch(
	() => [currentProject.value, props.viewId] as const,
	([newCurrentProject, newViewId]) => {
		if (!newCurrentProject) {
			baseStore.handleSetCurrentProject({project: null})
			return
		}
		
		baseStore.handleSetCurrentProject({
			project: newCurrentProject,
			currentProjectViewId: newViewId,
		})
	}, {
		deep: true,
		immediate: true,
	},
)

function redirectToDefaultViewIfNecessary() {
	if (props.viewId === 0 || !currentView.value) {
		// Ideally, we would do that in the router redirect, but the projects (and therefore, the views) 
		// are not always loaded then.

		const defaultView  = authStore.settings.frontendSettings.defaultView
		const views = currentProject.value?.views ?? []

		let view
		if (defaultView !== DEFAULT_PROJECT_VIEW_SETTINGS.FIRST) {
			view = views.find(view => view.view_kind === defaultView)
		}

		// Use the first view as fallback if the default view is not available
		if (view === undefined && views.length > 0) {
			view = views[0]
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
	currentProject,
	redirectToDefaultViewIfNecessary,
)

watchEffect(() => {
	// Don't save to history if the user is not authenticated (e.g., during logout)
	if (authStore.authenticated) {
		saveProjectToHistory({id: props.projectId})
	}
})
watchEffect(() => saveProjectView(props.projectId, props.viewId))

watchEffect(() => baseStore.setCurrentProjectViewId(props.viewId))
</script>

<template>
	<ProjectList
		v-if="currentView?.view_kind === 'list'"
		:project-id="projectId"
		:is-loading-project="isLoadingProject"
		:view-id
	/>
	<ProjectGantt
		v-if="currentView?.view_kind === 'gantt'"
		:project-id="projectId"
		:route
		:is-loading-project="isLoadingProject"
		:view-id
	/>
	<ProjectTable
		v-if="currentView?.view_kind === 'table'"
		:project-id="projectId"
		:is-loading-project="isLoadingProject"
		:view-id
	/>
	<ProjectKanban
		v-if="currentView?.view_kind === 'kanban'"
		:project-id="projectId"
		:is-loading-project="isLoadingProject"
		:view-id
	/>
</template>
