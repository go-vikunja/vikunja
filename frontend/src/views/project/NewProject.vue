<template>
	<CreateEdit
		v-model:loading="isSubmitting"
		:title="$t('project.create.header')"
		:primary-disabled="project.title === ''"
		@create="createProject()"
	>
		<FormField
			v-model="project.title"
			v-focus
			:label="$t('project.title')"
			:disabled="projectService.loading"
			:placeholder="$t('project.create.titlePlaceholder')"
			type="text"
			name="projectTitle"
			:error="showError && project.title === '' ? $t('project.create.addTitleRequired') : null"
			@keyup.enter="createProject()"
			@keyup.esc="$router.back()"
		/>
		<FormField
			v-if="projectStore.hasProjects"
			:label="$t('project.parent')"
		>
			<ProjectSearch v-model="parentProject" />
		</FormField>
		<FormField :label="$t('project.color')">
			<ColorPicker v-model="project.hexColor" />
		</FormField>
	</CreateEdit>
</template>

<script setup lang="ts">
import {ref, reactive, shallowReactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import FormField from '@/components/input/FormField.vue'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import type {IProject} from '@/modelTypes/IProject'

const props = defineProps<{
	parentProjectId?: number,
}>()

const {t} = useI18n({useScope: 'global'})

useTitle(() => t('project.create.header'))

const showError = ref(false)
const project = reactive(new ProjectModel())
const projectService = shallowReactive(new ProjectService())
const projectStore = useProjectStore()
const parentProject = ref<IProject | null>(null)
const isSubmitting = ref(false)

watch(
	() => props.parentProjectId,
	() => parentProject.value = projectStore.projects[props.parentProjectId],
	{immediate: true},
)

async function createProject() {
	if (project.title === '') {
		showError.value = true
		return
	}
	showError.value = false

	if (isSubmitting.value) {
		return
	}

	isSubmitting.value = true

	if (parentProject.value) {
		project.parentProjectId = parentProject.value.id
	}

	try {
		await projectStore.createProject(project)
		success({message: t('project.create.createdSuccess')})
	} finally {
		isSubmitting.value = false
	}
}
</script>
