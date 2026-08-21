<template>
	<div
		v-cy="'projects-list'"
		class="content loader-container"
		:class="{'is-loading': loading}"
	>
		<PageHeader :title="$t('project.title')">
			<template #actions>
				<FancyCheckbox
					v-model="showArchived"
					v-cy="'show-archived-check'"
				>
					{{ $t('project.showArchived') }}
				</FancyCheckbox>
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
			</template>
		</PageHeader>

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
import PageHeader from '@/components/layout/PageHeader.vue'

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
