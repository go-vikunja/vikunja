<template>
	<create-edit
		:title="$t('project.edit.header')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		@primary="save"
		:tertiary="$t('misc.delete')"
		@tertiary="$router.push({ name: 'project.settings.delete', params: { id: projectId } })"
	>
		<div class="field">
			<label class="label" for="title">{{ $t('project.title') }}</label>
			<div class="control">
				<input
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading || undefined"
					@keyup.enter="save"
					class="input"
					id="title"
					:placeholder="$t('project.edit.titlePlaceholder')"
					type="text"
					v-focus
					v-model="project.title"/>
			</div>
		</div>
		<div class="field">
			<label
				class="label"
				for="identifier"
				v-tooltip="$t('project.edit.identifierTooltip')">
				{{ $t('project.edit.identifier') }}
			</label>
			<div class="control">
				<input
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading || undefined"
					@keyup.enter="save"
					class="input"
					id="identifier"
					:placeholder="$t('project.edit.identifierPlaceholder')"
					type="text"
					v-focus
					v-model="project.identifier"/>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('project.parent') }}</label>
			<div class="control">
				<project-search v-model="parentProject"/>
			</div>
		</div>
		<div class="field">
			<label class="label" for="projectdescription">{{ $t('project.edit.description') }}</label>
			<div class="control">
				<Editor
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading"
					id="projectdescription"
					:placeholder="$t('project.edit.descriptionPlaceholder')"
					v-model="project.description"
				/>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('project.edit.color') }}</label>
			<div class="control">
				<color-picker v-model="project.hexColor"/>
			</div>
		</div>

	</create-edit>
</template>

<script lang="ts">
export default {name: 'project-setting-edit'}
</script>

<script setup lang="ts">
import {watch, ref, type PropType} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import Editor from '@/components/input/AsyncEditor'
import ColorPicker from '@/components/input/ColorPicker.vue'
import CreateEdit from '@/components/misc/create-edit.vue'
import ProjectSearch from '@/components/tasks/partials/projectSearch.vue'

import type {IProject} from '@/modelTypes/IProject'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useProject} from '@/stores/projects'

import {useTitle} from '@/composables/useTitle'

const props = defineProps({
	projectId: {
		type: Number as PropType<IProject['id']>,
		required: true,
	},
})

const router = useRouter()
const projectStore = useProjectStore()

const {t} = useI18n({useScope: 'global'})

const {project, save: saveProject, isLoading} = useProject(props.projectId)

const parentProject = ref<IProject | null>(null)
watch(
	() => project.parentProjectId,
	projectId => {
		if (project.parentProjectId) {
			parentProject.value = projectStore.projects[project.parentProjectId]
		}
	},
	{immediate: true},
)

useTitle(() => project?.title ? t('project.edit.title', {project: project.title}) : '')

async function save() {
	project.parentProjectId = parentProject.value?.id ?? project.parentProjectId
	await saveProject()
	await useBaseStore().handleSetCurrentProject({project})
	router.back()
}
</script>
