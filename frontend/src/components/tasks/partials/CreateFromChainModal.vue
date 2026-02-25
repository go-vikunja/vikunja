<template>
	<Modal
		:enabled="enabled"
		@close="$emit('close')"
	>
		<div class="create-from-chain-content">
			<div class="modal-header">
				{{ $t('task.chain.createFromChain') }}
			</div>

			<div
				v-if="loadingChains"
				class="has-text-centered p-4"
			>
				<span class="loader is-loading" />
			</div>

			<template v-else-if="chains.length === 0">
				<p class="has-text-centered has-text-grey p-4">
					{{ $t('task.chain.noChainsAvailable') }}
				</p>
			</template>

			<template v-else>
				<div class="edit-form">
					<!-- Chain selection -->
					<div class="field">
						<label class="label">{{ $t('task.chain.selectChain') }}</label>
						<div class="chain-select-list">
							<BaseButton
								v-for="chain in chains"
								:key="chain.id"
								class="chain-select-item"
								:class="{ 'is-selected': selectedChainId === chain.id }"
								@click="selectChain(chain)"
							>
								<Icon
									icon="link"
									class="chain-select-icon"
								/>
								<div class="chain-select-info">
									<span class="chain-select-title">{{ chain.title }}</span>
									<span class="chain-select-meta">
										{{ chain.steps?.length || 0 }} steps · {{ formatChainDuration(chain) }}
									</span>
								</div>
							</BaseButton>
						</div>
					</div>

					<!-- Target project -->
					<div class="field">
						<label class="label">Target Project</label>
						<div class="control">
							<select
								v-model="targetProjectId"
								class="input"
								@change="onProjectChanged"
							>
								<option
									v-for="project in availableProjects"
									:key="project.id"
									:value="project.id"
								>
									{{ project.title }}
								</option>
							</select>
						</div>
						<p class="help">Tasks will be created in this project.</p>
					</div>

					<!-- Anchor date -->
					<div class="field">
						<label class="label">{{ $t('task.chain.anchorDate') }}</label>
						<div class="control">
							<input
								v-model="anchorDate"
								class="input"
								type="date"
							>
						</div>
						<p class="help">{{ $t('task.chain.anchorDateHelp') }}</p>
					</div>

					<!-- Title prefix (mandatory, auto-populated) -->
					<div class="field">
						<label class="label">{{ $t('task.chain.titlePrefix') }} <span class="has-text-danger">*</span></label>
						<div class="control">
							<input
								v-model="titlePrefix"
								class="input"
								:class="{ 'is-danger': !titlePrefix.trim() }"
								type="text"
								:placeholder="$t('task.chain.titlePrefixPlaceholder')"
							>
						</div>
						<p class="help">Auto-generated from project name + timestamp. Editable.</p>
					</div>

					<!-- Step preview with calculated dates -->
					<div
						v-if="localSteps.length > 0"
						class="field"
					>
						<label class="label">{{ $t('task.chain.preview') }}</label>
						<div class="step-preview-list">
							<div
								v-for="(step, i) in computedPreview"
								:key="step._key"
								class="step-preview-item"
								draggable="true"
								:class="{ 'is-dragging': dragIndex === i, 'is-drag-over': dragOverIndex === i }"
								@dragstart="onDragStart(i, $event)"
								@dragover.prevent="onDragOver(i)"
								@dragleave="onDragLeave()"
								@drop.prevent="onDrop(i)"
								@dragend="onDragEnd"
							>
								<div class="step-preview-header">
									<span class="step-drag-handle">
										<Icon icon="grip-vertical" />
									</span>
									<span class="step-preview-number">{{ i + 1 }}</span>
									<div class="step-title-group">
										<span
											v-if="computedPrefix"
											class="step-title-prefix"
										>{{ computedPrefix }}</span>
										<input
											v-model="localSteps[i].title"
											class="step-title-edit"
											type="text"
											:placeholder="$t('task.chain.stepTitle')"
										>
									</div>
									<span class="step-preview-date">{{ step.calculatedDate }}</span>
									<BaseButton
										v-if="localSteps.length > 1"
										class="step-remove-btn"
										@click="removeStep(i)"
									>
										<Icon icon="times" />
									</BaseButton>
								</div>
								<!-- Editable description toggle -->
								<BaseButton
									class="step-desc-toggle"
									@click="toggleStepDescription(i)"
								>
									<Icon :icon="expandedDescriptions.has(i) ? 'chevron-up' : 'align-left'" />
									<span v-if="!expandedDescriptions.has(i)">
										{{ step.description ? $t('task.chain.editDescription') : $t('task.chain.addDescription') }}
									</span>
									<span v-else>{{ $t('task.chain.hideDescription') }}</span>
								</BaseButton>
								<div
									v-if="expandedDescriptions.has(i)"
									class="step-desc-editor"
								>
									<Editor
										:model-value="step.description || ''"
										:is-edit-enabled="true"
										:show-save="false"
										:placeholder="$t('task.chain.stepDescriptionPlaceholder')"
										class="step-rich-editor"
										@update:model-value="localSteps[i].description = $event"
									/>
								</div>
								<div class="step-preview-attachments">
									<div
										v-for="(file, fi) in localSteps[i]._files"
										:key="fi"
										class="step-file-tag"
									>
										<Icon icon="paperclip" />
										{{ file.name }}
										<BaseButton
											class="step-file-remove"
											@click="removeStepFile(i, fi)"
										>
											<Icon icon="times" />
										</BaseButton>
									</div>
									<label class="step-file-add">
										<Icon icon="plus" />
										{{ $t('task.chain.addAttachment') }}
										<input
											type="file"
											multiple
											class="hidden-file-input"
											@change="handleStepFiles(i, $event)"
										>
									</label>
								</div>
							</div>
						</div>
						<!-- Add step button -->
						<BaseButton
							class="add-step-btn"
							@click="addStep"
						>
							<Icon icon="plus" />
							{{ $t('task.chain.addStepToPreview') }}
						</BaseButton>
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
						:loading="creating"
						:disabled="!selectedChainId || !anchorDate || localSteps.length === 0 || !titlePrefix.trim()"
						@click="createFromChain"
					>
						{{ $t('task.chain.createTasks') }}
					</XButton>
				</div>
			</template>
		</div>
	</Modal>
