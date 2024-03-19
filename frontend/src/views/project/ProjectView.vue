<script setup lang="ts">
import {computed, watch} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useRoute, useRouter} from 'vue-router'

import ProjectList from '@/components/project/views/ProjectList.vue'
import ProjectGantt from '@/components/project/views/ProjectGantt.vue'
import ProjectTable from '@/components/project/views/ProjectTable.vue'
import ProjectKanban from '@/components/project/views/ProjectKanban.vue'

const {
	projectId,
	viewId,
} = defineProps<{
	projectId: number,
	viewId: number,
}>()

const router = useRouter()
const projectStore = useProjectStore()

const currentView = computed(() => {
	const project = projectStore.projects[projectId]

	return project?.views.find(v => v.id === viewId)
})

function redirectToFirstViewIfNecessary() {
	if (viewId === 0) {
		// Ideally, we would do that in the router redirect, but the projects (and therefore, the views) 
		// are not always loaded then.
		const firstViewId = projectStore.projects[projectId]?.views[0].id
		if (firstViewId) {
			router.replace({
				name: 'project.view',
				params: {
					projectId,
					viewId: firstViewId,
				},
			})
		}
	}
}

watch(
	() => viewId,
	redirectToFirstViewIfNecessary,
	{immediate: true},
)

watch(
	() => projectStore.projects[projectId],
	redirectToFirstViewIfNecessary,
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
