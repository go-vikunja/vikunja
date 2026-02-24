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
							v-tooltip="$t('task.autoTask.viewLog')"
							class="action-btn"
							@click.stop="openLogModal(tmpl)"
						>
							<Icon icon="history" />
						</BaseButton>
						<BaseButton
							v-tooltip="tmpl.active ? $t('task.autoTask.pause') : $t('task.autoTask.resume')"
							class="action-btn"
							@click.stop="togglePause(tmpl)"
						>
							<Icon :icon="tmpl.active ? 'stop' : 'play'" />
						</BaseButton>
						<BaseButton
							v-tooltip="$t('task.autoTask.sendNow')"
							class="action-btn send-now-btn"
							:disabled="!tmpl.active"
							@click.stop="triggerNow(tmpl)"
						>
							<Icon icon="forward" />
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
						<Icon :icon="['far', 'clock']" class="meta-icon" />
						{{ $t('task.autoTask.every') }}
						{{ tmpl.interval_value }}
						{{ $t('task.autoTask.units.' + tmpl.interval_unit) }}
					</span>
					<span
						v-if="tmpl.project_id"
						class="meta-item"
					>
						<Icon icon="layer-group" class="meta-icon" />
						{{ getProjectTitle(tmpl.project_id) }}
					</span>
					<span
						v-else
						class="meta-item"
					>
						<Icon icon="tasks" class="meta-icon" />
						{{ $t('task.autoTask.defaultProject') }}
					</span>
					<span
						v-if="tmpl.next_due_at"
						class="meta-item"
						:class="{'is-overdue': isOverdue(tmpl.next_due_at)}"
					>
						<Icon :icon="['far', 'calendar-alt']" class="meta-icon" />
						{{ $t('task.autoTask.nextDue') }}: {{ formatDate(tmpl.next_due_at) }}
					</span>
				</div>

				<!-- Log count indicator (click history button to view) -->
				<div
					v-if="tmpl.log && tmpl.log.length > 0"
					class="card-log-hint"
				>
					<Icon icon="history" class="meta-icon" />
					{{ tmpl.log.length }} {{ $t('task.autoTask.generationsRecorded') }}
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

					<!-- Description (rich text / markdown) -->
					<div class="field">
						<label class="label">{{ $t('task.autoTask.description') }}</label>
						<Editor
							v-model="editForm.description"
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
		<!-- Log viewer modal -->
		<Modal
			:enabled="showLogModal"
			@close="showLogModal = false"
			variant="scrolling"
		>
			<Card
				:title="$t('task.autoTask.generationLog') + (logTemplate ? ': ' + logTemplate.title : '')"
				:has-close="true"
				@close="showLogModal = false"
			>
				<div class="log-modal-content">
					<!-- Template summary -->
					<div
						v-if="logTemplate"
						class="log-summary"
					>
						<div class="log-summary-row">
							<span class="log-summary-label">{{ $t('task.autoTask.interval') }}:</span>
							<span>{{ $t('task.autoTask.every') }} {{ logTemplate.interval_value }} {{ $t('task.autoTask.units.' + logTemplate.interval_unit) }}</span>
						</div>
						<div
							v-if="logTemplate.last_created_at"
							class="log-summary-row"
						>
							<span class="log-summary-label">{{ $t('task.autoTask.lastGenerated') }}:</span>
							<span>{{ formatDate(logTemplate.last_created_at) }}</span>
						</div>
						<div
							v-if="logTemplate.last_completed_at"
							class="log-summary-row"
						>
							<span class="log-summary-label">{{ $t('task.autoTask.lastCompleted') }}:</span>
							<span>{{ formatDate(logTemplate.last_completed_at) }}</span>
						</div>
						<div
							v-if="logTemplate.next_due_at"
							class="log-summary-row"
						>
							<span class="log-summary-label">{{ $t('task.autoTask.nextDue') }}:</span>
							<span :class="{'has-text-danger': isOverdue(logTemplate.next_due_at)}">
								{{ formatDate(logTemplate.next_due_at) }}
							</span>
						</div>
						<div class="log-summary-row">
							<span class="log-summary-label">{{ $t('task.autoTask.status') }}:</span>
							<span :class="logTemplate.active ? 'has-text-success' : 'has-text-grey'">
								{{ logTemplate.active ? $t('task.autoTask.active') : $t('task.autoTask.paused') }}
							</span>
						</div>
					</div>

					<!-- Log entries -->
					<div class="log-divider" />
					<div
						v-if="logTemplate?.log?.length > 0"
						class="log-entries-modal"
					>
						<div
							v-for="entry in logTemplate.log"
							:key="entry.id"
							class="log-entry-modal"
							:class="{
								'is-completed-entry': entry.trigger_type === 'completed',
								'is-generation-entry': entry.trigger_type !== 'completed',
							}"
						>
							<div
								class="log-entry-icon"
								:class="{
									'is-completed': entry.trigger_type === 'completed',
									'is-manual': entry.trigger_type === 'manual',
								}"
							>
								<Icon :icon="logEntryIcon(entry)" />
							</div>
							<div class="log-entry-detail">
								<span class="log-entry-type">
									{{ logEntryLabel(entry) }}
								</span>
								<span class="log-entry-task-ref">
									{{ entry.task_title || ('Task #' + entry.task_id) }}
								</span>
								<div class="log-entry-meta">
									<span
										v-if="entry.trigger_type !== 'completed' && entry.task_done && entry.task_done_at"
										class="log-entry-completed"
									>
										<Icon icon="check" class="meta-inline-icon" />
										{{ $t('task.autoTask.completedAt') }} {{ formatDate(entry.task_done_at) }}
									</span>
									<span
										v-else-if="entry.trigger_type !== 'completed' && !entry.task_done"
										class="log-entry-open"
									>
										<Icon :icon="['far', 'clock']" class="meta-inline-icon" />
										{{ $t('task.autoTask.stillOpen') }}
									</span>
									<span
										v-if="entry.task_updated"
										class="log-entry-updated"
									>
										<Icon icon="pen" class="meta-inline-icon" />
										{{ $t('task.autoTask.lastUpdated') }} {{ formatDate(entry.task_updated) }}
									</span>
									<span
										v-if="entry.comment_count > 0"
										class="log-entry-comments"
									>
										<Icon :icon="['far', 'comments']" class="meta-inline-icon" />
										{{ entry.comment_count }} {{ entry.comment_count === 1 ? $t('task.autoTask.comment') : $t('task.autoTask.comments') }}
									</span>
								</div>
							</div>
							<span class="log-entry-date">{{ formatDate(entry.created) }}</span>
						</div>
					</div>
					<p
						v-else
						class="has-text-grey has-text-centered"
					>
						{{ $t('task.autoTask.noLogEntries') }}
					</p>
				</div>
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
import Editor from '@/components/input/AsyncEditor'

