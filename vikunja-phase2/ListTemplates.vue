<template>
	<div class="content-widescreen">
		<h2>{{ $t('task.template.manageTitle') }}</h2>
		<p class="has-text-grey">
			{{ $t('task.template.manageDescription') }}
		</p>

		<!-- Tabs -->
		<div class="template-tabs">
			<BaseButton
				class="template-tab"
				:class="{ 'is-active': activeTab === 'templates' }"
				@click="activeTab = 'templates'"
			>
				<Icon icon="layer-group" />
				{{ $t('task.template.navTitle') }}
			</BaseButton>
			<BaseButton
				class="template-tab"
				:class="{ 'is-active': activeTab === 'chains' }"
				@click="activeTab = 'chains'"
			>
				<Icon icon="link" />
				{{ $t('task.chain.chainsTab') }}
			</BaseButton>
		</div>

		<!-- ═══════ Templates tab ═══════ -->
		<template v-if="activeTab === 'templates'">
			<div class="tab-actions">
				<XButton
					variant="primary"
					icon="plus"
					:shadow="false"
					@click="startCreateTemplate"
				>
					{{ $t('task.template.createNew') }}
				</XButton>
			</div>

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
					:style="tmpl.hexColor ? {borderLeftColor: '#' + tmpl.hexColor, borderLeftWidth: '4px'} : {}"
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
							{{ priorityLabel(tmpl.priority) }}
						</span>
						<span
							v-if="tmpl.percentDone > 0"
							class="tag is-small"
						>
							{{ Math.round(tmpl.percentDone * 100) }}%
						</span>
						<span
							v-if="tmpl.repeatAfter > 0"
							class="tag is-small"
						>
							<Icon icon="history" class="mie-1" />
							Repeats
						</span>
						<span
							v-if="tmpl.hexColor"
							class="tag is-small color-tag"
							:style="{backgroundColor: '#' + tmpl.hexColor}"
						/>
					</div>
				</div>
			</div>
		</template>

		<!-- ═══════ Chains tab ═══════ -->
		<template v-if="activeTab === 'chains'">
			<ChainEditor />
		</template>

		<!-- Edit / Create template modal -->
		<Modal
			:enabled="showEditModal"
			@close="showEditModal = false"
		>
			<div class="edit-template-content">
				<div class="modal-header">
					{{ editingTemplate?.id ? $t('task.template.edit') : $t('task.template.createNew') }}
				</div>
				<div class="edit-form">
					<!-- Title -->
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

					<!-- Description -->
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

					<!-- Two-column row: Priority + Progress -->
					<div class="field-row">
						<div class="field">
							<label class="label">{{ $t('task.attributes.priority') }}</label>
							<div class="control">
								<PrioritySelect v-model="editForm.priority" />
							</div>
						</div>
						<div class="field">
							<label class="label">{{ $t('task.attributes.percentDone') }}</label>
							<div class="control">
								<PercentDoneSelect v-model="editForm.percentDone" />
							</div>
						</div>
					</div>

					<!-- Two-column row: Color + Repeat -->
					<div class="field-row">
						<div class="field">
							<label class="label">{{ $t('task.attributes.color') }}</label>
							<div class="control">
								<ColorPicker v-model="editForm.hexColor" />
							</div>
						</div>
						<div class="field">
							<label class="label">{{ $t('task.attributes.repeat') }}</label>
							<div class="control repeat-control">
								<input
									v-model.number="repeatDays"
									class="input"
									type="number"
									min="0"
									placeholder="0"
								>
								<span class="repeat-unit">days</span>
							</div>
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
						:disabled="!editForm.title"
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
import {ref, computed, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import Modal from '@/components/misc/Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Editor from '@/components/input/AsyncEditor'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import PercentDoneSelect from '@/components/tasks/partials/PercentDoneSelect.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import ChainEditor from '@/components/tasks/partials/ChainEditor.vue'

import TaskTemplateService from '@/services/taskTemplateService'
import TaskTemplateModel from '@/models/taskTemplate'

import type {ITaskTemplate} from '@/modelTypes/ITaskTemplate'

import {success} from '@/message'

const SECONDS_PER_DAY = 86400

const {t} = useI18n({useScope: 'global'})

const activeTab = ref<'templates' | 'chains'>('templates')

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
	percentDone: 0,
	hexColor: '',
	repeatAfter: 0,
	repeatMode: 0,
})

