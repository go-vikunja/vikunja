<template>
	<div>
		<div
			v-if="loading"
			class="has-text-centered p-4"
		>
			<span class="loader is-loading" />
		</div>

		<template v-else>
			<div class="chain-actions">
				<XButton
					variant="primary"
					icon="plus"
					:shadow="false"
					@click="startCreateChain"
				>
					{{ $t('task.chain.createNew') }}
				</XButton>
			</div>

			<div
				v-if="chains.length === 0"
				class="has-text-centered has-text-grey p-4"
			>
				{{ $t('task.chain.noChains') }}
			</div>

			<div
				v-else
				class="chain-grid"
			>
				<div
					v-for="chain in chains"
					:key="chain.id"
					class="chain-card"
				>
					<div class="chain-card-header">
						<span class="chain-card-title">
							<Icon
								icon="link"
								class="chain-icon"
							/>
							{{ chain.title }}
						</span>
						<div class="chain-card-actions">
							<BaseButton
								class="action-btn"
								@click="editChain(chain)"
							>
								<Icon icon="pen" />
							</BaseButton>
							<BaseButton
								class="action-btn is-danger"
								@click="confirmDeleteChain(chain)"
							>
								<Icon icon="trash-alt" />
							</BaseButton>
						</div>
					</div>
					<p
						v-if="chain.description"
						class="chain-card-description"
					>
						{{ truncate(chain.description, 100) }}
					</p>
					<!-- Step timeline preview -->
					<div class="chain-steps-preview">
						<div
							v-for="(step, i) in chain.steps"
							:key="i"
							class="step-preview"
						>
							<span class="step-dot" />
							<span class="step-info">
								<span class="step-title">{{ step.title }}</span>
								<span class="step-offset">
									Day {{ cumulativeDay(chain.steps, i) }}
								</span>
							</span>
						</div>
					</div>
					<div class="chain-timespan">
						<Icon icon="clock" class="timespan-icon" />
						{{ formatTimespan(totalDays(chain.steps)) }}
					</div>
				</div>
			</div>
		</template>

		<!-- Chain Edit Modal -->
		<Modal
			:enabled="showEditModal"
			@close="showEditModal = false"
		>
			<div class="edit-chain-content">
				<div class="modal-header">
					{{ editingChain?.id ? $t('task.chain.edit') : $t('task.chain.createNew') }}
				</div>

				<div class="edit-form">
					<div class="field">
						<label class="label">{{ $t('task.chain.chainName') }}</label>
						<div class="control">
							<input
								v-model="editForm.title"
								v-focus
								class="input"
								type="text"
								:placeholder="$t('task.chain.chainNamePlaceholder')"
							>
						</div>
					</div>

					<div class="field">
						<label class="label">{{ $t('task.attributes.description') }}</label>
						<div class="control">
							<textarea
								v-model="editForm.description"
								class="input textarea"
								rows="2"
								:placeholder="$t('task.chain.descriptionPlaceholder')"
							/>
						</div>
					</div>

					<!-- Steps editor -->
					<div class="field">
						<label class="label">{{ $t('task.chain.steps') }}</label>
						<div class="steps-editor">
							<div
								v-for="(step, i) in editForm.steps"
								:key="i"
								class="step-row"
							>
								<div class="step-number">
									{{ i + 1 }}
								</div>
								<div class="step-fields">
									<input
										v-model="step.title"
										class="input step-title-input"
										type="text"
										:placeholder="$t('task.chain.stepTitle')"
									>
									<div class="step-timing">
										<label class="step-label">{{ i === 0 ? $t('task.chain.offsetDays') : $t('task.chain.daysAfterPrev') }}</label>
										<input
											v-model.number="step.offset_days"
											class="input step-small-input"
											type="number"
											min="0"
										>
										<span class="step-day-indicator">→ Day {{ cumulativeDay(editForm.steps, i) }}</span>
										<label class="step-label">{{ $t('task.chain.durationDays') }}</label>
										<input
											v-model.number="step.duration_days"
											class="input step-small-input"
											type="number"
											min="1"
										>
									</div>
								</div>
								<BaseButton
									class="step-remove-btn"
									@click="removeStep(i)"
								>
									<Icon icon="times" />
								</BaseButton>
							</div>
							<XButton
								variant="secondary"
								icon="plus"
								:shadow="false"
								class="add-step-btn"
								@click="addStep"
							>
								{{ $t('task.chain.addStep') }}
							</XButton>
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
						:disabled="!editForm.title || editForm.steps.length === 0"
						@click="saveChain"
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
			<div class="delete-chain-content">
				<div class="modal-header">
					{{ $t('task.chain.deleteConfirmTitle') }}
				</div>
				<p>{{ $t('task.chain.deleteConfirmText', {title: deletingChain?.title}) }}</p>
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
						@click="deleteChainConfirmed"
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

