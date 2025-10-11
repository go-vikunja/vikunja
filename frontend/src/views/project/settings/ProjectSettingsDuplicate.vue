<template>
	<CreateEdit
		v-model:loading="loadingModel"
		:title="$t('project.duplicate.title')"
		primary-icon="paste"
		:primary-label="$t('project.duplicate.label')"
		@primary="duplicate"
	>
		<p>{{ $t('project.duplicate.text') }}</p>
		<ProjectSearch v-model="parentProject" />
	</CreateEdit>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProject, useProjectStore} from '@/stores/projects'
import type {IProject} from '@/modelTypes/IProject'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('project.duplicate.title'))

const route = useRoute()
const projectStore = useProjectStore()

const {project, isLoading, duplicateProject} = useProject(route.params.projectId)

const parentProject = ref<IProject | null>(null)
const isDuplicating = ref(false)

const loadingModel = computed({
	get: () => isDuplicating.value || isLoading.value,
	set(value: boolean) {
		isDuplicating.value = value
	},
})
watch(
	() => project.parentProjectId,
	parentProjectId => {
		parentProject.value = projectStore.projects[parentProjectId]
	},
	{immediate: true},
)

async function duplicate() {
	isDuplicating.value = true

	try {
		await duplicateProject(parentProject.value?.id ?? 0)
		success({message: t('project.duplicate.success')})
	} finally {
		isDuplicating.value = false
	}
}
</script>
