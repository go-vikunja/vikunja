<script setup lang="ts">
import {computed, ref, shallowReactive, watch, watchEffect} from 'vue'
import {useRoute, useRouter, type RouteLocationNormalizedLoaded} from 'vue-router'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'

import {getProjectViewId, saveProjectView} from '@/helpers/projectView'
import ProjectService from '@/services/project'

import ProjectList from '@/components/project/views/ProjectList.vue'
import ProjectGantt from '@/components/project/views/ProjectGantt.vue'
import ProjectTable from '@/components/project/views/ProjectTable.vue'
import ProjectKanban from '@/components/project/views/ProjectKanban.vue'

import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/modelTypes/IProjectView'
import {saveProjectToHistory} from '@/modules/projectHistory'

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
	props: route => {
		// https://github.com/posva/unplugin-vue-router/discussions/513#discussioncomment-10695660
		const castedRoute = route as RouteLocationNormalizedLoaded<'project'>
		return {
			projectId: Number(castedRoute.params.projectId),
			viewId: castedRoute.params.viewId ? parseInt(castedRoute.params.viewId): undefined,
		}
	},
})

const props = defineProps<{
	projectId: number,
	viewId?: number,
}>()

const router = useRouter()
const route = useRoute()

const baseStore = useBaseStore()
const projectStore = useProjectStore()
const authStore = useAuthStore()

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
	if (props.viewId === undefined || props.viewId === 0 || !currentView.value) {
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
	currentProject,
	redirectToDefaultViewIfNecessary,
)

// using a watcher instead of beforeEnter because beforeEnter is not called when only the viewId changes
watchEffect(() => saveProjectView(props.projectId, props.viewId))
watchEffect(() => saveProjectToHistory({id: props.projectId}))
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
