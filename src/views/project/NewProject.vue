<template>
	<create-edit :title="$t('project.create.header')" @create="createNewProject()" :primary-disabled="project.title === ''">
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
		<div class="field">
			<label class="label">{{ $t('project.color') }}</label>
			<div class="control">
				<color-picker v-model="project.hexColor" />
			</div>
		</div>
	</create-edit>
</template>

<script setup lang="ts">
import {ref, reactive, shallowReactive} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter, useRoute} from 'vue-router'

import ProjectService from '@/services/project'
import ProjectModel from '@/models/project'
import CreateEdit from '@/components/misc/create-edit.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'

const {t} = useI18n({useScope: 'global'})
const router = useRouter()
const route = useRoute()

useTitle(() => t('project.create.header'))

const showError = ref(false)
const project = reactive(new ProjectModel())
const projectService = shallowReactive(new ProjectService())
const projectStore = useProjectStore()

async function createNewProject() {
	if (project.title === '') {
		showError.value = true
		return
	}
	showError.value = false

	const newProject = await projectStore.createProject(project)
	await router.push({
		name: 'project.index',
		params: { projectId: newProject.id },
	})
	success({message: t('project.create.createdSuccess') })
}
</script>