<template>
	<Teleport to="body">
		<div
			class="vk-overlay"
			@click.self="$router.back()"
		>
			<div
				class="vk-modal"
				role="dialog"
				aria-modal="true"
				aria-labelledby="modal-title"
			>
				<div class="vk-modal-head">
					<h2
						id="modal-title"
						class="vk-modal-title"
					>
						{{ $t('label.create.title') }}
					</h2>
					<button
						class="vk-modal-close"
						aria-label="Close"
						@click="$router.back()"
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

				<div class="vk-modal-body">
					<div class="vk-field">
						<label
							class="vk-label"
							for="label-title"
						>
							{{ $t('label.attributes.title') }}
						</label>
						<input
							id="label-title"
							v-model="label.title"
							v-focus
							class="vk-input"
							type="text"
							:placeholder="$t('label.attributes.titlePlaceholder')"
							@keydown.enter="newLabel()"
							@keydown.esc="$router.back()"
						/>
						<p
							v-if="showError && label.title === ''"
							class="vk-field-error"
						>
							{{ $t('label.create.titleRequired') }}
						</p>
					</div>

					<div class="vk-field">
						<label class="vk-label">{{ $t('label.attributes.color') }}</label>
						<div class="vk-swatches">
							<button
								v-for="color in colors"
								:key="color"
								type="button"
								class="vk-swatch"
								:class="{'vk-swatch--active': selectedColor === color}"
								:style="{background: color}"
								:aria-label="`Select color ${color}`"
								@click="selectedColor = color"
							/>
						</div>
					</div>
				</div>

				<div class="vk-modal-foot">
					<button
						class="vk-btn-cancel"
						@click="$router.back()"
					>
						{{ $t('misc.cancel') }}
					</button>
					<button
						class="vk-btn-primary"
						:disabled="label.title.trim() === '' || loadingModel"
						@click="newLabel()"
					>
						<svg
							width="12"
							height="12"
							viewBox="0 0 16 16"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
						>
							<path d="M8 2v12M2 8h12" />
						</svg>
						{{ $t('misc.create') }}
					</button>
				</div>
			</div>
		</div>
	</Teleport>
</template>

<script setup lang="ts">
import {computed, onBeforeMount, ref} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import LabelModel from '@/models/label'
import {useLabelStore} from '@/stores/labels'
import {useTitle} from '@/composables/useTitle'
import {success} from '@/message'
import {getRandomColorHex} from '@/helpers/color/randomColor'

const router = useRouter()

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('label.create.title'))

const labelStore = useLabelStore()
const label = ref(new LabelModel())
const selectedColor = ref('')

const colors = [
	'#6c63f5', '#10b981', '#f59e0b', '#ef4444',
	'#ec4899', '#06b6d4', '#8b5cf6', '#f97316',
]

onBeforeMount(() => {
	label.value.hexColor = getRandomColorHex()
	selectedColor.value = label.value.hexColor
})

const showError = ref(false)
const loading = computed(() => labelStore.isLoading)
const isSubmitting = ref(false)

const loadingModel = computed({
	get: () => isSubmitting.value || loading.value,
	set(value: boolean) {
		isSubmitting.value = value
	},
})

async function newLabel() {
	if (label.value.title === '') {
		showError.value = true
		return
	}
	showError.value = false

	if (isSubmitting.value) {
		return
	}

	isSubmitting.value = true
	label.value.hexColor = selectedColor.value

	try {
		const newLabel = await labelStore.createLabel(label.value)
		router.push({
			name: 'labels.index',
			params: {id: newLabel.id},
		})
		success({message: t('label.create.success')})
	} finally {
		isSubmitting.value = false
	}
}
</script>

<style scoped>
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
	max-width: 460px;
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

.vk-field-error {
	margin-top: 6px;
	font-size: 12px;
	color: #ef4444;
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
