<template>
	<div
		class="loader-container task-view-container"
		:class="{
			'is-loading': taskService.loading || !visible,
			'is-modal': isModal,
		}"
	>
		<div class="task-view">
			<Heading
				:task="task"
				@update:task="Object.assign(task, $event)"
				:can-write="canWrite"
				ref="heading"
			/>
			<h6 class="subtitle" v-if="parent && parent.namespace && parent.list">
				{{ getNamespaceTitle(parent.namespace) }} &rsaquo;
				<router-link :to="{ name: 'list.index', params: { listId: parent.list.id } }">
					{{ getListTitle(parent.list) }}
				</router-link>
			</h6>

			<checklist-summary :task="task"/>

			<!-- Content and buttons -->
			<div class="columns mt-2">
				<!-- Content -->
				<div :class="{'is-two-thirds': canWrite}" class="column detail-content">
					<div class="columns details">
						<div class="column assignees" v-if="activeFields.assignees">
							<!-- Assignees -->
							<div class="detail-title">
								<icon icon="users"/>
								{{ $t('task.attributes.assignees') }}
							</div>
							<edit-assignees
								:disabled="!canWrite"
								:list-id="task.listId"
								:task-id="task.id"
								:ref="e => setFieldRef('assignees', e)"
								v-model="task.assignees"
							/>
						</div>
						<CustomTransition name="flash-background" appear>
							<div class="column" v-if="activeFields.priority">
								<!-- Priority -->
								<div class="detail-title">
									<icon icon="exclamation"/>
									{{ $t('task.attributes.priority') }}
								</div>
								<priority-select
									:disabled="!canWrite"
									@update:model-value="setPriority"
									:ref="e => setFieldRef('priority', e)"
									v-model="task.priority"/>
							</div>
						</CustomTransition>
						<CustomTransition name="flash-background" appear>
							<div class="column" v-if="activeFields.dueDate">
								<!-- Due Date -->
								<div class="detail-title">
									<icon icon="calendar"/>
									{{ $t('task.attributes.dueDate') }}
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.dueDate"
										@close-on-change="saveTask()"
										:choose-date-label="$t('task.detail.chooseDueDate')"
										:disabled="taskService.loading || !canWrite"
										:ref="e => setFieldRef('dueDate', e)"
									/>
									<BaseButton
										v-if="task.dueDate && canWrite"
										@click="() => {task.dueDate = null;saveTask()}"
										class="remove">
										<span class="icon is-small">
											<icon icon="times"></icon>
										</span>
									</BaseButton>
								</div>
							</div>
						</CustomTransition>
						<CustomTransition name="flash-background" appear>
							<div class="column" v-if="activeFields.percentDone">
								<!-- Progress -->
								<div class="detail-title">
									<icon icon="percent"/>
									{{ $t('task.attributes.percentDone') }}
								</div>
								<percent-done-select
									:disabled="!canWrite"
									@update:model-value="setPercentDone"
									:ref="e => setFieldRef('percentDone', e)"
									v-model="task.percentDone"/>
							</div>
						</CustomTransition>
						<CustomTransition name="flash-background" appear>
							<div class="column" v-if="activeFields.startDate">
								<!-- Start Date -->
								<div class="detail-title">
									<icon icon="play"/>
									{{ $t('task.attributes.startDate') }}
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.startDate"
										@close-on-change="saveTask()"
										:choose-date-label="$t('task.detail.chooseStartDate')"
										:disabled="taskService.loading || !canWrite"
										:ref="e => setFieldRef('startDate', e)"
									/>
									<BaseButton
										@click="() => {task.startDate = null;saveTask()}"
										v-if="task.startDate && canWrite"
										class="remove"
									>
										<span class="icon is-small">
											<icon icon="times"></icon>
										</span>
									</BaseButton>
								</div>
							</div>
						</CustomTransition>
						<CustomTransition name="flash-background" appear>
							<div class="column" v-if="activeFields.endDate">
								<!-- End Date -->
								<div class="detail-title">
									<icon icon="stop"/>
									{{ $t('task.attributes.endDate') }}
								</div>
								<div class="date-input">
									<datepicker
										v-model="task.endDate"
										@close-on-change="saveTask()"
										:choose-date-label="$t('task.detail.chooseEndDate')"
										:disabled="taskService.loading || !canWrite"
										:ref="e => setFieldRef('endDate', e)"
									/>
									<BaseButton
										@click="() => {task.endDate = null;saveTask()}"
										v-if="task.endDate && canWrite"
										class="remove">
										<span class="icon is-small">
											<icon icon="times"></icon>
										</span>
									</BaseButton>
								</div>
							</div>
						</CustomTransition>
						<CustomTransition name="flash-background" appear>
							<div class="column" v-if="activeFields.reminders">
								<!-- Reminders -->
								<div class="detail-title">
									<icon :icon="['far', 'clock']"/>
									{{ $t('task.attributes.reminders') }}
								</div>
								<reminders
									:disabled="!canWrite"
									:ref="e => setFieldRef('reminders', e)"
									v-model="task.reminderDates"
									@update:model-value="saveTask"
								/>
							</div>
						</CustomTransition>
						<CustomTransition name="flash-background" appear>
							<div class="column" v-if="activeFields.repeatAfter">
								<!-- Repeat after -->
								<div class="is-flex is-justify-content-space-between">
									<div class="detail-title">
										<icon icon="history"/>
										{{ $t('task.attributes.repeat') }}
									</div>
									<BaseButton
										@click="() => {task.repeatAfter.amount = 0;saveTask()}"
										v-if="canWrite"
										class="remove">
										<span class="icon is-small">
											<icon icon="times"></icon>
										</span>
									</BaseButton>
								</div>
								<repeat-after
									:disabled="!canWrite"
									:ref="e => setFieldRef('repeatAfter', e)"
									:model-value="task"
									@update:model-value="(newTask) => {
										Object.assign(task, newTask)
										saveTask()
									}"
								/>
							</div>
						</CustomTransition>
						<CustomTransition name="flash-background" appear>
							<div class="column" v-if="activeFields.color">
								<!-- Color -->
								<div class="detail-title">
									<icon icon="fill-drip"/>
									{{ $t('task.attributes.color') }}
								</div>
								<color-picker
									menu-position="bottom"
									:ref="e => setFieldRef('color', e)"
									v-model="taskColor"
									@update:model-value="saveTask"
								/>
							</div>
						</CustomTransition>
					</div>

					<!-- Labels -->
					<div class="labels-list details" v-if="activeFields.labels">
						<div class="detail-title">
							<span class="icon is-grey">
								<icon icon="tags"/>
							</span>
							{{ $t('task.attributes.labels') }}
						</div>
						<edit-labels
							:disabled="!canWrite" 
							:task-id="taskId"
							:ref="e => setFieldRef('labels', e)"
							v-model="task.labels"/>
					</div>

					<!-- Description -->
					<div class="details content description">
						<description
							:model-value="task"
							@update:modelValue="Object.assign(task, $event)"
							:can-write="canWrite"
							:attachment-upload="attachmentUpload"
						/>
					</div>

					<!-- Attachments -->
					<div class="content attachments" v-if="activeFields.attachments || hasAttachments">
						<attachments
							:edit-enabled="canWrite"
							:task="task"
							@task-changed="({coverImageAttachmentId}) => task.coverImageAttachmentId = coverImageAttachmentId"
							:ref="e => setFieldRef('attachments', e)"
						/>
					</div>

					<!-- Related Tasks -->
					<div class="content details mb-0" v-if="activeFields.relatedTasks">
						<h3>
							<span class="icon is-grey">
								<icon icon="sitemap"/>
							</span>
							{{ $t('task.attributes.relatedTasks') }}
						</h3>
						<related-tasks
							:edit-enabled="canWrite"
							:initial-related-tasks="task.relatedTasks"
							:list-id="task.listId"
							:show-no-relations-notice="true"
							:task-id="taskId"
							:ref="e => setFieldRef('relatedTasks', e)"
						/>
					</div>

					<!-- Move Task -->
					<div class="content details" v-if="activeFields.moveList">
						<h3>
							<span class="icon is-grey">
								<icon icon="list"/>
							</span>
							{{ $t('task.detail.move') }}
						</h3>
						<div class="field has-addons">
							<div class="control is-expanded">
								<list-search 
									@update:modelValue="changeList"
									:ref="e => setFieldRef('moveList', e)"
								/>
							</div>
						</div>
					</div>

					<!-- Comments -->
					<comments :can-write="canWrite" :task-id="taskId"/>
				</div>
				<div class="column is-one-third action-buttons d-print-none" v-if="canWrite || isModal">
					<template v-if="canWrite">
						<x-button
							:class="{'is-success': !task.done}"
							:shadow="task.done"
							@click="toggleTaskDone()"
							class="is-outlined has-no-border"
							icon="check-double"
							variant="secondary"
							v-shortcut="'t'"
						>
							{{ task.done ? $t('task.detail.undone') : $t('task.detail.done') }}
						</x-button>
						<task-subscription
							entity="task"
							:entity-id="task.id"
							:model-value="task.subscription"
							@update:model-value="sub => task.subscription = sub"
						/>
						<x-button
							@click="setFieldActive('assignees')"
							variant="secondary"
							v-shortcut="'a'"
							v-cy="'taskDetail.assign'"
						>
							<span class="icon is-small"><icon icon="users"/></span>
							{{ $t('task.detail.actions.assign') }}
						</x-button>
						<x-button
							@click="setFieldActive('labels')"
							variant="secondary"
							icon="tags"
							v-shortcut="'l'"
						>
							{{ $t('task.detail.actions.label') }}
						</x-button>
						<x-button
							@click="setFieldActive('priority')"
							variant="secondary"
							icon="exclamation"
						>
							{{ $t('task.detail.actions.priority') }}
						</x-button>
						<x-button
							@click="setFieldActive('dueDate')"
							variant="secondary"
							icon="calendar"
							v-shortcut="'d'"
						>
							{{ $t('task.detail.actions.dueDate') }}
						</x-button>
						<x-button
							@click="setFieldActive('startDate')"
							variant="secondary"
							icon="play"
						>
							{{ $t('task.detail.actions.startDate') }}
						</x-button>
						<x-button
							@click="setFieldActive('endDate')"
							variant="secondary"
							icon="stop"
						>
							{{ $t('task.detail.actions.endDate') }}
						</x-button>
						<x-button
							@click="setFieldActive('reminders')"
							variant="secondary"
							:icon="['far', 'clock']"
							v-shortcut="'Alt+r'"
						>
							{{ $t('task.detail.actions.reminders') }}
						</x-button>
						<x-button
							@click="setFieldActive('repeatAfter')"
							variant="secondary"
							icon="history"
						>
							{{ $t('task.detail.actions.repeatAfter') }}
						</x-button>
						<x-button
							@click="setFieldActive('percentDone')"
							variant="secondary"
							icon="percent"
						>
							{{ $t('task.detail.actions.percentDone') }}
						</x-button>
						<x-button
							@click="setFieldActive('attachments')"
							variant="secondary"
							icon="paperclip"
							v-shortcut="'f'"
						>
							{{ $t('task.detail.actions.attachments') }}
						</x-button>
						<x-button
							@click="setFieldActive('relatedTasks')"
							variant="secondary"
							icon="sitemap"
							v-shortcut="'r'"
						>
							{{ $t('task.detail.actions.relatedTasks') }}
						</x-button>
						<x-button
							@click="setFieldActive('moveList')"
							variant="secondary"
							icon="list"
							v-shortcut="'m'"
						>
							{{ $t('task.detail.actions.moveList') }}
						</x-button>
						<x-button
							@click="setFieldActive('color')"
							variant="secondary"
							icon="fill-drip"
							:icon-color="color"
							v-shortcut="'c'"
						>
							{{ $t('task.detail.actions.color') }}
						</x-button>
						<x-button
							@click="toggleFavorite"
							variant="secondary"
							:icon="task.isFavorite ? 'star' : ['far', 'star']"
						>
							{{
								task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')
							}}
						</x-button>
						<x-button
							@click="showDeleteModal = true"
							icon="trash-alt"
							:shadow="false"
							class="is-danger is-outlined has-no-border"
						>
							{{ $t('task.detail.actions.delete') }}
						</x-button>
					</template>

					<!-- Created / Updated [by] -->
					<created-updated :task="task"/>
				</div>
			</div>
			<!-- Created / Updated [by] -->
			<created-updated :task="task" v-if="!canWrite && !isModal"/>
		</div>

		<modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteTask()"
		>
			<template #header><span>{{ $t('task.detail.delete.header') }}</span></template>

			<template #text>
				<p>{{ $t('task.detail.delete.text1') }}<br/>
					{{ $t('task.detail.delete.text2') }}</p>
			</template>
		</modal>
	</div>
