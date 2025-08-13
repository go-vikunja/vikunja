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
			class="task-view"
		>
			<Heading
				ref="heading"
				:task="task"
				:can-write="canWrite"
				:has-close="isModal"
				@update:task="Object.assign(task, $event)"
				@close="$emit('close')"
			/>
			<h6
				v-if="project?.id"
				class="subtitle"
			>
				<template
					v-for="p in projectStore.getAncestors(project)"
					:key="p.id"
				>
					<a
						v-if="router.options.history.state.back?.includes('/projects/'+p.id+'/') || false"
						v-shortcut="p.id === project?.id ? 'u' : ''"
						@click="router.back()"
					>
						{{ getProjectTitle(p) }}
					</a>
					<RouterLink
						v-else
						v-shortcut="p.id === project?.id ? 'u' : ''"
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

			<ChecklistSummary :task="task" />

			<!-- Content and buttons -->
			<div class="columns mbs-2">
				<!-- Content -->
				<div
					:class="{'is-two-thirds': canWrite}"
					class="column detail-content"
				>
					<div class="columns details">
						<div
							v-if="activeFields.assignees"
							class="column assignees"
						>
							<!-- Assignees -->
							<div class="detail-title">
								<Icon icon="users" />
								{{ $t('task.attributes.assignees') }}
							</div>
							<EditAssignees
								v-if="canWrite"
								:ref="e => setFieldRef('assignees', e)"
								v-model="task.assignees"
								:project-id="task.projectId"
								:task-id="task.id"
							/>
							<AssigneeList
								v-else
								:assignees="task.assignees"
								class="mbs-2"
							/>
						</div>
						<CustomTransition
							name="flash-background"
							appear
						>
							<div
								v-if="activeFields.priority"
								class="column"
							>
								<!-- Priority -->
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
						</CustomTransition>
						<CustomTransition
							name="flash-background"
							appear
						>
							<div
								v-if="activeFields.dueDate"
								class="column"
							>
								<!-- Due Date -->
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
						</CustomTransition>
						<CustomTransition
							name="flash-background"
							appear
						>
							<div
								v-if="activeFields.percentDone"
								class="column"
							>
								<!-- Progress -->
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
						</CustomTransition>
						<CustomTransition
							name="flash-background"
							appear
						>
							<div
								v-if="activeFields.startDate"
								class="column"
							>
								<!-- Start Date -->
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
						</CustomTransition>
						<CustomTransition
							name="flash-background"
							appear
						>
							<div
								v-if="activeFields.endDate"
								class="column"
							>
								<!-- End Date -->
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
						</CustomTransition>
						<CustomTransition
							name="flash-background"
							appear
						>
							<div
								v-if="activeFields.reminders"
								class="column"
							>
								<!-- Reminders -->
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
							</div>
						</CustomTransition>
						<CustomTransition
							name="flash-background"
							appear
						>
							<div
								v-if="activeFields.repeatAfter"
								class="column"
							>
								<!-- Repeat after -->
								<div class="is-flex is-justify-content-space-between">
									<div class="detail-title">
										<Icon icon="history" />
										{{ $t('task.attributes.repeat') }}
									</div>
									<BaseButton
										v-if="canWrite"
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
							</div>
						</CustomTransition>
						<CustomTransition
							name="flash-background"
							appear
						>
							<div
								v-if="activeFields.color"
								class="column"
							>
								<!-- Color -->
								<div class="detail-title">
									<Icon icon="fill-drip" />
									{{ $t('task.attributes.color') }}
								</div>
								<ColorPicker
									:ref="e => setFieldRef('color', e)"
									v-model="taskColor"
									menu-position="bottom"
									@update:modelValue="saveTask()"
								/>
							</div>
						</CustomTransition>
					</div>

					<!-- Labels -->
					<div
						v-if="activeFields.labels"
						class="labels-list details"
					>
						<div class="detail-title">
							<span class="icon is-grey">
								<Icon icon="tags" />
							</span>
							{{ $t('task.attributes.labels') }}
						</div>
						<EditLabels
							:ref="e => setFieldRef('labels', e)"
							v-model="task.labels"
							:disabled="!canWrite"
							:task-id="taskId"
							:creatable="!authStore.isLinkShareAuth"
						/>
					</div>

					<!-- Description -->
					<div class="details content description">
						<Description
							:model-value="task"
							:can-write="canWrite"
							:attachment-upload="attachmentUpload"
							@update:modelValue="Object.assign(task, $event)"
						/>
					</div>
					
					<!-- Reactions -->
					<Reactions 
						v-model="task.reactions" 
						entity-kind="tasks"
						:entity-id="task.id"
						class="details"
						:disabled="!canWrite"
					/>

					<!-- Attachments -->
					<div
						v-if="activeFields.attachments || hasAttachments"
						class="content attachments"
					>
						<Attachments
							:ref="e => setFieldRef('attachments', e)"
							:edit-enabled="canWrite"
							:task="task"
							@taskChanged="({coverImageAttachmentId}) => task.coverImageAttachmentId = coverImageAttachmentId"
						/>
					</div>

					<!-- Related Tasks -->
					<div
						v-if="activeFields.relatedTasks"
						class="content details mbe-0"
					>
						<h3>
							<span class="icon is-grey">
								<Icon icon="sitemap" />
							</span>
							{{ $t('task.attributes.relatedTasks') }}
						</h3>
						<RelatedTasks
							:ref="e => setFieldRef('relatedTasks', e)"
							:edit-enabled="canWrite"
							:initial-related-tasks="task.relatedTasks"
							:project-id="task.projectId"
							:show-no-relations-notice="true"
							:task-id="taskId"
						/>
					</div>

					<!-- Move Task -->
					<div
						v-if="activeFields.moveProject"
						class="content details"
					>
						<h3>
							<span class="icon is-grey">
								<Icon icon="list" />
							</span>
							{{ $t('task.detail.move') }}
						</h3>
						<div class="field has-addons">
							<div class="control is-expanded">
								<ProjectSearch
									:ref="e => setFieldRef('moveProject', e)"
									:filter="project => project.id !== task.projectId"
									@update:modelValue="changeProject"
								/>
							</div>
						</div>
					</div>

					<!-- Comments -->
					<Comments
						:can-write="canWrite"
						:task-id="taskId"
						:initial-comments="task.comments"
					/>
				</div>
				
				<!-- Task Actions -->
				<div
					v-if="canWrite || isModal"
					class="column is-one-third action-buttons d-print-none"
				>
					<template v-if="canWrite">
						<XButton
							v-shortcut="'t'"
							:class="{'is-success': !task.done}"
							:shadow="task.done"
							class="is-outlined has-no-border"
							icon="check-double"
							variant="secondary"
							@click="toggleTaskDone()"
						>
							{{ task.done ? $t('task.detail.undone') : $t('task.detail.done') }}
						</XButton>
						<TaskSubscription
							entity="task"
							:entity-id="task.id"
							:model-value="task.subscription"
							@update:modelValue="sub => task.subscription = sub"
						/>
						<XButton
							v-shortcut="'s'"
							variant="secondary"
							:icon="task.isFavorite ? 'star' : ['far', 'star']"
							@click="toggleFavorite"
						>
							{{
								task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')
							}}
						</XButton>
						
						<span class="action-heading">{{ $t('task.detail.organization') }}</span>
						
						<XButton
							v-shortcut="'l'"
							variant="secondary"
							icon="tags"
							@click="setFieldActive('labels')"
						>
							{{ $t('task.detail.actions.label') }}
						</XButton>
						<XButton
							v-shortcut="'p'"
							variant="secondary"
							icon="exclamation-circle"
							@click="setFieldActive('priority')"
						>
							{{ $t('task.detail.actions.priority') }}
						</XButton>
						<XButton
							variant="secondary"
							icon="percent"
							@click="setFieldActive('percentDone')"
						>
							{{ $t('task.detail.actions.percentDone') }}
						</XButton>
						<XButton
							v-shortcut="'c'"
							variant="secondary"
							icon="fill-drip"
							:icon-color="color"
							@click="setFieldActive('color')"
						>
							{{ $t('task.detail.actions.color') }}
						</XButton>
						
						<span class="action-heading">{{ $t('task.detail.management') }}</span>

						<XButton
							v-shortcut="'a'"
							v-cy="'taskDetail.assign'"
							variant="secondary"
							icon="users"
							@click="setFieldActive('assignees')"
						>
							{{ $t('task.detail.actions.assign') }}
						</XButton>
						<XButton
							v-shortcut="'f'"
							variant="secondary"
							icon="paperclip"
							@click="setFieldActive('attachments')"
						>
							{{ $t('task.detail.actions.attachments') }}
						</XButton>
						<XButton
							v-shortcut="'r'"
							variant="secondary"
							icon="sitemap"
							@click="setRelatedTasksActive()"
						>
							{{ $t('task.detail.actions.relatedTasks') }}
						</XButton>
						<XButton
							v-shortcut="'m'"
							variant="secondary"
							icon="list"
							@click="setFieldActive('moveProject')"
						>
							{{ $t('task.detail.actions.moveProject') }}
						</XButton>
						
						<span class="action-heading">{{ $t('task.detail.dateAndTime') }}</span>
						
						<XButton
							v-shortcut="'d'"
							variant="secondary"
							icon="calendar"
							@click="setFieldActive('dueDate')"
						>
							{{ $t('task.detail.actions.dueDate') }}
						</XButton>
						<XButton
							variant="secondary"
							icon="play"
							@click="setFieldActive('startDate')"
						>
							{{ $t('task.detail.actions.startDate') }}
						</XButton>
						<XButton
							variant="secondary"
							icon="stop"
							@click="setFieldActive('endDate')"
						>
							{{ $t('task.detail.actions.endDate') }}
						</XButton>
						<XButton
							v-shortcut="'Alt+r'"
							variant="secondary"
							:icon="['far', 'clock']"
							@click="setFieldActive('reminders')"
						>
							{{ $t('task.detail.actions.reminders') }}
						</XButton>
						<XButton
							variant="secondary"
							icon="history"
							@click="setFieldActive('repeatAfter')"
						>
							{{ $t('task.detail.actions.repeatAfter') }}
						</XButton>
						<XButton
							v-shortcut="'Shift+Delete'"
							icon="trash-alt"
							:shadow="false"
							class="is-danger is-outlined has-no-border"
							@click="showDeleteModal = true"
						>
							{{ $t('task.detail.actions.delete') }}
						</XButton>
					</template>

					<!-- Created / Updated [by] -->
					<CreatedUpdated :task="task" />
				</div>
			</div>
			<!-- Created / Updated [by] -->
			<CreatedUpdated
				v-if="!canWrite && !isModal"
				:task="task"
			/>
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
				<p class="tw-text-balance">
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
import {PERMISSIONS} from '@/constants/permissions'

