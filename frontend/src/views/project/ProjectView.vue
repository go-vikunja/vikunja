<script setup lang="ts">
import {computed} from 'vue'
import {useProjectStore} from '@/stores/projects'

import ProjectList from '@/views/project/ProjectList.vue'
import ProjectGantt from '@/views/project/ProjectGantt.vue'
import ProjectTable from '@/views/project/ProjectTable.vue'
import ProjectKanban from '@/views/project/ProjectKanban.vue'
import {useRoute} from 'vue-router'

const {
	projectId,
	viewId,
} = defineProps<{
	projectId: number,
	viewId: number,
}>()

const projectStore = useProjectStore()

const currentView = computed(() => {
	const project = projectStore.projects[projectId]

	return project?.views.find(v => v.id === viewId)
})

const route = useRoute()
</script>

<template>
	<ProjectList
		v-if="currentView?.viewKind === 'list'"
		:project-id="projectId"
		:view="currentView"
	/>
	<ProjectGantt
		v-if="currentView?.viewKind === 'gantt'"
		:route
		:view="currentView"
	/>
	<ProjectTable
		v-if="currentView?.viewKind === 'table'"
		:project-id="projectId"
		:view="currentView"
	/>
	<ProjectKanban
		v-if="currentView?.viewKind === 'kanban'"
		:project-id="projectId"
		:view="currentView"
	/>
</template>

<style scoped lang="scss">

</style>