// Convert repeatAfter (seconds) to/from days for the UI
const repeatDays = computed({
	get: () => Math.round(editForm.value.repeatAfter / SECONDS_PER_DAY),
	set: (val: number) => {
		editForm.value.repeatAfter = (val || 0) * SECONDS_PER_DAY
	},
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

function startCreateTemplate() {
	editingTemplate.value = null
	editForm.value = {
		title: '',
		description: '',
		priority: 0,
		percentDone: 0,
		hexColor: '',
		repeatAfter: 0,
		repeatMode: 0,
	}
	showEditModal.value = true
}

function editTemplate(tmpl: ITaskTemplate) {
	editingTemplate.value = tmpl
	editForm.value = {
		title: tmpl.title,
		description: tmpl.description,
		priority: tmpl.priority,
		percentDone: tmpl.percentDone,
		hexColor: tmpl.hexColor,
		repeatAfter: tmpl.repeatAfter || 0,
		repeatMode: tmpl.repeatMode || 0,
	}
	showEditModal.value = true
}

async function saveTemplate() {
	saving.value = true
	try {
		const service = new TaskTemplateService()
		if (editingTemplate.value?.id) {
			// Update existing
			await service.update(new TaskTemplateModel({
				...editingTemplate.value,
				...editForm.value,
			}))
			success({message: t('task.template.updateSuccess')})
		} else {
			// Create new
			await service.create(new TaskTemplateModel({
				...editForm.value,
			}))
			success({message: t('task.template.saveSuccess')})
		}
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

function priorityLabel(priority: number): string {
	const labels: Record<number, string> = {1: 'Low', 2: 'Medium', 3: 'High', 4: 'Urgent', 5: 'DO NOW'}
	return labels[priority] || `Priority ${priority}`
}
</script>

<style lang="scss" scoped>
.content-widescreen {
	max-inline-size: 900px;
	margin: 0 auto;
	padding: 1.5rem;
}

.template-tabs {
	display: flex;
	gap: 0;
	border-block-end: 2px solid var(--grey-200);
	margin-block-end: 1.5rem;
}

.template-tab {
	display: inline-flex;
	align-items: center;
	gap: .4rem;
	padding: .6rem 1.25rem;
	font-weight: 600;
	color: var(--grey-500);
	border-block-end: 2px solid transparent;
	margin-block-end: -2px;
	transition: color $transition-duration, border-color $transition-duration;
	cursor: pointer;

	&:hover {
		color: var(--primary);
	}

	&.is-active {
		color: var(--primary);
		border-block-end-color: var(--primary);
	}
}

.tab-actions {
	margin-block-end: 1.5rem;
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
	flex-wrap: wrap;
	margin-block-start: .5rem;
}

.color-tag {
	inline-size: 1.5rem;
	block-size: 1.25rem;
	border: 1px solid var(--grey-300);
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

.edit-form {
	max-block-size: 70vh;
	overflow-y: auto;
	padding-inline-end: .25rem;
}

.field-row {
	display: flex;
	gap: 1rem;

	.field {
		flex: 1;
	}
}

.editor-wrapper {
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	padding: .5rem;
	min-block-size: 120px;
}

.repeat-control {
	display: flex;
	align-items: center;
	gap: .5rem;

	.input {
		max-inline-size: 80px;
	}
}

.repeat-unit {
	color: var(--grey-500);
	font-size: .9rem;
}

.field .label {
	color: var(--text);
	font-weight: 600;
}
</style>
