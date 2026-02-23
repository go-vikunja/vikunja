<template>
	<div class="content-widescreen">
		<h2>{{ $t('task.template.manageTitle') }}</h2>
		<p class="has-text-grey">
			{{ $t('task.template.manageDescription') }}
		</p>

		<div
			v-if="loading"
			class="has-text-centered p-4"
		>
			<span class="loader is-loading" />
		</div>

		<div
			v-else-if="templates.length === 0"
			class="has-text-centered has-text-grey p-4"
		>
			{{ $t('task.template.noTemplates') }}
		</div>

		<div
			v-else
			class="template-grid"
		>
			<div
				v-for="tmpl in templates"
				:key="tmpl.id"
				class="template-card"
			>
				<div class="template-card-header">
					<span class="template-card-title">{{ tmpl.title }}</span>
					<div class="template-card-actions">
						<BaseButton
							class="action-btn"
							@click="editTemplate(tmpl)"
						>
							<Icon icon="pen" />
						</BaseButton>
						<BaseButton
							class="action-btn is-danger"
							@click="confirmDelete(tmpl)"
						>
							<Icon icon="trash-alt" />
						</BaseButton>
					</div>
				</div>
				<p
					v-if="tmpl.description"
					class="template-card-description"
				>
					{{ truncate(stripHtml(tmpl.description), 120) }}
				</p>
				<div class="template-card-meta">
					<span
						v-if="tmpl.priority > 0"
						class="tag is-small"
						:class="priorityClass(tmpl.priority)"
					>
						Priority {{ tmpl.priority }}
					</span>
					<span
						v-if="tmpl.labelIds && tmpl.labelIds.length > 0"
						class="tag is-small is-info"
					>
						{{ tmpl.labelIds.length }} label(s)
					</span>
				</div>
			</div>
		</div>

		<!-- Edit modal -->
		<Modal
			:enabled="showEditModal"
			@close="showEditModal = false"
		>
			<div class="edit-template-content">
				<div class="modal-header">
					{{ editingTemplate?.id ? $t('task.template.edit') : $t('task.template.createNew') }}
				</div>
				<div class="content">
					<div class="field">
						<label class="label">{{ $t('task.template.templateName') }}</label>
						<div class="control">
							<input
								v-model="editForm.title"
								v-focus
								class="input"
								type="text"
							>
						</div>
					</div>
					<div class="field">
						<label class="label">{{ $t('task.attributes.description') }}</label>
						<div class="control editor-wrapper">
							<Editor
								v-model="editForm.description"
								:is-edit-enabled="true"
								:show-save="false"
								placeholder="Template description…"
							/>
						</div>
					</div>
					<div class="field">
						<label class="label">{{ $t('task.attributes.priority') }}</label>
						<div class="control">
							<PrioritySelect
								v-model="editForm.priority"
							/>
						</div>
					</div>
					<div class="field">
						<label class="label">{{ $t('task.attributes.color') }}</label>
						<div class="control">
							<ColorPicker v-model="editForm.hexColor" />
						</div>
					</div>
				</div>
				<div class="actions">
					<XButton
						variant="tertiary"
						class="has-text-danger"
						@click="showEditModal = false"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						variant="primary"
						:shadow="false"
						:loading="saving"
						@click="saveTemplate"
					>
						{{ $t('misc.save') }}
					</XButton>
				</div>
			</div>
		</Modal>

		<!-- Delete confirmation -->
		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
		>
			<div class="delete-template-content">
				<div class="modal-header">
					{{ $t('task.template.deleteConfirmTitle') }}
				</div>
				<p>{{ $t('task.template.deleteConfirmText', {title: deletingTemplate?.title}) }}</p>
				<div class="actions">
					<XButton
						variant="tertiary"
						@click="showDeleteModal = false"
					>
						{{ $t('misc.cancel') }}
					</XButton>
					<XButton
						variant="primary"
						class="is-danger"
						:shadow="false"
						:loading="deleting"
						@click="deleteTemplate"
					>
						{{ $t('misc.delete') }}
					</XButton>
				</div>
			</div>
		</Modal>
	</div>
</template>

