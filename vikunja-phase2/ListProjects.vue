<template>
	<div
		v-cy="'projects-list'"
		class="content-widescreen loader-container"
		:class="{'is-loading': loading}"
	>
		<h2>{{ $t('project.projects') }}</h2>
		<p class="has-text-grey">
			{{ $t('project.pageDescription') }}
		</p>

		<hr class="page-separator">

		<header class="project-header">
			<div class="action-buttons">
				<XButton
					:to="{name: 'filters.create'}"
					icon="filter"
					:shadow="false"
				>
					{{ $t('filters.create.title') }}
				</XButton>
				<XButton
					v-cy="'new-project'"
					:to="{name: 'project.create'}"
					icon="plus"
					:shadow="false"
				>
					{{ $t('project.create.header') }}
				</XButton>
			</div>
			<FancyCheckbox
				v-model="showArchived"
				v-cy="'show-archived-check'"
			>
				{{ $t('project.showArchived') }}
			</FancyCheckbox>
		</header>

		<ProjectCardGrid
			:projects="projects"
			:show-archived="showArchived"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import ProjectCardGrid from '@/components/project/partials/ProjectCardGrid.vue'

import {useTitle} from '@/composables/useTitle'
import {useStorage} from '@vueuse/core'

import {useProjectStore} from '@/stores/projects'

const {t} = useI18n()
const projectStore = useProjectStore()

useTitle(() => t('project.title'))
const showArchived = useStorage('showArchived', false)

const loading = computed(() => projectStore.isLoading)
const projects = computed(() => {
	return showArchived.value
		? projectStore.projectsArray
		: projectStore.projectsArray.filter(({isArchived}) => !isArchived)
})
</script>

<style lang="scss" scoped>
.content-widescreen {
	max-inline-size: 900px;
	margin: 0 auto;
	padding: 1.5rem 1rem;
}

.page-separator {
	border: none;
	border-block-start: 2px solid var(--grey-200);
	margin-block: 1rem 1.5rem;
}

.project-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 1rem;
	margin-block-end: 1.5rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.action-buttons {
	display: flex;
	justify-content: space-between;
	gap: 1rem;

	@media screen and (max-width: $tablet) {
		inline-size: 100%;
		flex-direction: column;
		align-items: stretch;
	}
}

.project:not(:first-child) {
	margin-block-start: 1rem;
}

.project-title {
	display: flex;
	align-items: center;
}

.is-archived {
	font-size: 0.75rem;
	border: 1px solid var(--grey-500);
	color: $grey !important;
	padding: 2px 4px;
	border-radius: 3px;
	font-family: $vikunja-font;
	background: var(--white-translucent);
	margin-inline-start: .5rem;
}
</style>
