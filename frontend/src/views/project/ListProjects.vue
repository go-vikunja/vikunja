<template>
	<div
		v-cy="'projects-list'"
		class="content loader-container"
		:class="{'is-loading': loading}"
	>
		<header class="project-header">
			<div
				class="project-ownership-filter"
				role="group"
				:aria-label="$t('project.ownershipFilter.label')"
				aria-describedby="project-filter-description"
			>
				<BaseButton
					v-for="option in ownershipOptions"
					:key="option"
					class="ownership-button"
					:aria-pressed="ownership === option"
					@click="ownership = option"
				>
					{{ $t(option === 'savedFilters' ? 'navigation.savedFilters' : `project.ownershipFilter.${option}`) }}
				</BaseButton>
			</div>

			<div class="action-buttons">
				<XButton
					:to="{name: 'filters.create'}"
					icon="filter"
				>
					{{ $t('filters.create.title') }}
				</XButton>
				<XButton
					v-cy="'new-project'"
					:to="{name: 'project.create'}"
					icon="plus"
				>
					{{ $t('project.create.header') }}
				</XButton>
			</div>
		</header>

		<div class="project-toolbar">
			<p
				id="project-filter-description"
				class="ownership-description"
			>
				{{ $t(`project.ownershipFilter.${ownership}Description`) }}
			</p>

			<FancyCheckbox
				v-model="showArchived"
				v-cy="'show-archived-check'"
				class="archive-toggle"
			>
				{{ $t('project.showArchived') }}
			</FancyCheckbox>
		</div>

		<ProjectCardGrid
			:projects="projects"
			:show-archived="showArchived"
		/>
		<div
			v-if="!loading && projects.length === 0 && ownership !== 'all'"
			class="ownership-empty"
			role="status"
		>
			<p>{{ $t(`project.ownershipFilter.${ownership}Empty`) }}</p>
			<XButton
				variant="secondary"
				@click="ownership = 'all'"
			>
				{{ $t('project.ownershipFilter.showAll') }}
			</XButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import ProjectCardGrid from '@/components/project/partials/ProjectCardGrid.vue'

import {useTitle} from '@/composables/useTitle'
import {useStorage} from '@vueuse/core'

import {useProjects} from '@/composables/useProjects'
import {useAuthStore} from '@/stores/auth'

const {t} = useI18n()
const projectList = useProjects()
const authStore = useAuthStore()
const ownershipOptions = ['all', 'personal', 'shared', 'savedFilters'] as const
const ownership = ref<typeof ownershipOptions[number]>('all')

useTitle(() => t('project.projects'))
const showArchived = useStorage('showArchived', false)

const loading = computed(() => projectList.isLoading)
const projects = computed(() => {
	const visibleProjects = showArchived.value
		? projectList.projectsArray
		: projectList.projectsArray.filter(({is_archived}) => !is_archived)
	if (ownership.value === 'all') {
		return visibleProjects
	}
	if (ownership.value === 'savedFilters') {
		return projectList.savedFilterProjects
	}
	const userId = authStore.info?.id
	return visibleProjects.filter(project => {
		if (project.id <= 0 || !project.owner?.id || !userId) {
			return false
		}
		return ownership.value === 'personal'
			? project.owner.id === userId
			: project.owner.id !== userId
	})
})
</script>

<style lang="scss" scoped>
.project-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	flex-wrap: wrap;
	gap: 1.25rem;
	margin-block-end: 1rem;
}

.action-buttons {
	display: flex;
	justify-content: space-between;
	gap: 1rem;
}

.project-toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	flex-wrap: wrap;
	gap: 1rem;
	margin-block-end: 1rem;
}

.project-ownership-filter {
	display: inline-flex;
	flex-wrap: wrap;
	max-inline-size: 100%;
	padding: .5rem;
	border-radius: $radius;
	background: var(--white);
	box-shadow: var(--shadow-sm);
	font-size: .75rem;
}

.ownership-button {
	display: block;
	padding: .25rem .5rem;
	border-radius: $radius;
	white-space: nowrap;
	color: var(--primary);
	transition: background-color 100ms, color 100ms, box-shadow 100ms;

	&:not(:last-child) {
		margin-inline-end: .5rem;
	}

	&:hover,
	&[aria-pressed="true"] {
		background: var(--switch-view-active-background);
		color: var(--switch-view-color);
	}

	&[aria-pressed="true"] {
		box-shadow: var(--shadow-xs);
		font-weight: bold;
	}

	&:focus-visible {
		outline: 2px solid var(--primary);
		outline-offset: 2px;
	}
}

.archive-toggle {
	flex-shrink: 0;
}

.content .ownership-description {
	margin: 0;
	color: var(--text-muted);
	font-size: .875rem;
	line-height: 1.5;
}

.ownership-empty {
	padding: 3rem 1.5rem;
	border: 1px dashed var(--grey-300);
	border-radius: .75rem;
	background: var(--white);
	color: var(--text-muted);
	text-align: center;
}

@media screen and (max-width: $tablet) {
	.project-header {
		align-items: flex-start;
		flex-direction: column;
		margin-block-end: 1rem;
	}

	.action-buttons {
		inline-size: 100%;
		flex-direction: column;
		align-items: stretch;
	}

	.archive-toggle {
		margin-inline-start: auto;
	}
}

@media (prefers-reduced-motion: reduce) {
	.ownership-button {
		transition: none;
	}
}
</style>
