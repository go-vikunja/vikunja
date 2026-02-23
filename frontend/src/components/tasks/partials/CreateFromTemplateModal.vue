<template>
	<Modal
		:enabled="enabled"
		@close="$emit('close')"
	>
		<div class="template-modal-content">
			<div class="modal-header">
				{{ $t('task.template.createFromTemplate') }}
			</div>
			<div class="content">
				<!-- Step 1: Select a template -->
				<div
					v-if="!selectedTemplate"
					class="template-list-section"
				>
					<div class="field">
						<div class="control">
							<input
								v-model="searchQuery"
								v-focus
								class="input"
								:placeholder="$t('task.template.searchPlaceholder')"
								type="text"
							>
						</div>
					</div>

					<div
						v-if="loading"
						class="has-text-centered p-4"
					>
						<span class="loader is-loading" />
					</div>

					<div
						v-else-if="filteredTemplates.length === 0"
						class="has-text-centered has-text-grey p-4"
					>
						{{ searchQuery ? $t('task.template.noResults') : $t('task.template.noTemplates') }}
					</div>

					<div
						v-else
						class="template-list"
					>
						<BaseButton
							v-for="tmpl in filteredTemplates"
							:key="tmpl.id"
							class="template-item"
							@click="selectTemplate(tmpl)"
						>
							<div class="template-item-header">
								<span class="template-title">{{ tmpl.title }}</span>
								<span
									v-if="tmpl.priority > 0"
									class="tag is-small"
									:class="priorityClass(tmpl.priority)"
								>
									P{{ tmpl.priority }}
								</span>
							</div>
							<p
								v-if="tmpl.description"
								class="template-description"
							>
								{{ truncate(stripHtml(tmpl.description), 80) }}
							</p>
						</BaseButton>
					</div>
				</div>

				<!-- Step 2: Configure and confirm -->
				<div v-else>
					<BaseButton
						class="back-link"
						@click="selectedTemplate = null"
					>
						<Icon icon="arrow-left" />
						{{ $t('task.template.backToList') }}
					</BaseButton>

					<div class="selected-template-info">
						<h4>{{ selectedTemplate.title }}</h4>
						<p
							v-if="selectedTemplate.description"
							class="has-text-grey"
						>
							{{ truncate(stripHtml(selectedTemplate.description), 120) }}
						</p>
					</div>

					<div class="field">
						<label class="label">{{ $t('task.template.taskTitle') }}</label>
						<div class="control">
							<input
								v-model="taskTitle"
								v-focus
								class="input"
								:placeholder="selectedTemplate.title"
								type="text"
								@keyup.enter="createFromTemplate"
							>
						</div>
						<p class="help">
							{{ $t('task.template.titleHint') }}
						</p>
					</div>

					<div class="field">
						<label class="label">{{ $t('task.template.targetProject') }}</label>
						<div class="control">
							<ProjectSearch
								v-model="selectedProject"
								@update:modelValue="handleProjectSelect"
							/>
						</div>
					</div>

					<div
						v-if="errMessage"
						class="notification is-danger is-light mt-2"
					>
						{{ errMessage }}
					</div>
				</div>
			</div>
			<div
				v-if="selectedTemplate"
				class="actions"
			>
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
					:disabled="!selectedProject || selectedProject.id === 0"
					@click="createFromTemplate"
				>
					{{ $t('task.template.create') }}
				</XButton>
			</div>
		</div>
	</Modal>
</template>

