<template>
	<div class="auto-tasks-container">
		<!-- Empty state -->
		<div
			v-if="!loading && templates.length === 0 && !showEditModal"
			class="has-text-centered p-4"
		>
			<p class="has-text-grey mbs-2">
				{{ $t('task.autoTask.noTemplates') }}
			</p>
		</div>

		<!-- Template list -->
		<div
			v-if="templates.length > 0"
			class="auto-task-list"
		>
			<div
				v-for="tmpl in templates"
				:key="tmpl.id"
				class="auto-task-card"
				:class="{'is-paused': !tmpl.active}"
			>
				<div class="card-header-row">
					<div class="card-title-group">
						<span
							class="status-dot"
							:class="tmpl.active ? 'is-active' : 'is-paused'"
							v-tooltip="tmpl.active ? $t('task.autoTask.active') : $t('task.autoTask.paused')"
						/>
						<h4 class="card-title">{{ tmpl.title }}</h4>
					</div>
					<div class="card-actions">
						<BaseButton
							v-tooltip="tmpl.active ? $t('task.autoTask.pause') : $t('task.autoTask.resume')"
							class="action-btn"
							@click.stop="togglePause(tmpl)"
						>
							<Icon :icon="tmpl.active ? 'pause' : 'play'" />
						</BaseButton>
						<BaseButton
							v-tooltip="$t('task.autoTask.sendNow')"
							class="action-btn send-now-btn"
							:disabled="!tmpl.active"
							@click.stop="triggerNow(tmpl)"
						>
							<Icon icon="paper-plane" />
						</BaseButton>
						<BaseButton
							v-tooltip="$t('task.autoTask.edit')"
							class="action-btn"
							@click.stop="editTemplate(tmpl)"
						>
							<Icon icon="pen" />
						</BaseButton>
						<BaseButton
							v-tooltip="$t('task.autoTask.delete')"
							class="action-btn delete-btn"
							@click.stop="confirmDelete(tmpl)"
						>
							<Icon icon="trash-alt" />
						</BaseButton>
					</div>
				</div>

				<div class="card-meta">
					<span class="meta-item">
						<Icon icon="clock" class="meta-icon" />
						{{ $t('task.autoTask.every') }}
						{{ tmpl.interval_value }}
						{{ $t('task.autoTask.units.' + tmpl.interval_unit) }}
					</span>
					<span
						v-if="tmpl.project_id"
						class="meta-item"
					>
						<Icon icon="folder" class="meta-icon" />
						{{ getProjectTitle(tmpl.project_id) }}
					</span>
					<span
						v-else
						class="meta-item"
					>
						<Icon icon="inbox" class="meta-icon" />
						{{ $t('task.autoTask.defaultProject') }}
					</span>
					<span
						v-if="tmpl.next_due_at"
						class="meta-item"
						:class="{'is-overdue': isOverdue(tmpl.next_due_at)}"
					>
						<Icon icon="calendar" class="meta-icon" />
						{{ $t('task.autoTask.nextDue') }}: {{ formatDate(tmpl.next_due_at) }}
					</span>
				</div>

				<!-- Generation log (collapsible) -->
				<div
					v-if="tmpl.log && tmpl.log.length > 0"
					class="card-log"
				>
					<BaseButton
						class="log-toggle"
						@click="toggleLog(tmpl.id)"
					>
						<Icon :icon="expandedLogs.has(tmpl.id) ? 'chevron-up' : 'chevron-down'" />
						{{ $t('task.autoTask.generationLog') }}
						({{ tmpl.log.length }})
					</BaseButton>
					<div
						v-if="expandedLogs.has(tmpl.id)"
						class="log-entries"
					>
						<div
							v-for="entry in tmpl.log"
							:key="entry.id"
							class="log-entry"
						>
							<Icon
								:icon="entry.trigger_type === 'manual' ? 'user' : 'robot'"
								class="log-icon"
							/>
							<span class="log-text">
								{{ entry.trigger_type === 'manual'
									? $t('task.autoTask.logManual')
									: $t('task.autoTask.logSystem') }}
							</span>
							<span class="log-date">{{ formatDate(entry.created) }}</span>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Edit/Create Modal -->
		<Modal
			:enabled="showEditModal"
			@close="closeModal"
			variant="scrolling"
		>
			<Card
				class="auto-task-modal"
				:title="editingTemplate ? $t('task.autoTask.editTemplate') : $t('task.autoTask.createNew')"
				:has-close="true"
				@close="closeModal"
			>
				<div class="modal-form">
					<!-- Title -->
					<div class="field">
						<label class="label">{{ $t('task.autoTask.taskTitle') }}</label>
						<input
							v-model="editForm.title"
							class="input"
							type="text"
							:placeholder="$t('task.autoTask.taskTitlePlaceholder')"
							@keyup.enter="saveTemplate"
						>
					</div>

					<!-- Description -->
					<div class="field">
						<label class="label">{{ $t('task.autoTask.description') }}</label>
						<textarea
							v-model="editForm.description"
							class="textarea"
							rows="3"
							:placeholder="$t('task.autoTask.descriptionPlaceholder')"
						/>
					</div>

					<!-- Interval row -->
					<div class="field">
						<label class="label">{{ $t('task.autoTask.interval') }}</label>
						<div class="interval-row">
							<span class="interval-label">{{ $t('task.autoTask.every') }}</span>
							<input
								v-model.number="editForm.interval_value"
								class="input interval-input"
								type="number"
								min="1"
							>
							<select
								v-model="editForm.interval_unit"
								class="input interval-select"
							>
								<option value="hours">{{ $t('task.autoTask.units.hours') }}</option>
								<option value="days">{{ $t('task.autoTask.units.days') }}</option>
								<option value="weeks">{{ $t('task.autoTask.units.weeks') }}</option>
								<option value="months">{{ $t('task.autoTask.units.months') }}</option>
							</select>
						</div>
					</div>

					<!-- Project -->
					<div class="field">
						<label class="label">{{ $t('task.autoTask.targetProject') }}</label>
						<ProjectSearch v-model="selectedProject" />
						<p class="help">{{ $t('task.autoTask.targetProjectHelp') }}</p>
					</div>

					<!-- Priority -->
					<div class="field">
						<label class="label">{{ $t('task.attributes.priority') }}</label>
						<PrioritySelect v-model="editForm.priority" />
					</div>

					<!-- Labels -->
					<div class="field">
						<label class="label">{{ $t('task.attributes.labels') }}</label>
						<EditLabels
							v-model="selectedLabels"
							:creatable="false"
						/>
					</div>

					<!-- Start date -->
					<div class="field">
						<label class="label">{{ $t('task.autoTask.startDate') }}</label>
						<Datepicker v-model="editStartDate" />
					</div>

					<!-- End date (optional) -->
					<div class="field">
						<label class="label">{{ $t('task.autoTask.endDate') }}</label>
						<Datepicker v-model="editEndDate" />
						<p class="help">{{ $t('task.autoTask.endDateHelp') }}</p>
					</div>

					<!-- Active toggle -->
					<div class="field">
						<FancyCheckbox
							v-model="editForm.active"
						>
							{{ $t('task.autoTask.activeLabel') }}
						</FancyCheckbox>
					</div>
				</div>

				<template #footer>
					<div class="modal-footer">
						<XButton
							variant="secondary"
							:shadow="false"
							@click="closeModal"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							variant="primary"
							:shadow="false"
							:loading="saving"
							@click="saveTemplate"
						>
							{{ editingTemplate ? $t('misc.save') : $t('task.autoTask.create') }}
						</XButton>
					</div>
				</template>
			</Card>
		</Modal>

		<!-- Delete confirmation -->
		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
		>
			<Card
				:title="$t('task.autoTask.deleteConfirmTitle')"
				:has-close="true"
				@close="showDeleteModal = false"
			>
				<p>{{ $t('task.autoTask.deleteConfirmText', {title: deletingTemplate?.title || ''}) }}</p>
				<template #footer>
					<div class="modal-footer">
						<XButton
							variant="secondary"
							:shadow="false"
							@click="showDeleteModal = false"
						>
							{{ $t('misc.cancel') }}
						</XButton>
						<XButton
							variant="danger"
							:shadow="false"
							@click="doDelete"
						>
							{{ $t('misc.delete') }}
						</XButton>
					</div>
				</template>
			</Card>
		</Modal>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import {
	getAllAutoTasks,
	createAutoTask,
	updateAutoTask,
	deleteAutoTask as deleteAutoTaskApi,
	triggerAutoTask,
	emptyAutoTaskTemplate,
} from '@/services/autoTaskApi'
import type {IAutoTaskTemplate} from '@/services/autoTaskApi'