</template>

<script lang="ts" setup>
import {ref, reactive, toRef, shallowReactive, computed, watch, nextTick, type PropType} from 'vue'
import {useRouter, type RouteLocation} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {unrefElement} from '@vueuse/core'
import {klona} from 'klona/lite'

import TaskService from '@/services/task'
import TaskModel, {TASK_DEFAULT_COLOR} from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'
import type {IList} from '@/modelTypes/IList'

import {PRIORITIES, type Priority} from '@/constants/priorities'
import {RIGHTS} from '@/constants/rights'

import BaseButton from '@/components/base/BaseButton.vue'

// partials
import Attachments from '@/components/tasks/partials/attachments.vue'
import ChecklistSummary from '@/components/tasks/partials/checklist-summary.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import Comments from '@/components/tasks/partials/comments.vue'
import CreatedUpdated from '@/components/tasks/partials/createdUpdated.vue'
import Datepicker from '@/components/input/datepicker.vue'
import Description from '@/components/tasks/partials/description.vue'
import EditAssignees from '@/components/tasks/partials/editAssignees.vue'
import EditLabels from '@/components/tasks/partials/editLabels.vue'
import Heading from '@/components/tasks/partials/heading.vue'
import ListSearch from '@/components/tasks/partials/listSearch.vue'
import PercentDoneSelect from '@/components/tasks/partials/percentDoneSelect.vue'
import PrioritySelect from '@/components/tasks/partials/prioritySelect.vue'
import RelatedTasks from '@/components/tasks/partials/relatedTasks.vue'
import Reminders from '@/components/tasks/partials/reminders.vue'
import RepeatAfter from '@/components/tasks/partials/repeatAfter.vue'
import TaskSubscription from '@/components/misc/subscription.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'

