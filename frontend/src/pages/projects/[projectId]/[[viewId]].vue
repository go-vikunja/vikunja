<script setup lang="ts">
import {computed, ref, shallowReactive, watch, watchEffect} from 'vue'
import {useRoute, useRouter} from 'vue-router'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'

import {saveProjectView} from '@/helpers/projectView'
import ProjectService from '@/services/project'

import ProjectList from '@/components/project/views/ProjectList.vue'
import ProjectGantt from '@/components/project/views/ProjectGantt.vue'
import ProjectTable from '@/components/project/views/ProjectTable.vue'
import ProjectKanban from '@/components/project/views/ProjectKanban.vue'

import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/modelTypes/IProjectView'
import {saveProjectToHistory} from '@/modules/projectHistory'

const props = defineProps<{
	projectId: number,
	viewId: number,
}>()

const router = useRouter()
const baseStore = useBaseStore()
const projectStore = useProjectStore()
const authStore = useAuthStore()
const route = useRoute()

const currentProject = computed(() => projectStore.projects[props.projectId])

const currentView = computed(() => {
	return currentProject.value?.views.find(v => v.id === props.viewId)
})

const projectService = shallowReactive(new ProjectService())
const isLoadingProject = computed(() => projectService.loading)
const loadedProjectId = ref(0)

watch(
	() => props.projectId,
	// loadProject
	async (projectIdToLoad, oldProjectIdToLoad) => {

		console.debug('Loading project, $route.params =', route.params, `, loadedProjectId = ${loadedProjectId.value}, currentProject = `, currentProject.value)


		if (projectIdToLoad !== oldProjectIdToLoad) {
			loadedProjectId.value = 0
		}

		try {
			const loadedProject = await projectService.get({id: projectIdToLoad})

			// Here, we only set the new project in the projectStore.
			// Setting that projet as the current one in the baseStore is handled by the watcher below.
			projectStore.setProject(loadedProject)
		} finally {
			loadedProjectId.value = projectIdToLoad
		}
	},
	{immediate: true},
)

watch(
	() => [currentProject.value, props.viewId],
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

		let view
		if (defaultView !== DEFAULT_PROJECT_VIEW_SETTINGS.FIRST) {
			view = currentProject.value?.views.find(v => v.viewKind === defaultView)
		}

		// Use the first view as fallback if the default view is not available
		if (view === undefined && currentProject.value?.views?.length > 0) {
			view = currentProject.value?.views[0]
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

watchEffect(() => saveProjectToHistory({id: props.projectId}))
watchEffect(() => saveProjectView(props.projectId, props.viewId))

watchEffect(() => baseStore.setCurrentProjectViewId(props.viewId))
</script>

<template>
	<ProjectList
		v-if="currentView?.viewKind === 'list'"
		:project-id="projectId"
		:is-loading-project="isLoadingProject"
		:view-id
	/>
	<ProjectGantt
		v-if="currentView?.viewKind === 'gantt'"
		:route
		:is-loading-project="isLoadingProject"
		:view-id
	/>
	<ProjectTable
		v-if="currentView?.viewKind === 'table'"
		:project-id="projectId"
		:is-loading-project="isLoadingProject"
		:view-id
	/>
	<ProjectKanban
		v-if="currentView?.viewKind === 'kanban'"
		:project-id="projectId"
		:is-loading-project="isLoadingProject"
		:view-id
	/>
</template>