import BaseButton from '@/components/base/BaseButton.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import Modal from '@/components/misc/Modal.vue'
import Card from '@/components/misc/Card.vue'
import Datepicker from '@/components/input/Datepicker.vue'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'

import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'
import type {ILabel} from '@/modelTypes/ILabel'

import {success} from '@/message'
import {useProjectStore} from '@/stores/projects'
import {formatDateLong} from '@/helpers/time/formatDate'

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()

const loading = ref(false)
const saving = ref(false)
const templates = ref<IAutoTaskTemplate[]>([])
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const editingTemplate = ref<IAutoTaskTemplate | null>(null)
const deletingTemplate = ref<IAutoTaskTemplate | null>(null)
const editForm = ref<IAutoTaskTemplate>(emptyAutoTaskTemplate())
const expandedLogs = ref<Set<number>>(new Set())

// Typed v-model intermediaries for Vikunja components
const selectedProject = ref<IProject>(new ProjectModel())
const selectedLabels = ref<ILabel[]>([])
const editStartDate = ref<Date | null>(new Date())
const editEndDate = ref<Date | null>(null)

// Sync project object ↔ editForm.project_id
watch(selectedProject, (proj) => {
	editForm.value.project_id = proj?.id || 0
})

onMounted(loadTemplates)

async function loadTemplates() {
	loading.value = true
	try {
		templates.value = await getAllAutoTasks()
	} finally {
		loading.value = false
	}
}