import {uploadFile} from '@/helpers/attachments'
import {getNamespaceTitle} from '@/helpers/getNamespaceTitle'
import {getListTitle} from '@/helpers/getListTitle'
import {scrollIntoView} from '@/helpers/scrollIntoView'

import {useBaseStore} from '@/stores/base'
import {useNamespaceStore} from '@/stores/namespaces'
import {useAttachmentStore} from '@/stores/attachments'
import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'

import {useTitle} from '@/composables/useTitle'

import {success} from '@/message'

const props = defineProps({
	taskId: {
		type: Number as PropType<ITask['id']>,
		required: true,
	},
	backdropView: {
		type: String as PropType<RouteLocation['fullPath']>,
	},
})

defineEmits(['close'])

const router = useRouter()
const {t} = useI18n({useScope: 'global'})

const baseStore = useBaseStore()
const namespaceStore = useNamespaceStore()
const attachmentStore = useAttachmentStore()
const taskStore = useTaskStore()
const kanbanStore = useKanbanStore()

const task = reactive<ITask>(new TaskModel())
useTitle(toRef(task, 'title'))

// We doubled the task color property here because verte does not have a real change property, leading
// to the color property change being triggered when the # is removed from it, leading to an update,
// which leads in turn to a change... This creates an infinite loop in which the task is updated, changed,
// updated, changed, updated and so on.
// To prevent this, we put the task color property in a seperate value which is set to the task color
// when it is saved and loaded.
const taskColor = ref<ITask['hexColor']>('')

