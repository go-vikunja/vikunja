<template>
	<nav class="menu" v-if="favoriteProjects">
		<ProjectsNavigation v-model="favoriteProjects" :can-edit-order="false"/>
	</nav>

	<nav class="menu">
		<ProjectsNavigation v-model="projects" :can-edit-order="true"/>
	</nav>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useProjectStore} from '@/stores/projects'
import ProjectsNavigation from '@/components/home/ProjectsNavigation.vue'

const projectStore = useProjectStore()

await projectStore.loadProjects()

const projects = computed({
	get() {
		return projectStore.notArchivedRootProjects
			.sort((a, b) => a.position - b.position)
	},
	set() {	}, // Vue will complain about the component not being writable - but we never need to write here. The setter is only here to silence the warning.
})
const favoriteProjects = computed(() => projectStore.projectsArray
	.filter(p => !p.isArchived && p.isFavorite)
	.sort((a, b) => a.position - b.position))
</script>

<style lang="scss" scoped>
.menu {
	padding-top: math.div($navbar-padding, 2);
}
</style>