function getProjectTitle(projectId: number): string {
	return projectStore.projects[projectId]?.title || `Project #${projectId}`
}

function formatDate(dateStr: string | null): string {
	if (!dateStr) return '—'
	return formatDateLong(new Date(dateStr))
}

function isOverdue(dateStr: string | null): boolean {
	if (!dateStr) return false
	return new Date(dateStr) < new Date()
}

function toggleLog(id: number) {
	if (expandedLogs.value.has(id)) {
		expandedLogs.value.delete(id)
	} else {
		expandedLogs.value.add(id)
	}
}

// --- CRUD ---

function openCreate() {
	editingTemplate.value = null
	editForm.value = emptyAutoTaskTemplate()
	selectedProject.value = new ProjectModel()
	selectedLabels.value = []
	editStartDate.value = new Date()
	editEndDate.value = null
	showEditModal.value = true
}

function editTemplate(tmpl: IAutoTaskTemplate) {
	editingTemplate.value = tmpl
	editForm.value = {...tmpl}

	// Load project object from store
	if (tmpl.project_id && projectStore.projects[tmpl.project_id]) {
		selectedProject.value = projectStore.projects[tmpl.project_id]
	} else {
		selectedProject.value = new ProjectModel()
	}

	// Labels: convert IDs to label objects from the store (best effort)
	selectedLabels.value = []

	// Dates
	editStartDate.value = tmpl.start_date ? new Date(tmpl.start_date) : new Date()
	editEndDate.value = tmpl.end_date ? new Date(tmpl.end_date) : null

	showEditModal.value = true
}

function closeModal() {
	showEditModal.value = false
	editingTemplate.value = null
}

async function saveTemplate() {
	if (!editForm.value.title.trim()) return

	// Sync typed values back to the flat form
	editForm.value.project_id = selectedProject.value?.id || 0
	editForm.value.label_ids = selectedLabels.value.map(l => l.id)
	editForm.value.start_date = editStartDate.value
		? editStartDate.value.toISOString()
		: new Date().toISOString()
	editForm.value.end_date = editEndDate.value
		? editEndDate.value.toISOString()
		: null

	saving.value = true
	try {
		if (editingTemplate.value) {
			await updateAutoTask(editForm.value)
			success({message: t('task.autoTask.updateSuccess')})
		} else {
			await createAutoTask(editForm.value)
			success({message: t('task.autoTask.createSuccess')})
		}
		closeModal()
		await loadTemplates()
	} finally {
		saving.value = false
	}
}

