<template>
    <ul class="project-grid">
			<li
				v-for="(item, index) in filteredProjects"
				:key="`project_${item.id}_${index}`"
				class="project-grid-item"
			>
				<ProjectCard :project="item" />
			</li>
    </ul>
</template>

<script lang="ts" setup>
import {computed, type PropType} from 'vue'
import type {IProject} from '@/modelTypes/IProject'

import ProjectCard from './ProjectCard.vue'

const props = defineProps({
	projects: {
		type: Array as PropType<IProject[]>,
		default: () => [],
	},
	showArchived: {
		default: false,
		type: Boolean,
	},
	itemLimit: {
		type: Boolean,
		default: false,
	},
})

const filteredProjects = computed(() => {
	return props.showArchived
		? props.projects
		: props.projects.filter(l => !l.isArchived)
})
</script>

<style lang="scss" scoped>
$project-height: 150px;
$project-spacing: 1rem;

.project-grid {
	margin: 0; // reset li
	project-style-type: none;
	display: grid;
	grid-template-columns: repeat(var(--project-columns), 1fr);
	grid-auto-rows: $project-height;
	gap: $project-spacing;

	@media screen and (min-width: $mobile) {
		--project-rows: 4;
		--project-columns: 1;
	}

	@media screen and (min-width: $mobile) and (max-width: $tablet) {
		--project-columns: 2;
	}

	@media screen and (min-width: $tablet) and (max-width: $widescreen) {
		--project-columns: 3;
		--project-rows: 3;
	}

	@media screen and (min-width: $widescreen) {
		--project-columns: 5;
		--project-rows: 2;
	}
}

.project-grid-item {
	display: grid;
	margin-top: 0; // remove padding coming form .content li + li
}
</style>