import BaseButton from '@/components/base/BaseButton.vue'

// partials
import Attachments from '@/components/tasks/partials/Attachments.vue'
import ChecklistSummary from '@/components/tasks/partials/ChecklistSummary.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
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

// See https://github.com/github/hotkey/discussions/85#discussioncomment-5214660
function saveTaskViaHotkey(event) {
	const hotkeyString = eventToHotkeyString(event)
	if (!hotkeyString) return
	if (hotkeyString !== 'Control+s' && hotkeyString !== 'Meta+s') return
	event.preventDefault()

	saveTask()
}

onMounted(() => {
	document.addEventListener('keydown', saveTaskViaHotkey)
})

onBeforeUnmount(() => {
	document.removeEventListener('keydown', saveTaskViaHotkey)
})

// We doubled the task color property here because verte does not have a real change property, leading
// to the color property change being triggered when the # is removed from it, leading to an update,
// which leads in turn to a change... This creates an infinite loop in which the task is updated, changed,
// updated, changed, updated and so on.
// To prevent this, we put the task color property in a separate value which is set to the task color
// when it is saved and loaded.
const taskColor = ref<ITask['hexColor']>('')

// Used to avoid flashing of empty elements if the task content is not yet loaded.
const visible = ref(false)

const project = computed(() => projectStore.projects[task.value.projectId])

