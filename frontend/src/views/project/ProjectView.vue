<script setup lang="ts">
import {computed, watch} from 'vue'
import {useProjectStore} from '@/stores/projects'
import {useRoute, useRouter} from 'vue-router'

import ProjectList from '@/views/project/ProjectList.vue'
import ProjectGantt from '@/views/project/ProjectGantt.vue'
import ProjectTable from '@/views/project/ProjectTable.vue'
import ProjectKanban from '@/views/project/ProjectKanban.vue'

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

watch(
	() => viewId,
	() => {
		if (viewId === 0) {
			// Ideally, we would do that in the router redirect, but we the projects (and therefore, the views) 
			// are not always loaded then.
			const viewId = projectStore.projects[projectId].views[0].id
			router.replace({
				name: 'project.view',
				params: {
					projectId,
					viewId,
				},
			})
		}
	},
	{immediate: true},
)

const route = useRoute()
</script>

<template>
	<ProjectList
		v-if="currentView?.viewKind === 'list'"
		:project-id="projectId"
		:viewId
	/>
	<ProjectGantt
		v-if="currentView?.viewKind === 'gantt'"
		:route
		:viewId
	/>
	<ProjectTable
		v-if="currentView?.viewKind === 'table'"
		:project-id="projectId"
		:viewId
	/>
	<ProjectKanban
		v-if="currentView?.viewKind === 'kanban'"
		:project-id="projectId"
		:viewId
	/>
</template>

<style scoped lang="scss">

</style>