// Used to avoid flashing of empty elements if the task content is not yet loaded.
const visible = ref(false)

const taskId = toRef(props, 'taskId')

const parent = computed(() => {
	if (!task.listId) {
		return {
			namespace: null,
			list: null,
		}
	}

	if (!namespaceStore.getListAndNamespaceById) {
		return null
	}

	return namespaceStore.getListAndNamespaceById(task.listId)
})

watch(
	parent,
	(parent) => {
		const parentList = parent !== null ? parent.list : null
		if (parentList !== null) {
			baseStore.handleSetCurrentList({list: parentList})
		}
	},
	{immediate: true},
)

const canWrite = computed(() => (
	task.maxRight !== null &&
	task.maxRight > RIGHTS.READ
))

const color = computed(() => {
	const color = task.getHexColor
		? task.getHexColor()
		: undefined
	
	return color === TASK_DEFAULT_COLOR
		? ''
		: color
})

const hasAttachments = computed(() => attachmentStore.attachments.length > 0)

const isModal = computed(() => Boolean(props.backdropView))

function attachmentUpload(file: File, onSuccess?: (url: string) => void) {
	return uploadFile(taskId.value, file, onSuccess)
}

const heading = ref<HTMLElement | null>(null)
async function scrollToHeading() {
	scrollIntoView(unrefElement(heading))
}

const taskService = shallowReactive(new TaskService())

