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
.project-grid {
	--project-grid-item-height: 150px;
	--project-grid-gap: 1rem;
	margin: 0; // reset li
	list-style-type: none;
	display: grid;
	grid-template-columns: repeat(var(--project-grid-columns), 1fr);
	grid-auto-rows: var(--project-grid-item-height);
	gap: var(--project-grid-gap);

	@media screen and (min-width: $mobile) {
		--project-grid-columns: 1;
	}

	@media screen and (min-width: $mobile) and (max-width: $tablet) {
		--project-grid-columns: 2;
	}

	@media screen and (min-width: $tablet) and (max-width: $widescreen) {
		--project-grid-columns: 3;
	}

	@media screen and (min-width: $widescreen) {
		--project-grid-columns: 5;
	}
}

.project-grid-item {
	display: grid;
	margin-top: 0; // remove padding coming form .content li + li
}
</style>