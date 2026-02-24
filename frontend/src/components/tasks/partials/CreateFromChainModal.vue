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
								@click="selectedChainId = chain.id"
							>
								<Icon
									icon="link"
									class="chain-select-icon"
								/>
								<div class="chain-select-info">
									<span class="chain-select-title">{{ chain.title }}</span>
									<span class="chain-select-meta">{{ chain.steps?.length || 0 }} steps</span>
								</div>
							</BaseButton>
						</div>
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

					<!-- Optional prefix -->
					<div class="field">
						<label class="label">{{ $t('task.chain.titlePrefix') }}</label>
						<div class="control">
							<input
								v-model="titlePrefix"
								class="input"
								type="text"
								:placeholder="$t('task.chain.titlePrefixPlaceholder')"
							>
						</div>
					</div>

					<!-- Step preview with calculated dates -->
					<div
						v-if="selectedChain"
						class="field"
					>
						<label class="label">{{ $t('task.chain.preview') }}</label>
						<div class="step-preview-list">
							<div
								v-for="(step, i) in previewSteps"
								:key="i"
								class="step-preview-item"
							>
								<div class="step-preview-header">
									<span class="step-preview-number">{{ i + 1 }}</span>
									<span class="step-preview-title">{{ titlePrefix }}{{ step.title }}</span>
									<span class="step-preview-date">{{ step.calculatedDate }}</span>
								</div>
								<div
									v-if="step.description"
									class="step-preview-desc"
								>
									<Icon icon="align-left" />
									<span>{{ $t('task.chain.hasDescription') }}</span>
								</div>
								<div class="step-preview-attachments">
									<div
										v-for="(file, fi) in (stepFiles[i] || [])"
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
						:disabled="!selectedChainId || !anchorDate"
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

import {getAllChains, createTasksFromChain} from '@/services/taskChainApi'
import type {ITaskChain} from '@/services/taskChainApi'
import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'

import {success} from '@/message'

const props = defineProps<{
	enabled: boolean
	projectId: number
}>()

const emit = defineEmits<{
	close: []
	created: []
}>()

const {t} = useI18n({useScope: 'global'})

const chains = ref<ITaskChain[]>([])
const loadingChains = ref(false)
const creating = ref(false)

const selectedChainId = ref<number | null>(null)
const anchorDate = ref(new Date().toISOString().split('T')[0])
const titlePrefix = ref('')
const stepFiles = ref<Record<number, File[]>>({})

function handleStepFiles(stepIndex: number, event: Event) {
	const input = event.target as HTMLInputElement
	if (!input.files) return
	if (!stepFiles.value[stepIndex]) {
		stepFiles.value[stepIndex] = []
	}
	stepFiles.value[stepIndex].push(...Array.from(input.files))
	input.value = ''
}

function removeStepFile(stepIndex: number, fileIndex: number) {
	stepFiles.value[stepIndex]?.splice(fileIndex, 1)
}

const selectedChain = computed(() => {
	if (!selectedChainId.value) return null
	return chains.value.find(c => c.id === selectedChainId.value) || null
})

const previewSteps = computed(() => {
	if (!selectedChain.value || !anchorDate.value) return []
	const anchor = new Date(anchorDate.value + 'T00:00:00')
	let cumulativeOffset = 0
	return selectedChain.value.steps.map(step => {
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
		stepFiles.value = {}
	}
})

onMounted(() => {
	if (props.enabled) {
		loadChains()
	}
})

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
	if (!selectedChainId.value || !anchorDate.value) return
	creating.value = true
	try {
		const createdTasks = await createTasksFromChain(selectedChainId.value, {
			target_project_id: props.projectId,
			anchor_date: new Date(anchorDate.value + 'T00:00:00').toISOString(),
			title_prefix: titlePrefix.value,
		})

		// Upload attachments per step to their corresponding created tasks
		if (createdTasks && Array.isArray(createdTasks)) {
			const http = AuthenticatedHTTPFactory()
			for (const [stepIndex, files] of Object.entries(stepFiles.value)) {
				const taskIndex = Number(stepIndex)
				if (taskIndex >= createdTasks.length || !files?.length) continue

				const taskId = createdTasks[taskIndex].id
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
		stepFiles.value = {}
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
}

.step-preview-header {
	display: flex;
	align-items: center;
	gap: .5rem;
}

.step-preview-desc {
	display: flex;
	align-items: center;
	gap: .3rem;
	font-size: .78rem;
	color: var(--grey-400);
	padding-inline-start: 1.75rem;
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

.step-preview-title {
	flex: 1;
	font-size: .85rem;
	color: var(--text);
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
