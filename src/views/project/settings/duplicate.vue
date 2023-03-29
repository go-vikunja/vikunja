<template>
	<create-edit
		:title="$t('project.duplicate.title')"
		primary-icon="paste"
		:primary-label="$t('project.duplicate.label')"
		@primary="duplicateProject"
		:loading="projectDuplicateService.loading"
	>
		<p>{{ $t('project.duplicate.text') }}</p>
		<project-search v-model="parentProject"/>
	</create-edit>
</template>

<script setup lang="ts">
import {ref, shallowReactive, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import ProjectDuplicateService from '@/services/projectDuplicateService'
import CreateEdit from '@/components/misc/create-edit.vue'
import ProjectSearch from '@/components/tasks/partials/projectSearch.vue'

import ProjectDuplicateModel from '@/models/projectDuplicateModel'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'
import type {IProject} from '@/modelTypes/IProject'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('project.duplicate.title'))

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()

const projectDuplicateService = shallowReactive(new ProjectDuplicateService())
const parentProject = ref<IProject | null>(null)
watch(
	() => route.params.projectId,
	projectId => {
		const project = projectStore.getProjectById(route.params.projectId)
		if (project.parentProjectId) {
			parentProject.value = projectStore.getProjectById(project.parentProjectId)
		}
	},
	{immediate: true},
)

async function duplicateProject() {
	const projectDuplicate = new ProjectDuplicateModel({
		// FIXME: should be parameter
		projectId: route.params.projectId,
		parentProjectId: parentProject.value.id,
	})

	const duplicate = await projectDuplicateService.create(projectDuplicate)

	projectStore.setProject(duplicate.project)
	success({message: t('project.duplicate.success')})
	router.push({name: 'project.index', params: {projectId: duplicate.project.id}})
}
</script>