import {getAllChains, createChain, updateChain, deleteChain as deleteChainApi} from '@/services/taskChainApi'
import type {ITaskChain, ITaskChainStep} from '@/services/taskChainApi'

import {success} from '@/message'

const {t} = useI18n({useScope: 'global'})

const chains = ref<ITaskChain[]>([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)

const showEditModal = ref(false)
const showDeleteModal = ref(false)
const editingChain = ref<ITaskChain | null>(null)
const deletingChain = ref<ITaskChain | null>(null)

function emptyStep(offset = 0): ITaskChainStep {
	return {
		sequence: 0,
		title: '',
		description: '',
		offset_days: offset,
		duration_days: 1,
		priority: 0,
		hex_color: '',
		label_ids: [],
	}
}

const editForm = ref<{
	title: string
	description: string
	steps: ITaskChainStep[]
}>({
	title: '',
	description: '',
	steps: [emptyStep(0)],
})

onMounted(() => loadChains())

async function loadChains() {
	loading.value = true
	try {
		chains.value = await getAllChains()
	} catch (e) {
		console.error('Failed to load chains:', e)
	} finally {
		loading.value = false
	}
}

function startCreateChain() {
	editingChain.value = null
	editForm.value = {
		title: '',
		description: '',
		steps: [emptyStep(0)],
	}
	showEditModal.value = true
}

function editChain(chain: ITaskChain) {
	editingChain.value = chain
	editForm.value = {
		title: chain.title,
		description: chain.description,
		steps: chain.steps.length > 0 ? [...chain.steps] : [emptyStep(0)],
	}
	showEditModal.value = true
}

function addStep() {
	const lastStep = editForm.value.steps[editForm.value.steps.length - 1]
	const nextOffset = lastStep ? lastStep.offset_days + lastStep.duration_days : 0
	editForm.value.steps.push(emptyStep(nextOffset))
}

function removeStep(index: number) {
	if (editForm.value.steps.length <= 1) return
	editForm.value.steps.splice(index, 1)
}

async function saveChain() {
	saving.value = true
	try {
		// Re-sequence steps
		const steps = editForm.value.steps.map((s, i) => ({...s, sequence: i}))

		if (editingChain.value?.id) {
			await updateChain({
				id: editingChain.value.id,
				title: editForm.value.title,
				description: editForm.value.description,
				steps,
			})
			success({message: t('task.chain.updateSuccess')})
		} else {
			await createChain({
				title: editForm.value.title,
				description: editForm.value.description,
				steps,
			})
			success({message: t('task.chain.createSuccess')})
		}
		showEditModal.value = false
		await loadChains()
	} catch (e) {
		console.error('Failed to save chain:', e)
	} finally {
		saving.value = false
	}
}

function confirmDeleteChain(chain: ITaskChain) {
	deletingChain.value = chain
	showDeleteModal.value = true
}

async function deleteChainConfirmed() {
	if (!deletingChain.value?.id) return
	deleting.value = true
	try {
		await deleteChainApi(deletingChain.value.id)
		success({message: t('task.chain.deleteSuccess')})
		showDeleteModal.value = false
		await loadChains()
	} catch (e) {
		console.error('Failed to delete chain:', e)
	} finally {
		deleting.value = false
	}
}

function cumulativeDay(steps: ITaskChainStep[], index: number): number {
	let total = 0
	for (let i = 0; i <= index; i++) {
		total += steps[i]?.offset_days || 0
	}
	return total
}

function totalDays(steps: ITaskChainStep[]): number {
	if (steps.length === 0) return 0
	const lastIndex = steps.length - 1
	return cumulativeDay(steps, lastIndex) + (steps[lastIndex]?.duration_days || 1)
}

function formatTimespan(days: number): string {
	if (days < 1) return '0 days'
	const weeks = Math.floor(days / 7)
	const remainingDays = days % 7
	const parts: string[] = []
	if (weeks > 0) parts.push(`${weeks}w`)
	if (remainingDays > 0) parts.push(`${remainingDays}d`)
	return parts.join(' ') + ` (${days} days)`
}

function truncate(text: string, length: number): string {
	if (text.length <= length) return text
	return text.substring(0, length) + '…'
}
</script>

<style lang="scss" scoped>
.chain-actions {
	margin-block-end: 1.5rem;
}

.chain-grid {
	display: grid;
	grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
	gap: 1rem;
}

.chain-card {
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	padding: 1rem;
	transition: box-shadow $transition-duration;

	&:hover {
		box-shadow: var(--shadow-sm);
	}
}

.chain-card-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
}

