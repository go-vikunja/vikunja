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
						New project
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
							for="proj-title"
						>
							Title
						</label>
						<input
							id="proj-title"
							v-model="project.title"
							v-focus
							class="vk-input"
							type="text"
							name="projectTitle"
							placeholder="The project's title goes here..."
							@keydown.enter="createProject()"
							@keydown.esc="$router.back()"
						/>
						<p
							v-if="showError && project.title === ''"
							class="vk-field-error"
						>
							{{ $t('project.create.addTitleRequired') }}
						</p>
					</div>

					<div
						v-if="projectStore.hasProjects"
						class="vk-field"
					>
						<label
							class="vk-label"
							for="proj-parent"
						>
							Parent project
						</label>
						<div class="vk-input-icon-wrap">
							<svg
								class="vk-input-icon"
								width="14"
								height="14"
								viewBox="0 0 16 16"
								fill="none"
								stroke="currentColor"
								stroke-width="1.5"
								stroke-linecap="round"
							>
								<circle
									cx="7"
									cy="7"
									r="5"
								/>
								<path d="M11 11l3 3" />
							</svg>
							<input
								id="proj-parent"
								v-model="parentProject"
								class="vk-input vk-input--icon"
								type="text"
								placeholder="Search for a project..."
								@focus="showParentSuggestions = true"
								@blur="hideParentSuggestions"
							/>
							<div
								v-if="showParentSuggestions && filteredParentProjects.length > 0"
								class="vk-parent-suggestions"
							>
								<button
									v-for="parent in filteredParentProjects"
									:key="parent.id"
									type="button"
									class="vk-parent-suggestion"
									@mousedown.prevent="selectParentProject(parent.id, parent.title)"
								>
									{{ parent.title }}
								</button>
							</div>
						</div>
					</div>

					<div class="vk-field">
						<label class="vk-label">Color</label>
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
						Cancel
					</button>
					<button
						class="vk-btn-primary"
						:disabled="project.title.trim() === '' || isSubmitting"
						@click="createProject()"
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
						Create
					</button>
				</div>
			</div>
		</div>
	</Teleport>
</template>

<script setup lang="ts">
import {computed, ref, reactive, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import ProjectModel from '@/models/project'

import {success} from '@/message'
import {useTitle} from '@/composables/useTitle'
import {useProjectStore} from '@/stores/projects'

const props = defineProps<{
	parentProjectId?: number,
}>()

const {t} = useI18n({useScope: 'global'})

useTitle(() => t('project.create.header'))

const showError = ref(false)
const project = reactive(new ProjectModel())
const projectStore = useProjectStore()
const parentProject = ref('')
const selectedParentProjectId = ref<number | null>(null)
const showParentSuggestions = ref(false)
const isSubmitting = ref(false)
const selectedColor = ref('#6c63f5')

const colors = [
	'#6c63f5', '#10b981', '#f59e0b', '#ef4444',
	'#ec4899', '#06b6d4', '#8b5cf6', '#f97316',
]

const filteredParentProjects = computed(() => {
	const query = parentProject.value.trim()
	if (query === '') {
		return []
	}

	return projectStore.searchProject(query, false)
		.filter(p => p.id > 0)
		.slice(0, 8)
})

watch(
	() => props.parentProjectId,
	() => {
		const initialParent = props.parentProjectId ? projectStore.projects[props.parentProjectId] : null
		parentProject.value = initialParent?.title ?? ''
		selectedParentProjectId.value = initialParent?.id ?? null
	},
	{immediate: true},
)

watch(parentProject, value => {
	if (value.trim() === '') {
		selectedParentProjectId.value = null
		return
	}

	const selectedTitle = selectedParentProjectId.value
		? projectStore.projects[selectedParentProjectId.value]?.title
		: null

	if (selectedTitle && value.trim().toLowerCase() !== selectedTitle.toLowerCase()) {
		selectedParentProjectId.value = null
	}
})

function selectParentProject(id: number, title: string) {
	selectedParentProjectId.value = id
	parentProject.value = title
	showParentSuggestions.value = false
}

function hideParentSuggestions() {
	setTimeout(() => {
		showParentSuggestions.value = false
	}, 100)
}

async function createProject() {
	if (project.title === '') {
		showError.value = true
		return
	}
	showError.value = false

	if (isSubmitting.value) {
		return
	}

	isSubmitting.value = true

	project.hexColor = selectedColor.value

	const parentProjectTitle = parentProject.value.trim()
	if (selectedParentProjectId.value !== null) {
		project.parentProjectId = selectedParentProjectId.value
	} else if (parentProjectTitle !== '') {
		const foundParentProject = projectStore.findProjectByExactname(parentProjectTitle)
		if (foundParentProject) {
			project.parentProjectId = foundParentProject.id
		}
	}

	try {
		await projectStore.createProject(project)
		success({message: t('project.create.createdSuccess')})
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
	background: #13141a;
	border: 0.5px solid #2a2b35;
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
	border-bottom: 0.5px solid #1e1f28;
}

.vk-modal-title {
	font-family: 'Playfair Display', serif;
	font-size: 16px;
	font-weight: 400;
	color: #f0ede8;
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
	color: #4a4b57;
	transition: all 0.12s;
}

.vk-modal-close:hover {
	background: #1e1f28;
	color: #a0a0b0;
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
	border-top: 0.5px solid #1e1f28;
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
	color: #6a6a7a;
	letter-spacing: 0.04em;
	text-transform: uppercase;
	margin-bottom: 7px;
}

.vk-input {
	width: 100%;
	background: #0e0f13;
	border: 0.5px solid #2a2b35;
	border-radius: 8px;
	padding: 10px 13px;
	font-size: 13px;
	color: #c0bdb8;
	font-family: 'DM Sans', sans-serif;
	outline: none;
	transition: border-color 0.15s;
	box-shadow: none;
	height: auto;
}

.vk-input::placeholder {
	color: #3a3b48;
}

.vk-input:focus {
	border-color: #6c63f5;
	box-shadow: none;
}

.vk-input-icon-wrap {
	position: relative;
}

.vk-parent-suggestions {
	position: absolute;
	left: 0;
	right: 0;
	top: calc(100% + 6px);
	background: #13141a;
	border: 0.5px solid #2a2b35;
	border-radius: 8px;
	overflow: hidden;
	box-shadow: 0 12px 28px rgba(0, 0, 0, 0.42);
	z-index: 2;
}

.vk-parent-suggestion {
	display: block;
	width: 100%;
	background: transparent;
	border: 0;
	padding: 9px 12px;
	text-align: left;
	color: #c0bdb8;
	font-family: 'DM Sans', sans-serif;
	font-size: 13px;
	cursor: pointer;
	transition: background 0.12s;
}

.vk-parent-suggestion:hover {
	background: #1e1f28;
}

.vk-input-icon {
	position: absolute;
	left: 11px;
	top: 50%;
	transform: translateY(-50%);
	color: #3a3b48;
	pointer-events: none;
}

.vk-input--icon {
	padding-left: 34px;
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
	border: 0.5px solid #2a2b35;
	border-radius: 8px;
	padding: 7px 16px;
	font-size: 12.5px;
	color: #6a6a7a;
	font-family: 'DM Sans', sans-serif;
	cursor: pointer;
	transition: all 0.12s;
}

.vk-btn-cancel:hover {
	background: #1e1f28;
	color: #a0a0b0;
}

.vk-btn-primary {
	display: inline-flex;
	align-items: center;
	gap: 6px;
	background: #6c63f5;
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