<script lang="ts" setup>
import {ref, computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import Modal from '@/components/misc/Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'

import TaskTemplateService from '@/services/taskTemplateService'
import TaskFromTemplateService from '@/services/taskFromTemplateService'
import TaskFromTemplateModel from '@/models/taskFromTemplate'

import type {ITaskTemplate} from '@/modelTypes/ITaskTemplate'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'

import {success} from '@/message'
import {useProjectStore} from '@/stores/projects'

const props = defineProps<{
	enabled: boolean,
	defaultProjectId?: number,
}>()

const emit = defineEmits<{
	'close': [],
	'created': [task: ITask],
}>()

const {t} = useI18n({useScope: 'global'})
const projectStore = useProjectStore()

const templates = ref<ITaskTemplate[]>([])
const selectedTemplate = ref<ITaskTemplate | null>(null)
const selectedProject = ref<IProject | null>(null)
const taskTitle = ref('')
const searchQuery = ref('')
const loading = ref(false)
const creating = ref(false)
const errMessage = ref('')

const filteredTemplates = computed(() => {
	if (!searchQuery.value) return templates.value
	const q = searchQuery.value.toLowerCase()
	return templates.value.filter(t =>
		t.title.toLowerCase().includes(q) ||
		(t.description && t.description.toLowerCase().includes(q)),
	)
})

// Load templates when modal opens
watch(() => props.enabled, async (newVal) => {
	if (newVal) {
		selectedTemplate.value = null
		taskTitle.value = ''
		searchQuery.value = ''
		errMessage.value = ''

		// Pre-select the current project if available
		if (props.defaultProjectId && props.defaultProjectId > 0) {
			const proj = projectStore.projects[props.defaultProjectId]
			selectedProject.value = proj || null
		} else {
			selectedProject.value = null
		}

		await loadTemplates()
	}
})

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

function selectTemplate(tmpl: ITaskTemplate) {
	selectedTemplate.value = tmpl
	taskTitle.value = ''
	errMessage.value = ''
}

function handleProjectSelect(project: IProject | null) {
	selectedProject.value = project
	errMessage.value = ''
}

async function createFromTemplate() {
	if (!selectedTemplate.value || !selectedProject.value || selectedProject.value.id === 0) {
		errMessage.value = t('task.template.selectProjectRequired')
		return
	}

	creating.value = true
	errMessage.value = ''

	try {
		const service = new TaskFromTemplateService()
		const result = await service.create(
			new TaskFromTemplateModel({
				templateId: selectedTemplate.value.id,
				targetProjectId: selectedProject.value.id,
				title: taskTitle.value || '',
			}),
		)

		success({message: t('task.template.createSuccess')})
		emit('created', result.createdTask!)
		emit('close')
	} catch (e: any) {
		errMessage.value = e?.message || t('task.template.createError')
	} finally {
		creating.value = false
	}
}

function stripHtml(html: string): string {
	const doc = new DOMParser().parseFromString(html, "text/html")
	return doc.body.textContent || ""
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
.template-modal-content {
	text-align: start;
	padding: 0 1rem;

	.modal-header {
		font-size: 2rem;
		font-weight: 700;
		text-align: center;
		margin-block-end: 1rem;
	}

	.actions {
		margin-block-start: 1.5rem;
		text-align: center;

		.button {
			margin: 0 0.5rem;
		}
	}
}

.template-list-section {
	min-block-size: 200px;
}

.template-list {
	max-block-size: 300px;
	overflow-y: auto;
}

.template-item {
	display: block;
	inline-size: 100%;
	padding: .75rem;
	text-align: start;
	border-radius: $radius;
	border: 1px solid var(--grey-200);
	margin-block-end: .5rem;
	transition: background-color $transition-duration;

	&:hover {
		background-color: var(--grey-100);
	}
}

.template-item-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
}

.template-title {
	font-weight: 600;
	color: var(--text);
}

.template-description {
	font-size: .85rem;
	color: var(--grey-500);
	margin-block-start: .25rem;
	margin-block-end: 0;
}

.back-link {
	color: var(--primary);
	font-size: .9rem;
	margin-block-end: 1rem;
	display: inline-flex;
	align-items: center;
	gap: .25rem;

	&:hover {
		text-decoration: underline;
	}
}

.selected-template-info {
	background: var(--grey-100);
	padding: .75rem;
	border-radius: $radius;
	margin-block-end: 1rem;

	h4 {
		margin: 0 0 .25rem;
		font-weight: 600;
	}

	p {
		margin: 0;
		font-size: .85rem;
	}
}

.field .label {
	color: var(--text);
	font-weight: 600;
}

.help {
	color: var(--grey-500);
	font-size: .85rem;
}
</style>
