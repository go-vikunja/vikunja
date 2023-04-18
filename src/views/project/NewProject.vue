<template>
	<create-edit
		:title="$t('project.create.header')"
		@create="createNewProject()"
		:primary-disabled="project.title === ''"
	>
		<div class="field">
			<label class="label" for="projectTitle">{{ $t('project.title') }}</label>
			<div
				:class="{ 'is-loading': projectService.loading }"
				class="control"
			>
				<input
					:class="{ disabled: projectService.loading }"
					@keyup.enter="createNewProject()"
					@keyup.esc="$router.back()"
					class="input"
					:placeholder="$t('project.create.titlePlaceholder')"
					type="text"
					name="projectTitle"
					v-focus
					v-model="project.title"
				/>
			</div>
		</div>
		<p class="help is-danger" v-if="showError && project.title === ''">
			{{ $t('project.create.addTitleRequired') }}
		</p>
		<div class="field" v-if="projectStore.hasProjects">
			<label class="label">{{ $t('project.parent') }}</label>
			<div class="control">
				<project-search v-model="parentProject"/>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('project.color') }}</label>
			<div class="control">
				<color-picker v-model="project.hexColor"/>
			</div>
		</div>
	</create-edit>
</template>

<script setup lang="ts">
import {ref, reactive, shallowReactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import CreateEdit from '@/components/misc/create-edit.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'
import ProjectSearch from '@/components/tasks/partials/projectSearch.vue'
import type {IProject} from '@/modelTypes/IProject'

const {t} = useI18n({useScope: 'global'})

useTitle(() => t('project.create.header'))

const showError = ref(false)
const project = reactive(new ProjectModel())
const projectService = shallowReactive(new ProjectService())
const projectStore = useProjectStore()
const parentProject = ref<IProject | null>(null)

const props = defineProps<{
	parentProjectId?: number,
}>()

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