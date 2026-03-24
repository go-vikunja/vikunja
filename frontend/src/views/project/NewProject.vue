<template>
	<CreateEdit
		v-model:loading="isSubmitting"
		:title="$t('project.create.header')"
		:primary-disabled="project.title === ''"
		@create="createProject()"
	>
		<FormField
			v-if="projectStore.hasTemplates"
			:label="$t('project.template.useTemplate')"
		>
			<Multiselect
				v-model="selectedTemplate"
				:options="templateOptions"
				:placeholder="$t('project.template.selectTemplate')"
				label="title"
				track-by="id"
			/>
		</FormField>
		<FormField
			v-model="project.title"
			v-focus
			:label="$t('project.title')"
			:disabled="projectService.loading"
			:loading="projectService.loading"
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
import {ref, reactive, shallowReactive, computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import ProjectDuplicateService from '@/services/projectDuplicateService'
import ProjectDuplicateModel from '@/models/projectDuplicateModel'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import FormField from '@/components/input/FormField.vue'
import Multiselect from '@/components/input/Multiselect.vue'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import type {IProject} from '@/modelTypes/IProject'

const props = defineProps<{
	parentProjectId?: number,
}>()

const {t} = useI18n({useScope: 'global'})
const router = useRouter()

useTitle(() => t('project.create.header'))

const showError = ref(false)
const project = reactive(new ProjectModel())
const projectService = shallowReactive(new ProjectService())
const projectStore = useProjectStore()
const parentProject = ref<IProject | null>(null)
const isSubmitting = ref(false)
const selectedTemplate = ref<IProject | null>(null)

const templateOptions = computed(() => projectStore.templateProjects as IProject[])

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
		if (selectedTemplate.value) {
			const duplicateService = new ProjectDuplicateService()
			const duplicate = new ProjectDuplicateModel({
				projectId: selectedTemplate.value.id,
				parentProjectId: project.parentProjectId,
			})
			const response = await duplicateService.create(duplicate)
			const newProject = response.duplicatedProject
			if (newProject) {
				if (project.title !== selectedTemplate.value.title) {
					const updatedProject = await projectService.update({...newProject, title: project.title})
					projectStore.setProject(updatedProject)
				} else {
					projectStore.setProject(newProject)
				}
				router.push({name: 'project.index', params: {projectId: newProject.id}})
			}
			success({message: t('project.create.createdSuccess')})
		} else {
			await projectStore.createProject(project)
			success({message: t('project.create.createdSuccess')})
		}
	} finally {
		isSubmitting.value = false
	}
}
</script>
