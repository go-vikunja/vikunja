<template>
	<CreateEdit
		v-model:loading="loadingModel"
		:title="$t('project.edit.header')"
		primary-icon=""
		:primary-label="$t('misc.save')"
		:tertiary="project.max_permission === PERMISSIONS.ADMIN ? $t('misc.delete') : undefined"
		@primary="save"
		@tertiary="$router.push({ name: 'project.settings.delete', params: { id: projectId } })"
	>
		<FormField
			id="title"
			v-model="project.title"
			v-focus
			:label="$t('project.title')"
			:disabled="isLoading"
			:placeholder="$t('project.edit.titlePlaceholder')"
			type="text"
			@keyup.enter="save"
		/>
		<FormField :label="$t('project.parent')">
			<ProjectSearch v-model="parentProject" />
		</FormField>
		<FormField :label="$t('project.edit.description')">
			<Editor
				id="projectdescription"
				v-model="project.description"
				:class="{ 'disabled': isLoading}"
				:disabled="isLoading"
				:placeholder="$t('project.edit.descriptionPlaceholder')"
			/>
		</FormField>

		<div class="columns">
			<div class="column">
				<FormField
					id="identifier"
					v-model="project.identifier"
					v-tooltip="$t('project.edit.identifierTooltip')"
					:label="$t('project.edit.identifier')"
					:disabled="isLoading"
					:placeholder="$t('project.edit.identifierPlaceholder')"
					type="text"
					maxlength="10"
					@keyup.enter="save"
				/>
			</div>

			<div class="column">
				<FormField :label="$t('project.edit.color')">
					<ColorPicker v-model="project.hex_color" />
				</FormField>
			</div>
		</div>
	</CreateEdit>
</template>

<script setup lang="ts">
import {computed, ref, watch} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'

import Editor from '@/components/input/AsyncEditor'
import ColorPicker from '@/components/input/ColorPicker.vue'
import CreateEdit from '@/components/misc/CreateEdit.vue'
import FormField from '@/components/input/FormField.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'

import type {ProjectResponse} from '@/client/queries/projects'

import {useBaseStore} from '@/stores/base'
import {useProjectNavigation} from '@/composables/useProjectNavigation'
import {useProject} from '@/composables/useProject'

import {useTitle} from '@/composables/useTitle'
import {PERMISSIONS} from '@/constants/permissions'

const props = defineProps<{
	projectId: number,
}>()

defineOptions({name: 'ProjectSettingEdit'})

const router = useRouter()
const projectStore = useProjectNavigation()

const {t} = useI18n({useScope: 'global'})

const {project, save: saveProject, isLoading} = useProject(() => props.projectId)

const parentProject = ref<ProjectResponse | null>(null)
const isSaving = ref(false)

const loadingModel = computed({
	get: () => isSaving.value || isLoading.value,
	set(value: boolean) {
		isSaving.value = value
	},
})
watch(
	() => projectStore.projects[project.value.parent_project_id],
	parent => parentProject.value = parent ?? null,
	{immediate: true},
)

useTitle(() => project.value.title ? t('project.edit.title', {project: project.value.title}) : '')

async function save() {
	if (isSaving.value) {
		return
	}

	isSaving.value = true

	try {
		project.value.parent_project_id = parentProject.value?.id ?? 0
		await saveProject()
		await useBaseStore().handleSetCurrentProject({project: project.value})
		router.back()
	} finally {
		isSaving.value = false
	}
}
</script>