<script lang="ts" setup>
import {ref, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import Modal from '@/components/misc/Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Editor from '@/components/input/AsyncEditor'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'

import TaskTemplateService from '@/services/taskTemplateService'
import TaskTemplateModel from '@/models/taskTemplate'

import type {ITaskTemplate} from '@/modelTypes/ITaskTemplate'

import {success} from '@/message'

const {t} = useI18n({useScope: 'global'})

const templates = ref<ITaskTemplate[]>([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)

const showEditModal = ref(false)
const showDeleteModal = ref(false)
const editingTemplate = ref<ITaskTemplate | null>(null)
const deletingTemplate = ref<ITaskTemplate | null>(null)

const editForm = ref({
	title: '',
	description: '',
	priority: 0,
	hexColor: '',
})

onMounted(() => loadTemplates())

async function loadTemplates() {
	loading.value = true
	try {
		const service = new TaskTemplateService()
		templates.value = await service.getAll({}, {}, 1, 50) as ITaskTemplate[]
	} catch (e) {
		console.error('Failed to load templates:', e)
	} finally {
		loading.value = false
	}
}

function editTemplate(tmpl: ITaskTemplate) {
	editingTemplate.value = tmpl
	editForm.value = {
		title: tmpl.title,
		description: tmpl.description,
		priority: tmpl.priority,
		hexColor: tmpl.hexColor,
	}
	showEditModal.value = true
}

async function saveTemplate() {
	if (!editingTemplate.value) return
	saving.value = true
	try {
		const service = new TaskTemplateService()
		await service.update(new TaskTemplateModel({
			...editingTemplate.value,
			...editForm.value,
		}))
		success({message: t('task.template.updateSuccess')})
		showEditModal.value = false
		await loadTemplates()
	} catch (e) {
		console.error('Failed to save template:', e)
	} finally {
		saving.value = false
	}
}

function confirmDelete(tmpl: ITaskTemplate) {
	deletingTemplate.value = tmpl
	showDeleteModal.value = true
}

async function deleteTemplate() {
	if (!deletingTemplate.value) return
	deleting.value = true
	try {
		const service = new TaskTemplateService()
		await service.delete(new TaskTemplateModel(deletingTemplate.value))
		success({message: t('task.template.deleteSuccess')})
		showDeleteModal.value = false
		await loadTemplates()
	} catch (e) {
		console.error('Failed to delete template:', e)
	} finally {
		deleting.value = false
	}
}

function stripHtml(html: string): string {
	const doc = new DOMParser().parseFromString(html, 'text/html')
	return doc.body.textContent || ''
}

function truncate(text: string, length: number): string {
	if (text.length <= length) return text
	return text.substring(0, length) + '…'
}

function priorityClass(priority: number): string {
	if (priority >= 4) return 'is-danger'
	if (priority >= 3) return 'is-warning'
	return 'is-info'
}
</script>

<style lang="scss" scoped>
.content-widescreen {
	max-inline-size: 900px;
	margin: 0 auto;
	padding: 1.5rem;
}

.template-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
	gap: 1rem;
}

.template-card {
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	padding: 1rem;
	transition: box-shadow $transition-duration;

	&:hover {
		box-shadow: var(--shadow-sm);
	}
}

.template-card-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
}

.template-card-title {
	font-weight: 600;
	font-size: 1.05rem;
	color: var(--text);
}

.template-card-actions {
	display: flex;
	gap: .25rem;
}

.action-btn {
	padding: .25rem .5rem;
	border-radius: $radius;
	color: var(--grey-500);
	transition: color $transition-duration;

	&:hover {
		color: var(--primary);
	}

	&.is-danger:hover {
		color: var(--danger);
	}
}

.template-card-description {
	color: var(--grey-500);
	font-size: .85rem;
	margin: .5rem 0;
}

.template-card-meta {
	display: flex;
	gap: .5rem;
	margin-block-start: .5rem;
}

.edit-template-content,
.delete-template-content {
	text-align: start;
	padding: 0 1rem;

	.modal-header {
		font-size: 1.75rem;
		font-weight: 700;
		text-align: center;
		margin-block-end: 1rem;
	}

	.actions {
		margin-block-start: 1.5rem;
		text-align: center;

		.button {
			margin: 0 .5rem;
		}
	}
}

.editor-wrapper {
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	padding: .5rem;
	min-block-size: 150px;
}

.field .label {
	color: var(--text);
	font-weight: 600;
}
</style>