.chain-card-title {
	font-weight: 600;
	font-size: 1.05rem;
	color: var(--text);
}

.chain-icon {
	color: var(--primary);
	margin-inline-end: .35rem;
}

.chain-card-actions {
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

.chain-card-description {
	color: var(--grey-500);
	font-size: .85rem;
	margin: .5rem 0;
}

// Step timeline preview
.chain-steps-preview {
	margin-block-start: .75rem;
	display: flex;
	flex-direction: column;
	gap: 0;
}

.step-preview {
	display: flex;
	align-items: center;
	gap: .5rem;
	position: relative;
	padding-block: .25rem;
	padding-inline-start: 1.25rem;

	&::before {
		content: '';
		position: absolute;
		inset-inline-start: .45rem;
		inset-block-start: 0;
		inset-block-end: 0;
		inline-size: 2px;
		background: var(--grey-200);
	}

	&:first-child::before {
		inset-block-start: 50%;
	}

	&:last-child::before {
		inset-block-end: 50%;
	}
}

.step-dot {
	position: absolute;
	inset-inline-start: 0;
	inline-size: 10px;
	block-size: 10px;
	border-radius: 50%;
	background: var(--primary);
	border: 2px solid var(--grey-100);
	z-index: 1;
}

.step-info {
	display: flex;
	align-items: center;
	gap: .5rem;
}

.step-title {
	font-size: .85rem;
	color: var(--text);
}

.step-offset {
	font-size: .75rem;
	color: var(--grey-400);
	font-family: monospace;
}

.chain-timespan {
	display: flex;
	align-items: center;
	gap: .35rem;
	font-size: .8rem;
	color: var(--grey-500);
	margin-block-start: .5rem;
	padding-block-start: .5rem;
	border-block-start: 1px solid var(--grey-200);
}

.edit-timespan {
	margin-block-start: .75rem;
	font-weight: 600;
	color: var(--primary);
}

.timespan-icon {
	font-size: .75rem;
}

.step-day-indicator {
	font-size: .75rem;
	color: var(--primary);
	font-weight: 600;
	white-space: nowrap;
}

// Edit modal
.edit-chain-content,
.delete-chain-content {
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

.steps-editor {
	display: flex;
	flex-direction: column;
	gap: .75rem;
}

.step-row {
	display: flex;
	align-items: flex-start;
	gap: .5rem;
	padding: .75rem;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	background: var(--grey-50);
}

.step-number {
	inline-size: 24px;
	block-size: 24px;
	border-radius: 50%;
	background: var(--primary);
	color: white;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: .75rem;
	font-weight: 700;
	flex-shrink: 0;
	margin-block-start: .35rem;
}

.step-fields {
	flex: 1;
	display: flex;
	flex-direction: column;
	gap: .5rem;
}

.step-title-input {
	font-weight: 500;
}

.step-timing {
	display: flex;
	align-items: center;
	gap: .5rem;
	flex-wrap: wrap;
}

.step-label {
	font-size: .8rem;
	color: var(--grey-500);
	white-space: nowrap;
}

.step-small-input {
	max-inline-size: 70px;
}

.step-remove-btn {
	color: var(--grey-400);
	padding: .25rem;
	margin-block-start: .35rem;

	&:hover {
		color: var(--danger);
	}
}

.add-step-btn {
	align-self: flex-start;
}

.textarea {
	min-block-size: 60px;
	resize: vertical;
}

.field .label {
	color: var(--text);
	font-weight: 600;
}
</style>