async function togglePause(tmpl: IAutoTaskTemplate) {
	const updated = {...tmpl, active: !tmpl.active}
	await updateAutoTask(updated)
	success({
		message: updated.active
			? t('task.autoTask.resumed')
			: t('task.autoTask.pausedSuccess'),
	})
	await loadTemplates()
}

async function triggerNow(tmpl: IAutoTaskTemplate) {
	try {
		await triggerAutoTask(tmpl.id!)
		success({message: t('task.autoTask.triggeredSuccess')})
		await loadTemplates()
	} catch (e: any) {
		if (e?.response?.data?.message || e?.message) {
			success({message: e?.response?.data?.message || e.message})
		}
	}
}

function confirmDelete(tmpl: IAutoTaskTemplate) {
	deletingTemplate.value = tmpl
	showDeleteModal.value = true
}

async function doDelete() {
	if (!deletingTemplate.value?.id) return
	await deleteAutoTaskApi(deletingTemplate.value.id)
	success({message: t('task.autoTask.deleteSuccess')})
	showDeleteModal.value = false
	deletingTemplate.value = null
	await loadTemplates()
}

// Expose for parent (ListTemplates)
defineExpose({openCreate})
</script>

<style lang="scss" scoped>
.auto-tasks-container {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.auto-task-list {
	display: flex;
	flex-direction: column;
	gap: .75rem;
}

.auto-task-card {
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	padding: 1rem;
	transition: border-color $transition, opacity $transition;

	&.is-paused {
		opacity: 0.6;
		border-style: dashed;
	}

	&:hover {
		border-color: var(--grey-300);
	}
}

.card-header-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: .5rem;
}

.card-title-group {
	display: flex;
	align-items: center;
	gap: .5rem;
	min-inline-size: 0;
}

.status-dot {
	flex-shrink: 0;
	inline-size: 10px;
	block-size: 10px;
	border-radius: 50%;

	&.is-active {
		background: var(--success);
	}

	&.is-paused {
		background: var(--grey-400);
	}
}

.card-title {
	font-size: 1rem;
	font-weight: 600;
	margin: 0;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.card-actions {
	display: flex;
	gap: .25rem;
	flex-shrink: 0;
}

.action-btn {
	padding: .25rem .4rem;
	border-radius: $radius;
	color: var(--grey-500);
	transition: color $transition, background $transition;

	&:hover {
		color: var(--primary);
		background: var(--grey-100);
	}
}

.send-now-btn:hover {
	color: var(--success);
}

.delete-btn:hover {
	color: var(--danger);
}

.card-meta {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem 1.25rem;
	margin-block-start: .5rem;
	font-size: .85rem;
	color: var(--grey-500);
}

.meta-item {
	display: inline-flex;
	align-items: center;
	gap: .3rem;

	&.is-overdue {
		color: var(--danger);
		font-weight: 600;
	}
}

.meta-icon {
	font-size: .75rem;
}

// Generation log
.card-log {
	margin-block-start: .75rem;
	border-block-start: 1px solid var(--grey-200);
	padding-block-start: .5rem;
}

.log-toggle {
	font-size: .8rem;
	color: var(--grey-500);
	display: flex;
	align-items: center;
	gap: .35rem;

	&:hover {
		color: var(--grey-700);
	}
}

.log-entries {
	margin-block-start: .5rem;
	display: flex;
	flex-direction: column;
	gap: .25rem;
}

.log-entry {
	display: flex;
	align-items: center;
	gap: .5rem;
	font-size: .8rem;
	color: var(--grey-500);
	padding: .15rem 0;
}

.log-icon {
	font-size: .7rem;
	flex-shrink: 0;
}

.log-text {
	flex: 1;
}

.log-date {
	color: var(--grey-400);
	white-space: nowrap;
}

// Modal form
.modal-form {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.interval-row {
	display: flex;
	align-items: center;
	gap: .5rem;
}

.interval-label {
	color: var(--grey-500);
	white-space: nowrap;
}

.interval-input {
	max-inline-size: 80px;
}

.interval-select {
	max-inline-size: 120px;
}

.modal-footer {
	display: flex;
	justify-content: flex-end;
	gap: .5rem;
}

.help {
	font-size: .8rem;
	color: var(--grey-400);
	margin-block-start: .25rem;
}
</style>
