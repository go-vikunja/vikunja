<template>
	<div
		:class="{ 'is-loading': loading}"
		class="loader-container"
	>
		<XButton
			:to="{name:'labels.create'}"
			class="is-pulled-right"
			icon="plus"
		>
			{{ $t('label.create.header') }}
		</XButton>

		<div class="content">
			<h1>{{ $t('label.manage') }}</h1>
			<p v-if="labelStore.labelsArray.length > 0">
				{{ $t('label.description') }}
			</p>
			<p
				v-else
				class="has-text-centered has-text-grey is-italic"
			>
				{{ $t('label.newCTA') }}
				<RouterLink :to="{name:'labels.create'}">
					{{ $t('label.create.title') }}.
				</RouterLink>
			</p>
		</div>

		<div class="columns">
			<div class="labels-list column">
				<div
					v-for="label in labelStore.labelsArray"
					:key="label.id"
					:style="getLabelStyles(label)"
					class="tag label-row"
				>
					<button
						type="button"
						class="label-row-link"
						:class="{'label-row-link--editable': userInfo.id === label.createdBy.id}"
						@click="handleLabelTitleClick(label)"
					>
						<span>{{ label.title }}</span>
					</button>
					<button
						v-if="userInfo.id === label.createdBy.id"
						type="button"
						class="label-edit-button is-small"
						aria-label="Edit label"
						@click.stop.prevent="editLabel(label)"
					>
						<Icon
							icon="pen"
							class="icon"
						/>
					</button>
				</div>
			</div>

			<Teleport to="body">
				<div
					v-if="isLabelEdit"
					class="vk-overlay"
					@click.self="isLabelEdit = false"
				>
					<div
						class="vk-modal"
						role="dialog"
						aria-modal="true"
						aria-labelledby="edit-label-modal-title"
					>
						<div class="vk-modal-head">
							<h2
								id="edit-label-modal-title"
								class="vk-modal-title"
							>
								{{ $t('label.edit.header') }}
							</h2>
							<button
								class="vk-modal-close"
								aria-label="Close"
								@click="isLabelEdit = false"
							>
								<svg
									width="14"
									height="14"
									viewBox="0 0 16 16"
									fill="none"
									stroke="currentColor"
									stroke-width="1.5"
									stroke-linecap="round"
								>
									<path d="M3 3l10 10M13 3L3 13" />
								</svg>
							</button>
						</div>

						<form @submit.prevent="editLabelSubmit()">
							<div class="vk-modal-body">
								<div class="vk-field">
									<label
										class="vk-label"
										for="edit-label-title"
									>
										{{ $t('label.attributes.title') }}
									</label>
									<input
										id="edit-label-title"
										v-model="labelEditLabel.title"
										class="vk-input"
										type="text"
										:placeholder="$t('label.attributes.titlePlaceholder')"
									/>
								</div>

								<div class="vk-field">
									<label class="vk-label">{{ $t('label.attributes.description') }}</label>
									<textarea
										v-model="labelEditLabel.description"
										class="vk-input vk-textarea"
										:placeholder="$t('label.attributes.description')"
									/>
								</div>

								<div class="vk-field">
									<label class="vk-label">{{ $t('label.attributes.color') }}</label>
									<div class="vk-swatches">
										<button
											v-for="color in colors"
											:key="color"
											type="button"
											class="vk-swatch"
											:class="{'vk-swatch--active': selectedEditColor === color}"
											:style="{background: color}"
											:aria-label="`Select color ${color}`"
											@click="selectedEditColor = color"
										/>
									</div>
								</div>
							</div>

							<div class="vk-modal-foot">
								<button
									type="button"
									class="vk-btn-delete"
									@click="showDeleteDialoge(labelEditLabel)"
								>
									<Icon icon="trash-alt" />
								</button>
								<button
									type="button"
									class="vk-btn-cancel"
									@click="isLabelEdit = false"
								>
									{{ $t('misc.cancel') }}
								</button>
								<button
									type="submit"
									class="vk-btn-primary"
									:disabled="labelEditLabel.title.trim() === ''"
								>
									{{ $t('misc.save') }}
								</button>
							</div>
						</form>
					</div>
				</div>
			</Teleport>

			<Modal
				:enabled="showDeleteModal"
				@close="showDeleteModal = false"
				@submit="deleteLabel(labelToDelete)"
			>
				<template #header>
					<span>{{ $t('task.label.delete.header') }}</span>
				</template>

				<template #text>
					<p>
						{{ $t('task.label.delete.text1') }}<br>
						{{ $t('task.label.delete.text2') }}
					</p>
				</template>
			</Modal>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import {useI18n} from 'vue-i18n'