const canWrite = computed(() => (
	task.value.maxPermission !== null &&
	task.value.maxPermission > PERMISSIONS.READ
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
			taskColor.value = task.value.hexColor
			setActiveFields()
		} catch (e) {
			if (e?.response?.status === 404) {
				router.replace({name: 'not-found'})
				return
			}

			throw e
		} finally {
			await nextTick()
			scrollToHeading()
			visible.value = true
		}
	}, {immediate: true})

type FieldType =
	| 'assignees'
	| 'attachments'
	| 'color'
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
	color: false,
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
	color: null,
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

	currentTask.hexColor = taskColor.value

	// If no end date is being set, but a start date and due date,
	// use the due date as the end date
	if (
		currentTask.endDate === null &&
		currentTask.startDate !== null &&
		currentTask.dueDate !== null
	) {
		currentTask.endDate = currentTask.dueDate
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
	padding-block-end: 0;

	@media screen and (min-width: $desktop) {
		padding-block-end: 1rem;
	}
}

.task-view {
	padding-block-start: 1rem;
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

	@media screen and (width <= calc(#{$desktop} + 1px)) {
		border-radius: 0;
	}
}

.task-view * {
	transition: opacity 50ms ease;
}

.is-loading .task-view * {
	opacity: 0;
}


.subtitle {
	color: var(--grey-500);
	margin-block-end: 1rem;

	a {
		color: var(--grey-800);
	}
}

h3 .button {
	vertical-align: middle;
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
	padding-inline-start: .5rem;
	line-height: 1;
}

:deep(.datepicker) {
	inline-size: 100%;

	.show {
		color: var(--text);
		padding: .25rem .5rem;
		transition: background-color $transition;
		border-radius: $radius;
		display: block;
		margin: .1rem 0;
		inline-size: 100%;
		text-align: start;

		&:hover {
			background: var(--white);
		}
	}

	&.disabled .show:hover {
		background: transparent;
	}
}

.details {
	padding-block-end: 0.75rem;
	flex-flow: row wrap;
	margin-block-end: 0;

	.detail-title {
		display: block;
		color: var(--grey-400);
	}

	.none {
		font-style: italic;
	}

	// Break after the 2nd element
	.column:nth-child(2n) {
		page-break-after: always; // CSS 2.1 syntax
		break-after: always; // New syntax
	}

}

.details.labels-list,
.assignees {
	:deep(.multiselect) {
		.input-wrapper {
			&:not(:focus-within, :hover) {
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

.attachments {
	margin-block-end: 0;

	table tr:last-child td {
		border-inline-end: none;
	}
}

.action-buttons {
	@media screen and (min-width: $tablet) {
		position: sticky;
		inset-block-start: $navbar-height + 1.5rem;
		align-self: flex-start;
	}

	.button {
		inline-size: 100%;
		margin-block-end: .5rem;
		justify-content: left;

		&.has-light-text {
			color: var(--white);
		}
	}
}

.is-modal .action-buttons {
	// we need same top margin for the modal close button 
	@media screen and (min-width: $tablet) {
		inset-block-start: 6.5rem;
	}
	// this is the moment when the fixed close button is outside the modal
	// => we can fill up the space again
	@media screen and (width >= calc(#{$desktop} + 84px)) {
		inset-block-start: 0;
	}
}

.checklist-summary {
	padding-inline-start: .25rem;
}

.detail-content {
	@media print {
		inline-size: 100% !important;
	}
}

.action-heading {
	text-transform: uppercase;
	color: var(--grey-700);
	font-size: .75rem;
	font-weight: 700;
	margin: .5rem 0;
	display: inline-block;
}
</style>
