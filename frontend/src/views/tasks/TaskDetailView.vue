<template>
	<div
		class="loader-container task-view-container"
		:class="{
			'is-loading': taskService.loading || !visible,
			'is-modal': isModal,
		}"
	>
		<!-- Removing everything until the task is loaded to prevent empty initialization of other components -->
		<div
			v-if="visible"
			class="task-view jira-layout-container"
		>
			<!-- Jira-like two-column layout -->
			<div class="jira-main-content">
				<Heading
					ref="heading"
					:task="task"
					:can-write="canWrite"
					:has-close="isModal"
					@update:task="Object.assign(task, $event)"
					@close="$emit('close')"
				/>
				<div
					v-if="project?.id"
					class="subtitle-container"
				>
					<h6 class="subtitle">
						<template
							v-for="p in projectStore.getAncestors(project)"
							:key="p.id"
						>
							<a
								v-if="router.options.history.state.back?.includes('/projects/'+p.id+'/') || false"
								@click="router.back()"
							>
								{{ getProjectTitle(p) }}
							</a>
							<RouterLink
								v-else
								:to="{ name: 'project.index', params: { projectId: p.id } }"
							>
								{{ getProjectTitle(p) }}
							</RouterLink>
							<span
								v-if="p.id !== project?.id"
								class="has-text-grey-light"
							> &gt; </span>
						</template>
					</h6>
					<BaseButton 
						v-if="canWrite"
						class="move-project-button is-small"
						@click="isMovingProject = !isMovingProject"
					>
						<Icon icon="exchange-alt" class="mr-1" /> {{ $t('task.detail.move') }}
					</BaseButton>
				</div>
				
				<div v-if="isMovingProject && canWrite" class="project-search-container">
					<ProjectSearch
						:ref="e => setFieldRef('moveProject', e)"
						:filter="project => project.id !== task.projectId"
						@update:modelValue="(p) => { changeProject(p); isMovingProject = false; }"
					/>
					<BaseButton 
						class="cancel-move-button is-small"
						@click="isMovingProject = false"
					>
						{{ $t('misc.cancel') }}
					</BaseButton>
				</div>

				<ChecklistSummary :task="task" />

				<!-- Description -->
				<div class="details content description">
					<Description
						:model-value="task"
						:can-write="canWrite"
						:attachment-upload="attachmentUpload"
						@update:modelValue="Object.assign(task, $event)"
					/>
					<Reactions
						v-if="task.id"
						v-model="task.reactions"
						entity-kind="tasks"
						:entity-id="task.id"
						class="reactions-in-description"
						:disabled="!canWrite"
					/>
				</div>

				<!-- Reactions -->
				<!-- Reactions component is now intended to be integrated within Description.vue -->
				<!-- <Reactions
					v-model="task.reactions"
					entity-kind="tasks"
					:entity-id="task.id"
					class="details reactions-container"
					:disabled="!canWrite"
				/> -->

				<!-- Tabs for Attachments, Related Tasks, Comments -->
				<div class="main-content-tabs-container">
					<div class="tabs is-boxed">
						<ul>
							<li :class="{'is-active': activeMainContentTab === 'comments'}">
								<a @click="setActiveMainContentTab('comments')">
									<span class="icon is-small"><Icon :icon="['far', 'comments']" /></span>
									<span>{{ $t('task.comment.title') }}</span>
								</a>
							</li>
							<li :class="{'is-active': activeMainContentTab === 'attachments'}">
								<a @click="setActiveMainContentTab('attachments')">
									<span class="icon is-small"><Icon icon="paperclip" /></span>
									<span>{{ $t('task.attachment.title') }}</span>
								</a>
							</li>
							<li :class="{'is-active': activeMainContentTab === 'relatedTasks'}">
								<a @click="setActiveMainContentTab('relatedTasks')">
									<span class="icon is-small"><Icon icon="sitemap" /></span>
									<span>{{ $t('task.attributes.relatedTasks') }}</span>
								</a>
							</li>
						</ul>
					</div>

					<div class="tab-content">
						<!-- Comments -->
						<div v-if="activeMainContentTab === 'comments'" class="comments-section">
							<Comments
								:can-write="canWrite"
								:task-id="taskId"
								:initial-comments="task.comments"
							/>
						</div>

						<!-- Attachments -->
						<div v-if="activeMainContentTab === 'attachments'" class="content attachments">
							<Attachments
								:ref="e => setFieldRef('attachments', e)"
								:edit-enabled="canWrite"
								:task="task"
								@taskChanged="({coverImageAttachmentId}) => task.coverImageAttachmentId = coverImageAttachmentId"
							/>
						</div>

						<!-- Related Tasks -->
						<div v-if="activeMainContentTab === 'relatedTasks'" class="content details related-tasks mb-0">
							<RelatedTasks
								:ref="e => setFieldRef('relatedTasks', e)"
								:edit-enabled="canWrite"
								:initial-related-tasks="task.relatedTasks"
								:project-id="task.projectId"
								:show-no-relations-notice="true"
								:task-id="taskId"
							/>
						</div>
					</div>
				</div>
			</div>

			<div class="jira-sidebar">
				<div class="sidebar-top-actions">
					<x-button
						v-if="canWrite"
						v-shortcut="'t'"
						:class="{'is-success': !task.done}"
						:shadow="task.done"
						class="is-outlined has-no-border task-status-button"
						icon="check-double"
						variant="secondary"
						@click="toggleTaskDone()"
					>
						{{ task.done ? $t('task.detail.undone') : $t('task.detail.done') }}
					</x-button>
					<!-- More top actions like share, watch etc. can be added here -->
				</div>

				<div class="sidebar-details">
					<h4 class="sidebar-section-title">{{ $t('task.detail.details') }}</h4>

					<!-- Assignees -->
					<div class="sidebar-attribute-item">
						<div class="detail-title is-flex is-justify-content-space-between">
							<span>
								<Icon icon="users" />
								{{ $t('task.attributes.assignees') }}
							</span>
							<BaseButton 
								v-if="canWrite && !isEditingAssignees"
								class="is-small is-text action-edit-button"
								@click="isEditingAssignees = true"
							>
								<Icon icon="pen" class="mr-1" /> {{ $t('input.editor.edit') }}
							</BaseButton>
							<BaseButton 
								v-if="canWrite && isEditingAssignees"
								class="is-small is-text action-edit-button"
								@click="() => { isEditingAssignees = false; /* saveTask will be triggered by EditAssignees component changes */ }"
							>
								<Icon icon="check" class="mr-1" /> {{ $t('input.editor.done') }}
							</BaseButton>
						</div>
						<EditAssignees
							v-if="canWrite && isEditingAssignees"
							:ref="e => setFieldRef('assignees', e)"
							v-model="task.assignees"
							:project-id="task.projectId"
							:task-id="task.id"
						/>
						<AssigneeList
							v-else
							:assignees="task.assignees"
							:can-write="canWrite" 
							:task-id="task.id" 
							class="mt-1"
							@update:assignees="(newAssignees) => { task.assignees = newAssignees; saveTask(); }"
						/>
					</div>

					<!-- Labels -->
					<div class="sidebar-attribute-item">
						<div class="detail-title is-flex is-justify-content-space-between">
							<span>
								<span class="icon is-grey">
									<Icon icon="tags" />
								</span>
								{{ $t('task.attributes.labels') }}
							</span>
							<BaseButton 
								v-if="canWrite && !isEditingLabels"
								class="is-small is-text action-edit-button"
								@click="isEditingLabels = true"
							>
								<Icon icon="pen" class="mr-1" /> {{ $t('input.editor.edit') }}
							</BaseButton>
							<BaseButton 
								v-if="canWrite && isEditingLabels"
								class="is-small is-text action-edit-button"
								@click="() => { isEditingLabels = false; /* saveTask will be triggered by EditLabels component changes */ }"
							>
								<Icon icon="check" class="mr-1" /> {{ $t('input.editor.done') }}
							</BaseButton>
						</div>
						<EditLabels
							v-if="canWrite && isEditingLabels"
							:ref="e => setFieldRef('labels', e)"
							v-model="task.labels"
							:disabled="!canWrite" 
							:task-id="taskId"
							:creatable="!authStore.isLinkShareAuth"
						/>
						<div v-else class="labels-display-area">
							<span v-if="!task.labels || task.labels.length === 0" class="has-text-grey-light is-italic">{{ $t('misc.none') }}</span>
							<span 
								v-for="label in task.labels" 
								:key="label.id" 
								class="tag is-rounded mr-1 mb-1"
								:style="{ backgroundColor: label.hexColor, color: label.fontColor || '#fff' }"
							>
								{{ label.title }}
								<button 
									v-if="canWrite"
									class="delete is-small"
									@click="removeLabel(label)"
								></button>
							</span>
						</div>
					</div>
					
					<!-- Reporter (Mimicking Jira's Reporter) -->
					<div v-if="task.createdBy" class="sidebar-attribute-item">
						<div class="detail-title">
							<Icon icon="user-check" />
							{{ $t('task.attributes.reporter') }}
						</div>
						<div class="reporter-info">
							<User
								:user="task.createdBy"
								:avatar-size="30"
								:show-username="true"
							/>
						</div>
					</div>


					<!-- Priority -->
					<div class="sidebar-attribute-item">
						<div class="detail-title">
							<Icon icon="exclamation-circle" />
							{{ $t('task.attributes.priority') }}
						</div>
						<PrioritySelect
							:ref="e => setFieldRef('priority', e)"
							v-model="task.priority"
							:disabled="!canWrite"
							@update:modelValue="setPriority"
						/>
					</div>

					<!-- PercentDone -->
					<div class="sidebar-attribute-item">
						<div class="detail-title">
							<Icon icon="percent" />
							{{ $t('task.attributes.percentDone') }}
						</div>
						<PercentDoneSelect
							:ref="e => setFieldRef('percentDone', e)"
							v-model="task.percentDone"
							:disabled="!canWrite"
							@update:modelValue="setPercentDone"
						/>
					</div>
					
					<!-- Due Date -->
					<div class="sidebar-attribute-item">
						<div class="detail-title">
							<Icon icon="calendar" />
							{{ $t('task.attributes.dueDate') }}
						</div>
						<div class="date-input">
							<Datepicker
								:ref="e => setFieldRef('dueDate', e)"
								v-model="task.dueDate"
								:choose-date-label="$t('task.detail.chooseDueDate')"
								:disabled="taskService.loading || !canWrite"
								@closeOnChange="saveTask()"
							/>
							<BaseButton
								v-if="task.dueDate && canWrite"
								class="remove"
								@click="() => {task.dueDate = null;saveTask()}"
							>
								<span class="icon is-small">
									<Icon icon="times" />
								</span>
							</BaseButton>
						</div>
					</div>

					<!-- Start Date -->
					<div class="sidebar-attribute-item">
						<div class="detail-title">
							<Icon icon="play" />
							{{ $t('task.attributes.startDate') }}
						</div>
						<div class="date-input">
							<Datepicker
								:ref="e => setFieldRef('startDate', e)"
								v-model="task.startDate"
								:choose-date-label="$t('task.detail.chooseStartDate')"
								:disabled="taskService.loading || !canWrite"
								@closeOnChange="saveTask()"
							/>
							<BaseButton
								v-if="task.startDate && canWrite"
								class="remove"
								@click="() => {task.startDate = null;saveTask()}"
							>
								<span class="icon is-small">
									<Icon icon="times" />
								</span>
							</BaseButton>
						</div>
					</div>
					
					<!-- End Date -->
					<div class="sidebar-attribute-item">
						<div class="detail-title">
							<Icon icon="stop" />
							{{ $t('task.attributes.endDate') }}
						</div>
						<div class="date-input">
							<Datepicker
								:ref="e => setFieldRef('endDate', e)"
								v-model="task.endDate"
								:choose-date-label="$t('task.detail.chooseEndDate')"
								:disabled="taskService.loading || !canWrite"
								@closeOnChange="saveTask()"
							/>
							<BaseButton
								v-if="task.endDate && canWrite"
								class="remove"
								@click="() => {task.endDate = null;saveTask()}"
							>
								<span class="icon is-small">
									<Icon icon="times" />
								</span>
							</BaseButton>
						</div>
					</div>

					<!-- Reminders -->
					<div v-if="activeFields.reminders || canWrite" class="sidebar-attribute-item">
						<div class="detail-title">
							<Icon :icon="['far', 'clock']" />
							{{ $t('task.attributes.reminders') }}
						</div>
						<Reminders
							:ref="e => setFieldRef('reminders', e)"
							v-model="task"
							:disabled="!canWrite"
							@update:modelValue="saveTask()"
						/>
						<!-- <x-button v-if="!activeFields.reminders && canWrite" @click="setFieldActive('reminders')"> {{ $t('task.detail.actions.addReminders') }} </x-button> -->
					</div>

					<!-- Repeat after -->
					<div v-if="canWrite || activeFields.repeatAfter" class="sidebar-attribute-item">
						<div class="is-flex is-justify-content-space-between">
							<div class="detail-title">
								<Icon icon="history" />
								{{ $t('task.attributes.repeat') }}
							</div>
							<BaseButton
								v-if="canWrite && (task.repeatAfter?.amount > 0 || task.repeatMode !== TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT)"
								class="remove"
								@click="removeRepeatAfter"
							>
								<span class="icon is-small">
									<Icon icon="times" />
								</span>
							</BaseButton>
						</div>
						<RepeatAfter
							:ref="e => setFieldRef('repeatAfter', e)"
							v-model="task"
							:disabled="!canWrite"
							@update:modelValue="saveTask()"
						/>
                         <!-- <x-button v-if="!activeFields.repeatAfter && canWrite" @click="setFieldActive('repeatAfter')"> {{ $t('task.detail.actions.setupRepeat') }} </x-button> -->
					</div>
				</div>
				
				<div class="sidebar-actions">
					<h4 class="sidebar-section-title">{{ $t('task.detail.actions.title') }}</h4>  <!-- Add to translations -->
					<TaskSubscription
						entity="task"
						:entity-id="task.id"
						:model-value="task.subscription"
						class="sidebar-action-button"
						@update:modelValue="sub => task.subscription = sub"
					/>
					<x-button
						v-if="canWrite"
						v-shortcut="'s'"
						variant="secondary"
						:icon="task.isFavorite ? 'star' : ['far', 'star']"
						class="sidebar-action-button"
						@click="toggleFavorite"
					>
						{{
							task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')
						}}
					</x-button>
					
					<!-- Move Task - 已移动到标题下方 -->

					<x-button
						v-if="canWrite"
						v-shortcut="'Shift+Delete'"
						icon="trash-alt"
						:shadow="false"
						class="is-danger is-outlined has-no-border sidebar-action-button"
						@click="showDeleteModal = true"
					>
						{{ $t('task.detail.actions.delete') }}
					</x-button>
				</div>
				
				<!-- Created / Updated -->
				<CreatedUpdated :task="task" class="sidebar-created-updated" />
			</div>
			<!-- The old layout for permanent attributes, content, and action buttons is removed -->
			<!-- Columns structure is removed -->
		</div>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteTask()"
		>
			<template #header>
				<span>{{ $t('task.detail.delete.header') }}</span>
			</template>

			<template #text>
				<p class="tw-text-balance !tw-mb-0">
					{{ $t('task.detail.delete.text1') }}
				</p>
				<p class="tw-text-balance">
					{{ $t('task.detail.delete.text2') }}
				</p>
			</template>
		</Modal>
	</div>
</template>

<script lang="ts" setup>
import {ref, reactive, shallowReactive, computed, watch, nextTick, onMounted, onBeforeUnmount} from 'vue'
import {useRouter, type RouteLocation} from 'vue-router'
import {storeToRefs} from 'pinia'
import {useI18n} from 'vue-i18n'
import {unrefElement} from '@vueuse/core'
import {klona} from 'klona/lite'
import {eventToHotkeyString} from '@github/hotkey'

import TaskService from '@/services/task'
import TaskModel from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'

import {PRIORITIES, type Priority} from '@/constants/priorities'
import {RIGHTS} from '@/constants/rights'

import BaseButton from '@/components/base/BaseButton.vue'

// partials
import Attachments from '@/components/tasks/partials/Attachments.vue'
import ChecklistSummary from '@/components/tasks/partials/ChecklistSummary.vue'
import Comments from '@/components/tasks/partials/Comments.vue'
import CreatedUpdated from '@/components/tasks/partials/CreatedUpdated.vue'
import Datepicker from '@/components/input/Datepicker.vue'
import Description from '@/components/tasks/partials/Description.vue'
import EditAssignees from '@/components/tasks/partials/EditAssignees.vue'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import Heading from '@/components/tasks/partials/Heading.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import PercentDoneSelect from '@/components/tasks/partials/PercentDoneSelect.vue'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import RelatedTasks from '@/components/tasks/partials/RelatedTasks.vue'
import Reminders from '@/components/tasks/partials/Reminders.vue'
import RepeatAfter from '@/components/tasks/partials/RepeatAfter.vue'
import TaskSubscription from '@/components/misc/Subscription.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import Reactions from '@/components/input/Reactions.vue'
import User from '@/components/misc/User.vue'

import {uploadFile} from '@/helpers/attachments'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {scrollIntoView} from '@/helpers/scrollIntoView'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import {playPopSound} from '@/helpers/playPop'

import {useAttachmentStore} from '@/stores/attachments'
import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'

import {useTitle} from '@/composables/useTitle'

import {success} from '@/message'
import type {Action as MessageAction} from '@/message'

const props = defineProps<{
	taskId: ITask['id'],
	backdropView?: RouteLocation['fullPath'],
}>()

defineEmits<{
	'close': [],
}>()

const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const projectStore = useProjectStore()
const attachmentStore = useAttachmentStore()
const {hasAttachments} = storeToRefs(attachmentStore)
const taskStore = useTaskStore()
const kanbanStore = useKanbanStore()
const authStore = useAuthStore()
const baseStore = useBaseStore()

const task = ref<ITask>(new TaskModel())
const taskTitle = computed(() => task.value.title)
useTitle(taskTitle)

const activeMainContentTab = ref('comments') // Default to 'comments'

const isEditingAssignees = ref(false)
const isEditingLabels = ref(false)
const isMovingProject = ref(false)

function setActiveMainContentTab(tabName: string) {
	activeMainContentTab.value = tabName
}

// Used to avoid flashing of empty elements if the task content is not yet loaded.
const visible = ref(false)

const project = computed(() => projectStore.projects[task.value.projectId])

const canWrite = computed(() => (
	task.value.maxRight !== null &&
	task.value.maxRight > RIGHTS.READ
))

const color = computed(() => {
	const color = task.value.getHexColor
		? task.value.getHexColor()
		: undefined

	return color
})

const isModal = computed(() => Boolean(props.backdropView))

function attachmentUpload(file: File, onSuccess?: (url: string) => void) {
	return uploadFile(props.taskId, file, onSuccess)
}

const heading = ref<HTMLElement | null>(null)

async function scrollToHeading() {
	scrollIntoView(unrefElement(heading))
}

const taskService = shallowReactive(new TaskService())

// load task
watch(
	() => props.taskId,
	async (id) => {
		if (id === undefined) {
			return
		}

		try {
			const loaded = await taskService.get({id}, {expand: ['reactions', 'comments']})
			Object.assign(task.value, loaded)
			attachmentStore.set(task.value.attachments)
			setActiveFields()
		} finally {
			await nextTick()
			scrollToHeading()
			visible.value = true
		}
	}, {immediate: true})

type FieldType =
	| 'assignees'
	| 'attachments'
	| 'dueDate'
	| 'endDate'
	| 'labels'
	| 'moveProject'
	| 'percentDone'
	| 'priority'
	| 'relatedTasks'
	| 'reminders'
	| 'repeatAfter'
	| 'startDate'

const activeFields: { [type in FieldType]: boolean } = reactive({
	assignees: false,
	attachments: false,
	dueDate: false,
	endDate: false,
	labels: false,
	moveProject: false,
	percentDone: false,
	priority: false,
	relatedTasks: false,
	reminders: false,
	repeatAfter: false,
	startDate: false,
})

function setActiveFields() {
	// FIXME: are these lines necessary?
	// task.startDate = task.startDate || null
	// task.endDate = task.endDate || null

	// Set all active fields based on values in the model
	activeFields.assignees = task.value.assignees.length > 0
	activeFields.attachments = task.value.attachments.length > 0
	activeFields.dueDate = task.value.dueDate !== null
	activeFields.endDate = task.value.endDate !== null
	activeFields.labels = task.value.labels.length > 0
	activeFields.moveProject = false // moveProject is an action, not a field with persistent state reflecting in activeFields
	activeFields.percentDone = task.value.percentDone > 0
	activeFields.priority = task.value.priority !== PRIORITIES.UNSET
	activeFields.relatedTasks = Object.keys(task.value.relatedTasks).length > 0
	activeFields.reminders = task.value.reminders.length > 0
	activeFields.repeatAfter = task.value.repeatAfter?.amount > 0 || task.value.repeatMode !== TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT
	activeFields.startDate = task.value.startDate !== null
}

const activeFieldElements: { [id in FieldType]: HTMLElement | null } = reactive({
	assignees: null,
	attachments: null,
	dueDate: null,
	endDate: null,
	labels: null,
	moveProject: null,
	percentDone: null,
	priority: null,
	relatedTasks: null,
	reminders: null,
	repeatAfter: null,
	startDate: null,
})

function setFieldRef(name, e) {
	activeFieldElements[name] = unrefElement(e)
}

function setFieldActive(fieldName: keyof typeof activeFields) {
	activeFields[fieldName] = true
	nextTick(() => {
		const el = activeFieldElements[fieldName]

		if (!el) {
			return
		}

		el.focus()

		// scroll the field to the center of the screen if not in viewport already
		scrollIntoView(el)
	})
}

async function saveTask(
	currentTask: ITask | null = null,
	undoCallback?: () => void,
) {
	if (currentTask === null) {
		currentTask = klona(task.value)
	}

	if (!canWrite.value) {
		return
	}

	const updatedTask = await taskStore.update(currentTask) // TODO: markraw ?
	Object.assign(task.value, updatedTask)
	setActiveFields()

	let actions: MessageAction[] = []
	if (undoCallback) {
		actions = [{
			title: t('task.undo'),
			callback: undoCallback,
		}]
	}
	success({message: t('task.detail.updateSuccess')}, actions)
}

const showDeleteModal = ref(false)

async function deleteTask() {
	await taskStore.delete(task.value)
	success({message: t('task.detail.deleteSuccess')})
	router.push({name: 'project.index', params: {projectId: task.value.projectId}})
}

async function toggleTaskDone() {
	const newTask = {
		...task.value,
		done: !task.value.done,
	}

	if (newTask.done) {
		playPopSound()
	}

	await saveTask(
		newTask,
		toggleTaskDone,
	)
}

async function changeProject(project: IProject) {
	kanbanStore.removeTaskInBucket(task.value)
	await saveTask({
		...task.value,
		projectId: project.id,
	})
	baseStore.setCurrentProject(project)
}

async function toggleFavorite() {
	const newTask = await taskStore.toggleFavorite(task.value)
	Object.assign(task.value, newTask)
}

async function setPriority(priority: Priority) {
	const newTask: ITask = {
		...task.value,
		priority,
	}

	return saveTask(newTask)
}

async function setPercentDone(percentDone: number) {
	const newTask: ITask = {
		...task.value,
		percentDone,
	}

	return saveTask(newTask)
}

async function removeRepeatAfter() {
	task.value.repeatAfter.amount = 0
	task.value.repeatMode = TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT
	await saveTask()
}

function removeLabel(labelToRemove: ITask['labels'][number]) {
	task.value.labels = task.value.labels.filter(label => label.id !== labelToRemove.id)
	saveTask() // Save immediately after removing a label
}

function setRelatedTasksActive() {
	setFieldActive('relatedTasks')

	// If the related tasks are already available, show the form again
	const el = activeFieldElements['relatedTasks']
	for (const child in el?.children) {
		if (el?.children[child]?.id === 'showRelatedTasksFormButton') {
			el?.children[child]?.click()
			break
		}
	}
}
</script>

<style lang="scss" scoped>
.task-view-container {
	// simulate sass lighten($primary, 30) by increasing lightness 30% to 73%
	--primary-light: hsla(var(--primary-h), var(--primary-s), 73%, var(--primary-a));
	padding-bottom: 0;

	@media screen and (min-width: $desktop) {
		padding-bottom: 1rem;
	}
}

.task-view {
	padding-top: 1rem;
	padding-inline: .5rem;
	background-color: var(--site-background);

	@media screen and (min-width: $desktop) {
		padding: 1rem;
	}
}

.is-modal .task-view {
	border-radius: $radius;
	padding: 1rem;
	color: var(--text);
	background-color: var(--site-background) !important;

	@media screen and (max-width: calc(#{$desktop} + 1px)) {
		border-radius: 0;
	}
}

.task-view * {
	transition: opacity 50ms ease;
}

.is-loading .task-view * {
	opacity: 0;
}

// Jira-like layout specific styles
.jira-layout-container {
  display: flex;
  flex-wrap: wrap; // Allow sidebar to wrap on smaller screens if necessary
}

.jira-main-content {
  flex: 3; // Takes 3 parts of the space
  min-width: 0; // Prevents overflow issues
  padding-right: 1rem; // Space between main content and sidebar
}

.jira-sidebar {
  flex: 1; // Takes 1 part of the space
  min-width: 280px; // Minimum width for the sidebar
  // Styles for the sidebar itself, like background, padding can be added here
}

// Responsive adjustments for Jira layout
@media screen and (max-width: $tablet) { // Adjust breakpoint as needed
  .jira-layout-container {
    flex-direction: column;
  }
  .jira-main-content {
    padding-right: 0; // No space needed when stacked
    margin-bottom: 1rem; // Space before sidebar content when stacked
  }
  .jira-sidebar {
    min-width: 100%; // Full width on smaller screens
  }
}


.subtitle {
	color: var(--grey-500);
	margin-bottom: 1rem;

	a {
		color: var(--grey-800);
	}
}

h3 {
	margin-bottom: 0.75rem;
	padding-bottom: 0.5rem;
	border-bottom: 1px solid var(--grey-200);

	.button {
		vertical-align: middle;
	}
}

.icon.is-grey {
	color: var(--grey-400);
}

.date-input {
	display: flex;
	align-items: center;
}

.remove {
	color: var(--danger);
	vertical-align: middle;
	padding-left: .5rem;
	line-height: 1;
}

:deep(.datepicker) {
	width: 100%;

	.show {
		color: var(--text);
		padding: .25rem .5rem;
		transition: background-color $transition;
		border-radius: $radius;
		display: block;
		margin: .1rem 0;
		width: 100%;
		text-align: left;

		&:hover {
			background: var(--white);
		}
	}

	&.disabled .show:hover {
		background: transparent;
	}
}

// Adjusted .details for the new layout - less aggressive margin/padding if it's part of a larger section
.details {
	padding-bottom: 1rem;
	margin-bottom: 1rem; // Reduced margin for tighter layout
	// background-color: var(--scheme-main-bis); // This might be too much now, consider removing or adjusting
	// border-radius: $radius; // Keep if desired
	// padding: 1rem; // Keep if desired
	// box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05); // Keep if desired

	.detail-title {
		display: block;
		color: var(--grey-400);
		margin-bottom: 0.25rem; // Reduced margin for tighter layout
		font-weight: 500;
		font-size: 0.875rem; // Slightly smaller for sidebar
	}

	.none {
		font-style: italic;
	}
}

.content.description,
.content.attachments,
.content.related-tasks,
.comments-section,
.reactions-container {
	margin-top: 1rem;
	margin-bottom: 1.5rem;
	padding: 1rem;
	background-color: var(--scheme-main-bis);
	border-radius: $radius;
	box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.comments-section {
    margin-top: 1.5rem; // Keep specific margin for comments if needed
}


.details.labels-list,
.assignees {
	:deep(.multiselect) {
		.input-wrapper {
			&:not(:focus-within):not(:hover) {
				background: transparent;
				border-color: transparent;
			}
		}
	}
}

:deep(.details),
:deep(.heading) {
	.input:not(.has-defaults),
	.textarea,
	.select:not(.has-defaults) select {
		cursor: pointer;
		transition: all $transition-duration;

		&::placeholder {
			color: var(--text-light);
			opacity: 1;
			font-style: italic;
		}

		&:not(:disabled) {
			&:hover,
			&:active,
			&:focus {
				background: var(--scheme-main);
				border-color: var(--border);
				cursor: text;
			}

			&:hover,
			&:active {
				cursor: text;
				border-color: var(--link)
			}
		}
	}

	.select:not(.has-defaults):after {
		opacity: 0;
	}

	.select:not(.has-defaults):hover:after {
		opacity: 1;
	}
}

.attachments table tr:last-child td {
	border-bottom: none;
}

// Sidebar specific styles
.sidebar-top-actions {
  margin-bottom: 1rem;
  .task-status-button {
    width: 100%;
  }
}

.sidebar-details {
  // Styling for the details group in sidebar
  // Could add a border, padding, etc.
  // background-color: var(--scheme-main-ter);
  // border-radius: $radius;
  // padding: 0.75rem;
}

.sidebar-section-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-light);
  text-transform: uppercase;
  margin-bottom: 0.75rem;
  padding-bottom: 0.25rem;
  border-bottom: 1px solid var(--grey-200);
}

.sidebar-attribute-item {
  margin-bottom: 1rem;
  .detail-title {
    font-size: 0.8rem; 
    font-weight: 600;
    color: var(--grey-700);
    margin-bottom: 0.3rem;
    display: flex;
    align-items: center;
    .icon {
      margin-right: 0.3rem;
    }
  }
  // Make inputs and selects in sidebar more compact
  :deep(.select select), 
  :deep(.input) {
    font-size: 0.85rem;
    height: auto;
    padding-top: 0.3em;
    padding-bottom: 0.3em;
  }
  :deep(.datepicker .show) {
    font-size: 0.85rem;
    padding: 0.3rem 0.5rem;
  }
  :deep(.multiselect__tags), 
  :deep(.multiselect__input),
  :deep(.multiselect__single) {
    font-size: 0.85rem;
    padding-top: 3px;
    padding-bottom: 3px;
    min-height: auto;
  }
  :deep(.multiselect__tag) {
    padding: 3px 8px;
    margin-bottom: 3px;
    margin-right: 4px;
  }
  .reporter-info {
    font-size: 0.85rem;
    color: var(--text);
  }
}

.sidebar-actions {
  margin-top: 1.5rem;
  .sidebar-action-button {
    width: 100%;
    margin-bottom: .5rem;
    justify-content: left;
    font-size: 0.85rem;
    padding-top: 0.4em;
    padding-bottom: 0.4em;
    &.has-light-text {
      color: var(--white);
    }
  }
}

.sidebar-created-updated {
    margin-top: 1rem;
    font-size: 0.8rem;
    color: var(--grey-500);
     border-top: 1px solid var(--grey-200);
    padding-top: 0.75rem;
}


.checklist-summary {
	padding-left: .25rem;
}


// Removing old layout styles or styles that conflict
.action-buttons, .permanent-attribute-item, .task-permanent-attributes, .permanent-attribute-row, .permanent-attribute-cell {
    // These classes are from the old layout, their specific styles might not be needed 
    // or might conflict. Review and remove/adjust as necessary.
    // For now, let's ensure they don't apply conflicting flex/column styles by resetting some properties if they were previously used for layout.
    // This is a placeholder for more specific cleanup if needed.
}

.action-heading {
	display: none; // No longer used in the new layout
}


// Ensure deep styles from old layout are overridden or adjusted if necessary
// :deep(.task-permanent-attributes .select select), 
// :deep(.task-permanent-attributes .input) { ... }
// The above styles were very specific to the old .task-permanent-attributes container.
// New sidebar item styles should handle this now.

// Styles for the new main content tabs
.main-content-tabs-container {
  margin-top: 1.5rem;
  // Removing custom styles for .tabs ul, .tabs li.is-active a, .tabs a
  // to rely more on default Bulma styling for tabs is-boxed
  .tab-content {
    margin-top: 1rem; // Or 0 if tabs is-boxed provides enough separation
    padding: 1rem;
    background-color: var(--scheme-main-bis);
    border-radius: $radius;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  }
}

.action-edit-button {
  padding: 0.25em 0.5em; // Make edit/done buttons smaller
  font-size: 0.75rem; // Smaller font for edit/done
  color: var(--link);
  &:hover {
    color: var(--link-hover);
    background-color: var(--link-light);
  }
}

.labels-display-area {
  .tag {
    // Styles for displayed labels, if needed beyond Bulma defaults
    // Example: ensure consistent height if delete button makes them taller
    align-items: center;
    display: inline-flex;
    .delete {
      margin-left: 0.25rem;
    }
  }
  .is-italic {
    font-size: 0.85rem;
  }
}

.subtitle-container {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
  
  .subtitle {
    margin-bottom: 0;
    margin-right: 0.5rem;
  }
  
  .move-project-button {
    font-size: 0.75rem;
    padding: 0.25em 0.5em;
    color: var(--link);
    
    &:hover {
      color: var(--link-hover);
      background-color: var(--link-light);
    }
  }
}

.project-search-container {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
  
  :deep(.multiselect) {
    flex-grow: 1;
    margin-right: 0.5rem;
  }
  
  .cancel-move-button {
    font-size: 0.75rem;
    padding: 0.25em 0.75em;
  }
}

</style>