<template>
	<nav class="menu" v-if="favoriteProjects">
		<ProjectsNavigation :model-value="favoriteProjects" :can-edit-order="false"/>
	</nav>

	<nav class="menu">
		<ProjectsNavigation :model-value="projects" :can-edit-order="true"/>
	</nav>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useProjectStore} from '@/stores/projects'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'

const projectStore = useProjectStore()

await projectStore.loadProjects()

const projects = computed(() => projectStore.notArchivedRootProjects
	.sort((a, b) => a.position - b.position))
const favoriteProjects = computed(() => projectStore.favoriteProjects
	.sort((a, b) => a.position - b.position))
</script>

<style lang="scss" scoped>
.menu + .menu{
	padding-top: math.div($navbar-padding, 2);
}
</style>