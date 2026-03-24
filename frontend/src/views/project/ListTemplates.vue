<template>
	<div
		class="content loader-container"
		:class="{'is-loading': loading}"
	>
		<header class="project-header">
			<h3>{{ $t('project.template.title') }}</h3>
		</header>

		<p v-if="!loading && templates.length === 0">
			{{ $t('project.template.none') }}
		</p>

		<ProjectCardGrid
			v-else
			:projects="templates"
			:show-archived="false"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import ProjectCardGrid from '@/components/project/partials/ProjectCardGrid.vue'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'
import type {IProject} from '@/modelTypes/IProject'

const {t} = useI18n()
const projectStore = useProjectStore()

useTitle(() => t('project.template.title'))

const loading = computed(() => projectStore.isLoading)
const templates = computed(() => projectStore.templateProjects as IProject[])
</script>

<style lang="scss" scoped>
.project-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	gap: 1rem;
	margin-block-end: 1rem;
}
</style>
