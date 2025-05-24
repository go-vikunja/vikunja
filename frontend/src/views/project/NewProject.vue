<template>
	<CreateEdit
		:title="$t('project.create.header')"
		:primary-disabled="project.title === ''"
		@create="createNewProject()"
	>
		<div class="field">
			<label
				class="label"
				for="projectTitle"
			>{{ $t('project.title') }}</label>
			<div
				:class="{ 'is-loading': projectService.loading }"
				class="control"
			>
				<input
					v-model="project.title"
					v-focus
					:class="{ disabled: projectService.loading }"
					class="input"
					:placeholder="$t('project.create.titlePlaceholder')"
					type="text"
					name="projectTitle"
					@keyup.enter="createNewProject()"
					@keyup.esc="$router.back()"
				>
			</div>
		</div>
		<p
			v-if="showError && project.title === ''"
			class="help is-danger"
		>
			{{ $t('project.create.addTitleRequired') }}
		</p>
		<div
			v-if="projectStore.hasProjects"
			class="field"
		>
			<label class="label">{{ $t('project.parent') }}</label>
			<div class="control">
				<ProjectSearch v-model="parentProject" />
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('project.color') }}</label>
			<div class="control">
				<ColorPicker v-model="project.hexColor" />
			</div>
		</div>
	</CreateEdit>
</template>

<script setup lang="ts">
import {ref, reactive, shallowReactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'

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

watch(
	() => props.parentProjectId,
	() => parentProject.value = projectStore.projects[props.parentProjectId],
	{immediate: true},
)

async function createNewProject() {
	if (project.title === '') {
		showError.value = true
		return
	}
	showError.value = false

	if (parentProject.value) {
		project.parentProjectId = parentProject.value.id
	}

	await projectStore.createProject(project)
	success({message: t('project.create.createdSuccess')})
}
</script>
