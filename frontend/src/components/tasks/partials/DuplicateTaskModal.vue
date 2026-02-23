<template>
	<Modal
		:enabled="enabled"
		@close="$emit('close')"
	>
		<div class="duplicate-modal-content">
			<div class="modal-header">
				{{ $t('task.duplicate.title') }}
			</div>
			<div class="content">
				<p class="duplicate-info">
					{{ $t('task.duplicate.text', {task: task?.title}) }}
				</p>
				<div class="field">
					<label class="label">{{ $t('task.duplicate.selectProject') }}</label>
					<div class="control">
						<ProjectSearch
							v-model="selectedProject"
							@update:modelValue="handleProjectSelect"
						/>
					</div>
				</div>
				<p class="help">
					{{ $t('task.duplicate.metadataHint') }}
				</p>
				<div
					v-if="errMessage"
					class="notification is-danger is-light mt-2"
				>
					{{ errMessage }}
				</div>
			</div>
			<div class="actions">
				<XButton
					variant="tertiary"
					class="has-text-danger"
					@click="$emit('close')"
				>
					{{ $t('misc.cancel') }}
				</XButton>
				<XButton
					variant="primary"
					:shadow="false"
					:loading="loading"
					:disabled="selectedProject === null || selectedProject.id === 0"
					@click="duplicateTask"
				>
					{{ $t('task.duplicate.action') }}
				</XButton>
			</div>
		</div>
	</Modal>
</template>

<script lang="ts" setup>
import {ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import Modal from '@/components/misc/Modal.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'

import TaskDuplicateService from '@/services/taskDuplicateService'
import TaskDuplicateModel from '@/models/taskDuplicate'

import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'

import {success} from '@/message'

const props = defineProps<{
	enabled: boolean,
	task: ITask | null,
}>()

const emit = defineEmits<{
	'close': [],
	'duplicated': [task: ITask],
}>()

const {t} = useI18n({useScope: 'global'})

const selectedProject = ref<IProject | null>(null)
const loading = ref(false)
const errMessage = ref('')

// Reset state when modal opens
watch(() => props.enabled, (newVal) => {
	if (newVal) {
		selectedProject.value = null
		errMessage.value = ''
	}
})

function handleProjectSelect(project: IProject | null) {
	selectedProject.value = project
	errMessage.value = ''
}

async function duplicateTask() {
	if (!props.task || !selectedProject.value || selectedProject.value.id === 0) {
		errMessage.value = t('task.duplicate.selectProjectRequired')
		return
	}

	loading.value = true
	errMessage.value = ''

	try {
		const taskDuplicateService = new TaskDuplicateService()
		const duplicate = await taskDuplicateService.create(
			new TaskDuplicateModel({
				taskId: props.task.id,
				targetProjectId: selectedProject.value.id,
			}),
		)

		success({message: t('task.duplicate.success')})
		emit('duplicated', duplicate.duplicatedTask!)
		emit('close')
	} catch (e: any) {
		errMessage.value = e?.message || t('task.duplicate.error')
	} finally {
		loading.value = false
	}
}
</script>

<style lang="scss" scoped>
.duplicate-modal-content {
	text-align: start;
	padding: 0 1rem;

	.modal-header {
		font-size: 2rem;
		font-weight: 700;
		text-align: center;
		margin-block-end: 1rem;
	}

	.actions {
		margin-block-start: 1.5rem;
		text-align: center;

		.button {
			margin: 0 0.5rem;
		}
	}
}

.duplicate-info {
	margin-block-end: 1rem;
	color: var(--grey-500);
}

.field .label {
	color: var(--text);
	font-weight: 600;
}

.help {
	color: var(--grey-500);
	font-size: .85rem;
	margin-block-start: .5rem;
}
</style>