import LabelModel from '@/models/label'
import type {ILabel} from '@/modelTypes/ILabel'
import {useAuthStore} from '@/stores/auth'
import {useLabelStore} from '@/stores/labels'

import { useTitle } from '@/composables/useTitle'
import {useLabelStyles} from '@/composables/useLabelStyles'

const {t} = useI18n({useScope: 'global'})

const labelEditLabel = ref<ILabel>(new LabelModel())
const isLabelEdit = ref(false)
const showDeleteModal = ref(false)
const labelToDelete = ref<ILabel | undefined>(undefined)
const selectedEditColor = ref('#6c63f5')

const colors = [
	'#6c63f5', '#10b981', '#f59e0b', '#ef4444',
	'#ec4899', '#06b6d4', '#8b5cf6', '#f97316',
]

useTitle(() => t('label.title'))

const authStore = useAuthStore()
const userInfo = computed(() => authStore.info)

const labelStore = useLabelStore()
labelStore.loadAllLabels()

const loading = computed(() => labelStore.isLoading)
const {getLabelStyles} = useLabelStyles()

function deleteLabel(label?: ILabel) {
	if (!label) {
		return
	}

	showDeleteModal.value = false
	isLabelEdit.value = false
	return labelStore.deleteLabel(label)
}

function editLabelSubmit() {
	labelEditLabel.value.hexColor = selectedEditColor.value
	return labelStore.updateLabel(labelEditLabel.value)
}

function editLabel(label: ILabel) {
	if (label.createdBy.id !== userInfo.value.id) {
		return
	}
	// Duplicating the label to make sure it does not look like changes take effect immediatly as the label 
	// object passed to this function here still has a reference to the store.
	labelEditLabel.value = new LabelModel({
		...label,
		// The model does not support passing dates into it directly so we need to convert them first				
		created: +label.created,
		updated: +label.updated,
	})
	selectedEditColor.value = labelEditLabel.value.hexColor || colors[0]
	isLabelEdit.value = true
}

function showDeleteDialoge(label: ILabel) {
	labelToDelete.value = label
	showDeleteModal.value = true
}

function handleLabelTitleClick(label: ILabel) {
	if (label.createdBy.id === userInfo.value.id) {
		editLabel(label)
	}
}
</script>

<style lang="scss" scoped>
.vk-overlay {
	position: fixed;
	inset: 0;
	background: rgba(0, 0, 0, 0.6);
	display: flex;
	align-items: center;
	justify-content: center;
	z-index: 1000;
	padding: 24px;
}

.vk-modal {
	width: 100%;
	max-width: 520px;
	background: var(--vk-bg-panel);
	border: 0.5px solid var(--vk-border-mid);
	border-radius: 14px;
	overflow: hidden;
	font-family: 'DM Sans', sans-serif;
	box-shadow: 0 22px 70px rgba(0, 0, 0, 0.55);
}

.vk-modal-head {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 18px 20px;
	border-bottom: 0.5px solid var(--vk-border);
}

.vk-modal-title {
	font-family: 'Playfair Display', serif;
	font-size: 16px;
	font-weight: 400;
	color: var(--vk-text-primary);
	margin: 0;
	padding: 0;
	border: none;
}

.vk-modal-close {
	width: 28px;
	height: 28px;
	border-radius: 7px;
	background: transparent;
	border: none;
	display: flex;
	align-items: center;
	justify-content: center;
	cursor: pointer;
	color: var(--vk-text-muted);
	transition: all 0.12s;
}

.vk-modal-close:hover {
	background: var(--vk-bg-hover);
	color: var(--vk-text-secondary);
}

.vk-modal-body {
	padding: 20px;
}

