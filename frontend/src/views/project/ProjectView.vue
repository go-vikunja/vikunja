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

// watchEffect would be called every time the prop would get a value assigned, even if that value was the same as before.
// This resulted in loading and setting the project multiple times, even when navigating away from it.
// This caused wired bugs where the project background would be set on the home page but only right after setting a new 
// project background and then navigating to home. It also highlighted the project in the menu and didn't allow changing any
// of it, most likely due to the rights not being properly populated.
watch(
	() => props.projectId,
	// loadProject
	async (projectIdToLoad) => {
		// // Don't load the project if we either already loaded it or aren't dealing with a project at all currently and
		// // the currently loaded project has the right set.
		// if (
		// 	(
		// 		projectIdToLoad === loadedProjectId.value ||
		// 		typeof projectIdToLoad === 'undefined' ||
		// 		projectIdToLoad === currentProject.value?.id
		// 	)
		// 	&& typeof currentProject.value !== 'undefined' && currentProject.value.maxRight !== null
		// ) {
		// 	loadedProjectId.value = projectIdToLoad
		// 	return
		// }

		console.debug('Loading project, $route.params =', route.params, `, loadedProjectId = ${loadedProjectId.value}, currentProject = `, currentProject.value)

		// Set the current project to the one we're about to load so that the title is already shown at the top
		loadedProjectId.value = 0
		const projectFromStore = projectStore.projects[projectIdToLoad]
		if (projectFromStore) {
			baseStore.handleSetCurrentProject({project: projectFromStore, currentProjectViewId: props.viewId})
		}

		try {
			const loadedProject = await projectService.get({id: projectIdToLoad})
			baseStore.handleSetCurrentProject({project: loadedProject, currentProjectViewId: props.viewId})
		} finally {
			loadedProjectId.value = projectIdToLoad
		}
	},
	{immediate: true},
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
		:is-loading-project
		:view-id
	/>
	<ProjectGantt
		v-if="currentView?.viewKind === 'gantt'"
		:route
		:is-loading-project
		:view-id
	/>
	<ProjectTable
		v-if="currentView?.viewKind === 'table'"
		:project-id="projectId"
		:is-loading-project
		:view-id
	/>
	<ProjectKanban
		v-if="currentView?.viewKind === 'kanban'"
		:project-id="projectId"
		:is-loading-project
		:view-id
	/>
</template>