</template>

<script lang="ts" setup>
import {ref, computed, onMounted, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import Modal from '@/components/misc/Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import Editor from '@/components/input/AsyncEditor'

import {getAllChains, createTasksFromChain, unitToDays} from '@/services/taskChainApi'
import type {ITaskChain, ITaskChainStep} from '@/services/taskChainApi'
import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {useProjectStore} from '@/stores/projects'

import {success} from '@/message'
import {useDragReorder} from '@/composables/useDragReorder'

interface LocalStep extends ITaskChainStep {
	_key: number // stable key for v-for
	_files: File[] // in-memory file uploads for this step
}

const props = defineProps<{
	enabled: boolean
	projectId: number
}>()

const emit = defineEmits<{
	close: []
	created: []
}>()

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()

const chains = ref<ITaskChain[]>([])
const loadingChains = ref(false)
const creating = ref(false)

const selectedChainId = ref<number | null>(null)
const anchorDate = ref(new Date().toISOString().split('T')[0])
const titlePrefix = ref('')
const targetProjectId = ref(props.projectId)
const expandedDescriptions = ref<Set<number>>(new Set())

// All non-archived projects for the target selector
const availableProjects = computed(() => {
	return Object.values(projectStore.projects)
		.filter((p: any) => !p.isArchived && p.id > 0)
		.sort((a: any, b: any) => a.title.localeCompare(b.title))
})

/**
 * Generate a prefix from the project name + current timestamp.
 * Format: ProjectName_YYMMDD-HHmm
 */
function generatePrefix(projId?: number): string {
	const pid = projId ?? targetProjectId.value
	const project = projectStore.projects[pid]
	const name = project?.title || `P${pid}`
	const now = new Date()
	const yy = String(now.getFullYear()).slice(-2)
	const mm = String(now.getMonth() + 1).padStart(2, '0')
	const dd = String(now.getDate()).padStart(2, '0')
	const hh = String(now.getHours()).padStart(2, '0')
	const min = String(now.getMinutes()).padStart(2, '0')
	return `${name}_${yy}${mm}${dd}-${hh}${min}`
}

/** Re-generate prefix when target project changes */
function onProjectChanged() {
	titlePrefix.value = generatePrefix()
}

/** Format total chain duration for display in selection list */
function formatChainDuration(chain: ITaskChain): string {
	if (!chain.steps || chain.steps.length === 0) return '0d'
	const lastStep = chain.steps.reduce((max, s) => {
		const end = unitToDays(s.offset_days, s.offset_unit) + unitToDays(s.duration_days, s.duration_unit)
		return end > max ? end : max
	}, 0)
	if (lastStep < 1) {
		const hours = Math.round(lastStep * 24)
		return `${hours}h`
	}
	if (lastStep >= 7 && lastStep % 7 === 0) {
		return `${lastStep / 7}w`
	}
	if (lastStep >= 30) {
		return `~${Math.round(lastStep / 30)}mo`
	}
	return `${Math.round(lastStep)}d`
}

// Mutable local copy of steps — initialized from chain, freely editable
const localSteps = ref<LocalStep[]>([])
let nextKey = 0

// Drag-to-reorder for steps in the preview
const {dragIndex, dragOverIndex, onDragStart, onDragOver, onDragLeave, onDrop, onDragEnd} = useDragReorder(localSteps)

function cloneStepsFromChain(chain: ITaskChain) {
	nextKey = 0
	localSteps.value = (chain.steps || []).map(s => ({
		...s,
		description: s.description || '',
		_key: nextKey++,
		_files: [],
	}))
}

function selectChain(chain: ITaskChain) {
	selectedChainId.value = chain.id ?? null
	cloneStepsFromChain(chain)
	expandedDescriptions.value = new Set()
}

/**
 * Mirrors the backend separator logic from task_from_chain.go:
 * If the prefix doesn't end with a separator character, append "_"
 */
const computedPrefix = computed(() => {
	const prefix = titlePrefix.value
	if (!prefix) return ''
	const separators = [' ', '_', '-', ':', '/', '.']
	const lastChar = prefix[prefix.length - 1]
	const sep = separators.includes(lastChar) ? '' : '_'
	return prefix + sep
})

function formatPreviewTitle(stepTitle: string): string {
	return computedPrefix.value + stepTitle
}

function toggleStepDescription(index: number) {
	const s = new Set(expandedDescriptions.value)
	if (s.has(index)) {
		s.delete(index)
	} else {
		s.add(index)
	}
	expandedDescriptions.value = s
}

function removeStep(index: number) {
	if (localSteps.value.length <= 1) return
	localSteps.value.splice(index, 1)
	// Adjust expanded descriptions indices
	const newExpanded = new Set<number>()
	for (const idx of expandedDescriptions.value) {
		if (idx < index) newExpanded.add(idx)
		else if (idx > index) newExpanded.add(idx - 1)
	}
	expandedDescriptions.value = newExpanded
}

function addStep() {
	const last = localSteps.value[localSteps.value.length - 1]
	const nextOffset = last ? last.offset_days + last.duration_days : 0
	localSteps.value.push({
		sequence: localSteps.value.length,
		title: '',
		description: '',
		offset_days: nextOffset,
		duration_days: 1,
		priority: 0,
		hex_color: '',
		label_ids: [],
		_key: nextKey++,
		_files: [],
	})
	// Auto-expand the new step so user can set title
	expandedDescriptions.value = new Set([...expandedDescriptions.value, localSteps.value.length - 1])
}

function handleStepFiles(stepIndex: number, event: Event) {
	const input = event.target as HTMLInputElement
	if (!input.files) return
	localSteps.value[stepIndex]._files.push(...Array.from(input.files))
	input.value = ''
}

function removeStepFile(stepIndex: number, fileIndex: number) {
	localSteps.value[stepIndex]._files.splice(fileIndex, 1)
}

const selectedChain = computed(() => {
	if (!selectedChainId.value) return null
	return chains.value.find(c => c.id === selectedChainId.value) || null
})

const computedPreview = computed(() => {
	if (!anchorDate.value || localSteps.value.length === 0) return []
	const anchor = new Date(anchorDate.value + 'T00:00:00')
	let cumulativeOffset = 0
	return localSteps.value.map(step => {
		cumulativeOffset += step.offset_days
		const date = new Date(anchor)
		date.setDate(date.getDate() + cumulativeOffset)
		return {
			...step,
			calculatedDate: date.toLocaleDateString(undefined, {
				weekday: 'short',
				month: 'short',
				day: 'numeric',
			}),
		}
	})
})

watch(() => props.enabled, (val) => {
	if (val) {
		loadChains()
		resetState()
		// Reset target to current project and auto-populate prefix
		targetProjectId.value = props.projectId
		titlePrefix.value = generatePrefix()
	}
})

onMounted(() => {
	if (props.enabled) {
		loadChains()
	}
})

function resetState() {
	expandedDescriptions.value = new Set()
	if (selectedChain.value) {
		cloneStepsFromChain(selectedChain.value)
	}
}

async function loadChains() {
	loadingChains.value = true
	try {
		chains.value = await getAllChains()
	} catch (e) {
		console.error('Failed to load chains:', e)
	} finally {
		loadingChains.value = false
	}
}

async function createFromChain() {
	if (!selectedChainId.value || !anchorDate.value || localSteps.value.length === 0) return
	creating.value = true
	try {
		// Build custom_steps from localSteps (strip _key)
		const customSteps = localSteps.value.map((s, i) => ({
			sequence: i,
			title: s.title || `Step ${i + 1}`,
			description: s.description || '',
			offset_days: s.offset_days,
			duration_days: s.duration_days,
			priority: s.priority,
			hex_color: s.hex_color,
			label_ids: s.label_ids || [],
		}))

		const createdTasks = await createTasksFromChain(selectedChainId.value, {
			target_project_id: targetProjectId.value,
			anchor_date: new Date(anchorDate.value + 'T00:00:00').toISOString(),
			title_prefix: titlePrefix.value,
			custom_steps: customSteps,
		})

		// Upload attachments per step to their corresponding created tasks
		if (createdTasks && Array.isArray(createdTasks)) {
			const http = AuthenticatedHTTPFactory()
			for (let i = 0; i < localSteps.value.length && i < createdTasks.length; i++) {
				const files = localSteps.value[i]._files
				if (!files?.length) continue

				const taskId = createdTasks[i].id
				for (const file of files) {
					const formData = new FormData()
					formData.append('files', file)
					try {
						await http.put(`/tasks/${taskId}/attachments`, formData, {
							headers: {'Content-Type': 'multipart/form-data'},
						})
					} catch (e) {
						console.error(`Failed to upload ${file.name} to task ${taskId}:`, e)
					}
				}
			}
		}

		success({message: t('task.chain.createTasksSuccess')})
		emit('created')
		emit('close')
	} catch (e) {
		console.error('Failed to create tasks from chain:', e)
	} finally {
		creating.value = false
		resetState()
	}
}
</script>

<style lang="scss" scoped>
.create-from-chain-content {
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
	max-block-size: 65vh;
	overflow-y: auto;
}

.chain-select-list {
	display: flex;
	flex-direction: column;
	gap: .5rem;
}

.chain-select-item {
	display: flex;
	align-items: center;
	gap: .75rem;
	padding: .75rem;
	border: 2px solid var(--grey-200);
	border-radius: $radius;
	cursor: pointer;
	transition: border-color $transition, background $transition;
	text-align: start;

	&:hover {
		border-color: var(--primary);
		background: var(--grey-50);
	}

	&.is-selected {
		border-color: var(--primary);
		background: rgba(var(--primary-rgb, 66, 153, 225), 0.08);
	}
}

.chain-select-icon {
	color: var(--primary);
	font-size: 1.1rem;
}

.chain-select-info {
	display: flex;
	flex-direction: column;
}

.chain-select-title {
	font-weight: 600;
	color: var(--text);
}

.chain-select-meta {
	font-size: .8rem;
	color: var(--grey-400);
}

.step-preview-list {
	display: flex;
	flex-direction: column;
	gap: .35rem;
}

.step-preview-item {
	display: flex;
	flex-direction: column;
	gap: .35rem;
	padding: .5rem;
	border-radius: $radius;
	background: var(--grey-50);
	border: 1px solid transparent;
	transition: opacity 150ms, border-color 150ms, box-shadow 150ms;

	&.is-dragging {
		opacity: .4;
	}

	&.is-drag-over {
		border-color: var(--primary);
		box-shadow: 0 -2px 0 0 var(--primary);
	}
}

.step-drag-handle {
	color: var(--grey-300);
	cursor: grab;
	flex-shrink: 0;
	transition: color $transition-duration;

	&:hover {
		color: var(--grey-500);
	}

	&:active {
		cursor: grabbing;
	}
}

.step-preview-header {
	display: flex;
	align-items: center;
	gap: .5rem;
}

.step-title-group {
	flex: 1;
	display: flex;
	align-items: center;
	gap: 0;
	min-inline-size: 0;
}

.step-title-prefix {
	font-size: .85rem;
	color: var(--grey-400);
	white-space: nowrap;
	flex-shrink: 0;
}

.step-title-edit {
	flex: 1;
	font-size: .85rem;
	font-weight: 500;
	color: var(--text);
	background: transparent;
	border: none;
	border-block-end: 1px solid transparent;
	padding: .1rem .25rem;
	min-inline-size: 60px;
	outline: none;
	transition: border-color $transition-duration;

	&:hover,
	&:focus {
		border-block-end-color: var(--primary);
	}

	&::placeholder {
		color: var(--grey-300);
		font-style: italic;
	}
}

.step-remove-btn {
	color: var(--grey-400);
	padding: .15rem .3rem;
	flex-shrink: 0;
	transition: color $transition-duration;

	&:hover {
		color: var(--danger);
	}
}

.step-desc-toggle {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	font-size: .78rem;
	color: var(--grey-400);
	padding-inline-start: 1.75rem;
	cursor: pointer;
	transition: color $transition-duration;

	&:hover {
		color: var(--primary);
	}
}

.step-desc-editor {
	padding-inline-start: 1.75rem;
}

.step-rich-editor {
	min-block-size: 80px;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	padding: .25rem;
}

.add-step-btn {
	display: inline-flex;
	align-items: center;
	gap: .35rem;
	font-size: .85rem;
	color: var(--primary);
	margin-block-start: .5rem;
	cursor: pointer;
	padding: .35rem .5rem;
	border-radius: $radius;
	transition: background $transition-duration;

	&:hover {
		background: var(--grey-50);
	}
}

.step-preview-attachments {
	display: flex;
	flex-wrap: wrap;
	gap: .35rem;
	padding-inline-start: 1.75rem;
}

.step-file-tag {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	font-size: .78rem;
	background: var(--grey-100);
	padding: .15rem .4rem;
	border-radius: $radius;
}

.step-file-remove {
	color: var(--danger);
	cursor: pointer;
	font-size: .7rem;
}

.step-file-add {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	font-size: .78rem;
	color: var(--primary);
	cursor: pointer;

	&:hover {
		text-decoration: underline;
	}
}

.hidden-file-input {
	display: none;
}

.step-preview-number {
	inline-size: 20px;
	block-size: 20px;
	border-radius: 50%;
	background: var(--primary);
	color: white;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: .7rem;
	font-weight: 700;
	flex-shrink: 0;
}

.step-preview-date {
	font-size: .8rem;
	color: var(--grey-500);
	font-family: monospace;
	white-space: nowrap;
}

.help {
	font-size: .8rem;
	color: var(--grey-400);
	margin-block-start: .25rem;
}

.field .label {
	color: var(--text);
	font-weight: 600;
}
</style>