.vk-modal-foot {
	display: flex;
	align-items: center;
	justify-content: flex-end;
	gap: 8px;
	padding: 14px 20px;
	border-top: 0.5px solid var(--vk-border);
}

.vk-field {
	margin-bottom: 18px;
}

.vk-field:last-child {
	margin-bottom: 0;
}

.vk-label {
	display: block;
	font-size: 11px;
	font-weight: 500;
	color: var(--vk-text-secondary);
	letter-spacing: 0.04em;
	text-transform: uppercase;
	margin-bottom: 7px;
}

.vk-input {
	width: 100%;
	background: var(--vk-bg);
	border: 0.5px solid var(--vk-border-mid);
	border-radius: 8px;
	padding: 10px 13px;
	font-size: 13px;
	color: var(--vk-text-secondary);
	font-family: 'DM Sans', sans-serif;
	outline: none;
	transition: border-color 0.15s;
	box-shadow: none;
	height: auto;
}

.vk-input::placeholder {
	color: var(--vk-input-placeholder);
}

.vk-input:focus {
	border-color: var(--vk-accent);
	box-shadow: none;
}

.vk-textarea {
	min-height: 110px;
	resize: vertical;
}

.vk-swatches {
	display: flex;
	gap: 8px;
	flex-wrap: wrap;
}

.vk-swatch {
	width: 26px;
	height: 26px;
	border-radius: 6px;
	cursor: pointer;
	border: 2px solid transparent;
	transition: all 0.15s;
	outline: none;
}

.vk-swatch:hover {
	transform: scale(1.12);
}

.vk-swatch--active {
	border-color: #fff;
	transform: scale(1.08);
}

.vk-btn-delete {
	inline-size: 34px;
	block-size: 34px;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	background: rgba(239, 68, 68, 0.15);
	border: 0.5px solid rgba(239, 68, 68, 0.35);
	color: #ef4444;
	border-radius: 8px;
	cursor: pointer;
	margin-inline-end: auto;
	transition: background 0.12s;
}

.vk-btn-delete:hover {
	background: rgba(239, 68, 68, 0.25);
}

.vk-btn-cancel {
	background: transparent;
	border: 0.5px solid var(--vk-border-mid);
	border-radius: 8px;
	padding: 7px 16px;
	font-size: 12.5px;
	color: var(--vk-text-secondary);
	font-family: 'DM Sans', sans-serif;
	cursor: pointer;
	transition: all 0.12s;
}

.vk-btn-cancel:hover {
	background: var(--vk-bg-hover);
	color: var(--vk-text-secondary);
}

.vk-btn-primary {
	display: inline-flex;
	align-items: center;
	gap: 6px;
	background: var(--vk-accent);
	border: none;
	border-radius: 8px;
	padding: 7px 18px;
	font-size: 12.5px;
	color: #fff;
	font-family: 'DM Sans', sans-serif;
	cursor: pointer;
	font-weight: 500;
	transition: opacity 0.12s;
}

.vk-btn-primary:hover {
	opacity: 0.85;
}

.vk-btn-primary:disabled {
	opacity: 0.4;
	cursor: not-allowed;
}

.label-edit-button {
	inline-size: 1.45rem;
	block-size: 1.45rem;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	background: transparent;
	color: #ffffff;
	margin-inline-start: .25rem;
	padding: 0;
	border: 0;
	border-radius: 999px;
	cursor: pointer;

	.icon {
		inline-size: 1rem;
		block-size: .5rem;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background-color: rgba(0,0,0,0.2);
	}
}

.label-row {
	display: inline-flex;
	align-items: center;
	gap: 0;
	padding-right: .15rem;
}

.label-row-link {
	color: inherit;
	text-decoration: none;
	background: transparent;
	border: 0;
	padding: 0;
	font: inherit;
	cursor: default;
}

.label-row-link--editable {
	cursor: pointer;
}

@media (max-width: 640px) {
	.vk-overlay {
		padding: 14px;
	}

	.vk-modal {
		max-width: 100%;
	}

	.vk-modal-body,
	.vk-modal-foot,
	.vk-modal-head {
		padding-left: 16px;
		padding-right: 16px;
	}
}
</style>