import ProjectModel from '@/models/project'
import type {IProject} from '@/modelTypes/IProject'
import type {ILabel} from '@/modelTypes/ILabel'

import {success} from '@/message'
import {useProjectStore} from '@/stores/projects'
import {useLabelStore} from '@/stores/labels'
import {formatDateLong} from '@/helpers/time/formatDate'

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()
const labelStore = useLabelStore()

const loading = ref(false)
const saving = ref(false)
const templates = ref<IAutoTaskTemplate[]>([])
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const editingTemplate = ref<IAutoTaskTemplate | null>(null)
const deletingTemplate = ref<IAutoTaskTemplate | null>(null)
const editForm = ref<IAutoTaskTemplate>(emptyAutoTaskTemplate())
const showLogModal = ref(false)
const logTemplate = ref<IAutoTaskTemplate | null>(null)

// Typed v-model intermediaries for Vikunja components
const selectedProject = ref<IProject>(new ProjectModel())
const selectedLabels = ref<ILabel[]>([])
const editStartDate = ref<Date | null>(new Date())
const editEndDate = ref<Date | null>(null)

// Sync project object ↔ editForm.project_id
watch(selectedProject, (proj) => {
	editForm.value.project_id = proj?.id || 0
})

onMounted(() => {
	loadTemplates()
	labelStore.loadAllLabels()
})

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

function openLogModal(tmpl: IAutoTaskTemplate) {
	logTemplate.value = tmpl
	showLogModal.value = true
}

function logEntryIcon(entry: any): string | string[] {
	switch (entry.trigger_type) {
		case 'completed': return 'check'
		case 'manual': return 'user'
		default: return 'bolt'
	}
}

function logEntryLabel(entry: any): string {
	switch (entry.trigger_type) {
		case 'completed': return t('task.autoTask.logCompleted')
		case 'manual': return t('task.autoTask.logManual')
		default: return t('task.autoTask.logSystem')
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

// Log count hint on card
.card-log-hint {
	margin-block-start: .5rem;
	font-size: .8rem;
	color: var(--grey-400);
	display: flex;
	align-items: center;
	gap: .35rem;
}

// Log viewer modal
.log-modal-content {
	min-inline-size: 400px;
}

.log-summary {
	display: flex;
	flex-direction: column;
	gap: .5rem;
}

.log-summary-row {
	display: flex;
	gap: .75rem;
	font-size: .9rem;
}

.log-summary-label {
	color: var(--grey-500);
	min-inline-size: 120px;
	font-weight: 500;
}

.log-divider {
	border-block-start: 1px solid var(--grey-200);
	margin-block: 1rem;
}

.log-entries-modal {
	display: flex;
	flex-direction: column;
	gap: .5rem;
}

.log-entry-modal {
	display: flex;
	align-items: center;
	gap: .75rem;
	padding: .5rem .75rem;
	border-radius: $radius;
	background: var(--grey-100);
	font-size: .85rem;
}

.log-entry-icon {
	flex-shrink: 0;
	inline-size: 28px;
	block-size: 28px;
	border-radius: 50%;
	background: var(--grey-200);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: .7rem;
	color: var(--grey-500);

	&.is-completed {
		background: var(--success);
		color: var(--white);
	}

	&.is-manual {
		background: var(--primary);
		color: var(--white);
	}
}

.log-entry-detail {
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: .15rem;
}

.log-entry-type {
	font-weight: 500;
	font-size: .8rem;
	color: var(--grey-500);
}

.log-entry-task-ref {
	font-weight: 600;
	font-size: .9rem;
}

.meta-inline-icon {
	font-size: .65rem;
	margin-inline-end: .15rem;
}

.log-entry-meta {
	font-size: .75rem;
	color: var(--grey-400);
	display: flex;
	flex-wrap: wrap;
	gap: .35rem .75rem;
	margin-block-start: .1rem;
}

.log-entry-open {
	color: var(--warning);
}

.log-entry-completed {
	color: var(--success);
}

.log-entry-comments {
	color: var(--grey-500);
}

.log-entry-date {
	color: var(--grey-400);
	white-space: nowrap;
	font-size: .8rem;
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
