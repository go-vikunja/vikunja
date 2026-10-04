<template>
	<ProjectWrapper
		:project-id="projectId"
		:view-id="0"
		:is-loading-project="isLoadingProject"
	>
		<RiskListPage :project-id="projectId" />
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {useQuery} from '@tanstack/vue-query'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import RiskListPage from '@/components/risks/RiskListPage.vue'

import {projectQuery} from '@/client/queries/projects'
import {useTitle} from '@/composables/useTitle'
import {useBaseStore} from '@/stores/base'

const props = defineProps<{
	projectId: number
}>()

const {t} = useI18n({useScope: 'global'})
const baseStore = useBaseStore()

// The wrapper shows the title and the view tabs of the current project, so it is set the way a view does.
const project = useQuery(computed(() => projectQuery(props.projectId)))
const isLoadingProject = project.isPending

watch(
	() => project.data.value,
	current => baseStore.setCurrentProject(current ?? null, 0),
	{immediate: true, deep: true},
)

useTitle(() => t('risks.title'))
</script>
