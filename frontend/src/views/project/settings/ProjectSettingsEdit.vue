<template>
	<CreateEdit
		:title="$t('project.edit.header')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		:tertiary="project.maxPermission === PERMISSIONS.ADMIN ? $t('misc.delete') : undefined"
		@primary="save"
		@tertiary="$router.push({ name: 'project.settings.delete', params: { id: projectId } })"
	>
		<div class="field">
			<label
				class="label"
				for="title"
			>{{ $t('project.title') }}</label>
			<div class="control">
				<input
					id="title"
					v-model="project.title"
					v-focus
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading || undefined"
					class="input"
					:placeholder="$t('project.edit.titlePlaceholder')"
					type="text"
					@keyup.enter="save"
				>
			</div>
		</div>
		<div class="field">
			<label class="label">{{ $t('project.parent') }}</label>
			<div class="control">
				<ProjectSearch v-model="parentProject" />
			</div>
		</div>
		<div class="field">
			<label
				class="label"
				for="projectdescription"
			>{{ $t('project.edit.description') }}</label>
			<div class="control">
				<Editor
					id="projectdescription"
					v-model="project.description"
					:class="{ 'disabled': isLoading}"
					:disabled="isLoading"
					:placeholder="$t('project.edit.descriptionPlaceholder')"
				/>
			</div>
		</div>

		<div class="columns">
			<div class="column field">
				<label
					v-tooltip="$t('project.edit.identifierTooltip')"
					class="label"
					for="identifier"
				>
					{{ $t('project.edit.identifier') }}
				</label>
				<div class="control">
					<input
						id="identifier"
						v-model="project.identifier"
						v-focus
						:class="{ 'disabled': isLoading}"
						:disabled="isLoading || undefined"
						class="input"
						:placeholder="$t('project.edit.identifierPlaceholder')"
						type="text"
						maxlength="10"
						@keyup.enter="save"
					>
				</div>
			</div>

			<div class="column field">
				<label class="label">{{ $t('project.edit.color') }}</label>
				<div class="control">
					<ColorPicker v-model="project.hexColor" />
				</div>
			</div>
		</div>
	</CreateEdit>
</template>

<script setup lang="ts">
import {watch, ref} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import Editor from '@/components/input/AsyncEditor'
import ColorPicker from '@/components/input/ColorPicker.vue'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'

import type {IProject} from '@/modelTypes/IProject'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'
import {useProject} from '@/stores/projects'

import {useTitle} from '@/composables/useTitle'
import {PERMISSIONS} from '@/constants/permissions'

const props = defineProps<{
	projectId: IProject['id'],
}>()

defineOptions({name: 'ProjectSettingEdit'})

const router = useRouter()
const projectStore = useProjectStore()

const {t} = useI18n({useScope: 'global'})

const {project, save: saveProject, isLoading} = useProject(() => props.projectId)

const parentProject = ref<IProject | null>(null)
watch(
	() => project.parentProjectId,
	parentProjectId => {
		if (parentProjectId) {
			parentProject.value = projectStore.projects[parentProjectId]
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
