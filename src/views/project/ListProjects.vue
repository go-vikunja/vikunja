<template>
	<div class="content loader-container" :class="{'is-loading': loading}" v-cy="'projects-list'">
		<header class="project-header">
			<fancycheckbox v-model="showArchived" v-cy="'show-archived-check'">
				{{ $t('project.showArchived') }}
			</fancycheckbox>

			<div class="action-buttons">
				<x-button :to="{name: 'filters.create'}" icon="filter">
					{{ $t('filters.create.title') }}
				</x-button>
				<x-button :to="{name: 'project.create'}" icon="plus" v-cy="'new-project'">
					{{ $t('project.create.header') }}
				</x-button>
			</div>
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

import BaseButton from '@/components/base/BaseButton.vue'
import Fancycheckbox from '@/components/input/fancycheckbox.vue'
import ProjectCardGrid from '@/components/project/partials/ProjectCardGrid.vue'

import {getProjectTitle} from '@/helpers/getProjectTitle'
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
.project-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 1rem;
	margin-bottom: 1rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.action-buttons {
	display: flex;
	justify-content: space-between;
	gap: 1rem;

	@media screen and (max-width: $tablet) {
		width: 100%;
		flex-direction: column;
		align-items: stretch;
	}
}

.project:not(:first-child) {
	margin-top: 1rem;
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
	margin-left: .5rem;
}
</style>