// load task
watch(taskId, async (id) => {
	if (id === undefined) {
		return
	}

	try {
		Object.assign(task, await taskService.get({id}))
		attachmentStore.set(task.attachments)
		taskColor.value = task.hexColor
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
	| 'color'
	| 'dueDate'
	| 'endDate'
	| 'labels'
	| 'moveList'
	| 'percentDone'
	| 'priority'
	| 'relatedTasks'
	| 'reminders'
	| 'repeatAfter'
	| 'startDate'

const activeFields : {[type in FieldType]: boolean} = reactive({
	assignees: false,
	attachments: false,
	color: false,
	dueDate: false,
	endDate: false,
	labels: false,
	moveList: false,
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
	activeFields.assignees = task.assignees.length > 0
	activeFields.attachments = task.attachments.length > 0
	activeFields.dueDate = task.dueDate !== null
	activeFields.endDate = task.endDate !== null
	activeFields.labels = task.labels.length > 0
	activeFields.percentDone = task.percentDone > 0
	activeFields.priority = task.priority !== PRIORITIES.UNSET
	activeFields.relatedTasks = Object.keys(task.relatedTasks).length > 0
	activeFields.reminders = task.reminderDates.length > 0
	activeFields.repeatAfter = task.repeatAfter.amount > 0
	activeFields.startDate = task.startDate !== null
}

const activeFieldElements : {[id in FieldType]: HTMLElement | null} = reactive({
	assignees: null,
	attachments: null,
	color: null,
	dueDate: null,
	endDate: null,
	labels: null,
	moveList: null,
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

async function saveTask(args?: {
	task: ITask,
	undoCallback?: () => void,
}) {
	const {
		task: currentTask,
		undoCallback,
	} = {
		...{
			task: klona(task),
		},
		...args,
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

	const newTask = await taskStore.update(currentTask) // TODO: markraw ?
	Object.assign(task, newTask)
	setActiveFields()

	let actions = []
	if (undoCallback !== null) {
		actions = [{
			title: 'Undo',
			callback: undoCallback,
		}]
	}
	success({message: t('task.detail.updateSuccess')}, actions)
}

const showDeleteModal = ref(false)
async function deleteTask() {
	await taskStore.delete(task)
	success({message: t('task.detail.deleteSuccess')})
	router.push({name: 'list.index', params: {listId: task.listId}})
}

function toggleTaskDone() {
	const newTask = {
		...task,
		done: !task.done,
	}

	saveTask({
		task: newTask,
		undoCallback: toggleTaskDone,
	})
}

async function changeList(list: IList) {
	kanbanStore.removeTaskInBucket(task)
	await saveTask({
		task: {
			...task,
			listId: list.id,
		},
	})
}

async function toggleFavorite() {
	task.isFavorite = !task.isFavorite
	const newTask = await taskService.update(task)
	Object.assign(task, newTask)
	await namespaceStore.loadNamespacesIfFavoritesDontExist()
}

async function setPriority(priority: Priority) {
	const newTask: ITask = {
		...task,
		priority,
	}

	return saveTask({
		task: newTask,
	})
}

async function setPercentDone(percentDone: number) {
	const newTask: ITask = {
		...task,
		percentDone,
	}

	return saveTask({
		task: newTask,
	})
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

  @media screen and (max-width: 800px) {
    border-radius: 0;
    padding-top: 2rem;
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
	margin-bottom: 1rem;

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

.details {
	padding-bottom: 0.75rem;
	flex-flow: row wrap;
	margin-bottom: 0;

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

.attachments {
	margin-bottom: 0;

	table tr:last-child td {
		border-bottom: none;
	}
}

.action-buttons {
	@media screen and (min-width: $tablet) {
		position: sticky;
		top: $navbar-height + 1.5rem;
		align-self: flex-start;
	}
	
	.button {
		width: 100%;
		margin-bottom: .5rem;
		justify-content: left; 

		&.has-light-text {
			color: var(--white);
		}
	}
}

.is-modal .action-buttons {
	// we need same top margin for the modal close button 
	@media screen and (min-width: $tablet) {
		top: 6.5rem;
	}
	// this is the moment when the fixed close button is outside the modal
	// => we can fill up the space again
	@media screen and (min-width: calc(#{$desktop} + 84px)) {
		top: 0;
	}
}

.checklist-summary {
	padding-left: .25rem;
}

.detail-content {
  @media print {
		width: 100% !important;
  }
}
</style>