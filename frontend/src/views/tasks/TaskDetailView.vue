<template>
	<div
		ref="taskViewContainer"
		tabindex="-1"
		class="loader-container task-view-container"
		:class="{
			'is-loading': taskBusy || (!visible && !hasTrashNotice),
			'is-modal': isModal,
		}"
	>
		<div
			v-if="hasTrashNotice"
			class="task-view"
		>
			<BaseButton
				v-if="isModal"
				class="is-pulled-end"
				:aria-label="$t('misc.close')"
				@click="$emit('close')"
			>
				<Icon icon="times" />
			</BaseButton>
			<TaskInTrashNotice
				v-if="trashedTask"
				:in-trash="trashedTask"
			/>
			<ProjectInTrashNotice
				v-else-if="trashedProject"
				:in-trash="trashedProject"
			/>
		</div>
		<!-- Removing everything until the task is loaded to prevent empty initialization of other components -->
		<div
			v-else-if="visible"
			class="task-view"
		>
			<BaseButton
				v-if="!isModal"
				class="back-button mbs-2"
				@click="goBack()"
			>
				<Icon icon="arrow-left" />
				{{ $t('task.detail.back') }}
			</BaseButton>
			<Heading
				ref="heading"
				:task="task"
				:can-write="canWrite"
				:has-close="isModal"
				@close="$emit('close')"
			/>
			<nav
				v-if="project?.id"
				aria-label="Breadcrumb"
				class="subtitle"
			>
				<template
					v-for="p in projectList.getAncestors(project)"
					:key="p.id"
				>
					<a
						v-if="typeof router.options.history.state.back === 'string' && router.options.history.state.back.includes('/projects/'+p.id+'/')"
						v-shortcut="p.id === project?.id ? SHORTCUTS.taskDetail.openProject : ''"
						@click="router.back()"
					>
						{{ getProjectTitle(p) }}
					</a>
					<RouterLink
						v-else
						v-shortcut="p.id === project?.id ? SHORTCUTS.taskDetail.openProject : ''"
						:to="{ name: 'project.index', params: { projectId: p.id } }"
					>
						{{ getProjectTitle(p) }}
					</RouterLink>
					<span
						v-if="p.id !== project?.id"
						class="has-text-grey-light"
					> &gt; </span>
				</template>
				<BucketSelect
					:task="task"
					:can-write="canWrite"
				/>
			</nav>

			<ChecklistSummary :task="task" />

			<div class="columns task-columns mbs-2">
				<div class="column is-two-thirds detail-content">
					<!-- Description -->
					<div class="details content description">
						<Description
							:model-value="task"
							:can-write="canWrite"
						/>
					</div>

					<!-- Reactions -->
					<Reactions
						:model-value="task.reactions"
						entity-kind="tasks"
						:entity-id="task.id"
						class="details d-print-none"
						:disabled="!canWrite"
					/>

					<!-- Attachments -->
					<div
						v-show="activeSections.attachments || hasAttachments"
						class="content attachments"
					>
						<Attachments
							:ref="e => { setFieldRef('attachments', e); attachmentsRef = e as any }"
							:edit-enabled="canWrite"
							:task="task"
						/>
					</div>

					<!-- Time Tracking -->
					<div
						v-if="timeTrackingEnabled && activeSections.timeTracking"
						:ref="e => setFieldRef('timeTracking', e)"
						class="content time-tracking"
					>
						<TaskTimeTracking :task-id="task.id" />
					</div>

					<!-- Related Tasks -->
					<div
						v-if="activeSections.relatedTasks"
						class="content details mbe-0"
					>
						<h2 class="task-section-title">
							<span class="icon is-grey">
								<Icon icon="sitemap" />
							</span>
							{{ $t('task.attributes.relatedTasks') }}
						</h2>
						<RelatedTasks
							:ref="e => setFieldRef('relatedTasks', e)"
							:edit-enabled="canWrite"
							:initial-related-tasks="task.related_tasks"
							:project-id="task.project_id"
							:show-no-relations-notice="true"
							:task-id="taskId"
						/>
					</div>

					<!-- Move Task -->
					<div
						v-if="activeSections.moveProject"
						class="content details"
					>
						<h2 class="task-section-title">
							<span class="icon is-grey">
								<Icon icon="list" />
							</span>
							{{ $t('task.detail.move') }}
						</h2>
						<div class="field has-addons">
							<div class="control is-expanded">
								<ProjectSearch
									:ref="e => setFieldRef('moveProject', e)"
									:filter="project => project.id !== task.project_id"
									@update:modelValue="changeProject"
								/>
							</div>
						</div>
					</div>

					<!-- Comments -->
					<Comments
						:can-write="canWrite"
						:task-id="taskId"
						:project-id="task.project_id"
					/>

					<!-- Marker element for scroll-to-bottom button visibility -->
					<div
						ref="contentBottomMarker"
						class="content-bottom-marker"
					/>
				</div>

				<aside class="column is-one-third action-buttons task-sidebar">
					<div
						v-if="canWrite"
						class="sidebar-toolbar d-print-none"
					>
						<XButton
							v-shortcut="SHORTCUTS.taskDetail.done"
							:class="{'is-pending': !task.done}"
							class="button--mark-done"
							icon="check-double"
							variant="secondary"
							@click="toggleTaskDone()"
						>
							{{ task.done ? $t('task.detail.undone') : $t('task.detail.done') }}
						</XButton>
						<BaseButton
							v-shortcut="SHORTCUTS.taskDetail.favorite"
							v-tooltip="favoriteLabel"
							class="toolbar-icon-button"
							:class="{'is-favorite': task.is_favorite}"
							:aria-label="favoriteLabel"
							:aria-pressed="Boolean(task.is_favorite)"
							@click="toggleFavorite"
						>
							<Icon :icon="task.is_favorite ? 'star' : ['far', 'star']" />
						</BaseButton>
						<Dropdown
							class="task-actions-menu"
							:trigger-label="$t('task.detail.moreActions')"
						>
							<template #trigger="{toggleOpen, open}">
								<BaseButton
									v-tooltip="$t('task.detail.moreActions')"
									v-cy="'taskDetail.moreActions'"
									class="toolbar-icon-button"
									:aria-label="$t('task.detail.moreActions')"
									:aria-expanded="open"
									@click="toggleOpen"
								>
									<Icon icon="ellipsis-h" />
								</BaseButton>
							</template>
							<template #default="{close}">
								<div @click="close">
									<TaskSubscription
										entity="task"
										type="dropdown"
										:entity-id="task.id"
										:model-value="task.subscription ?? null"
										@toggle="toggleSubscription"
									/>
									<DropdownItem
										icon="paperclip"
										@click="openAttachments()"
									>
										{{ $t('task.detail.actions.attachments') }}
									</DropdownItem>
									<DropdownItem
										icon="sitemap"
										@click="setRelatedTasksActive()"
									>
										{{ $t('task.detail.actions.relatedTasks') }}
									</DropdownItem>
									<DropdownItem
										v-if="timeTrackingEnabled"
										v-cy="'taskTrackTimeAction'"
										:icon="['far', 'clock']"
										@click="setFieldActive('timeTracking')"
									>
										{{ $t('task.detail.actions.timeTracking') }}
									</DropdownItem>
									<DropdownItem
										icon="list"
										@click="setFieldActive('moveProject')"
									>
										{{ $t('task.detail.actions.moveProject') }}
									</DropdownItem>
									<DropdownItem
										icon="copy"
										@click="duplicateCurrentTask"
									>
										{{ $t('task.detail.actions.duplicate') }}
									</DropdownItem>
									<hr class="dropdown-divider">
									<DropdownItem
										icon="trash-alt"
										icon-class="has-text-danger"
										class="has-text-danger"
										@click="deleteTask()"
									>
										{{ $t('task.detail.actions.delete') }}
									</DropdownItem>
								</div>
							</template>
						</Dropdown>
					</div>

					<dl
						v-if="canWrite || hasAnyProperty"
						class="property-list"
					>
						<TaskPropertyRow
							name="assignees"
							icon="users"
							:label="$t('task.attributes.assignees')"
							:is-set="isSet.assignees"
							:can-write="canWrite"
						>
							<EditAssignees
								v-if="canWrite"
								:ref="e => setFieldRef('assignees', e)"
								v-model="task.assignees"
								:project-id="task.project_id"
								:task-id="task.id"
							/>
							<AssigneeList
								v-else
								:assignees="task.assignees ?? []"
							/>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="labels"
							icon="tags"
							:label="$t('task.attributes.labels')"
							:is-set="isSet.labels"
							:can-write="canWrite"
						>
							<EditLabels
								:ref="e => setFieldRef('labels', e)"
								v-model="task.labels"
								:disabled="!canWrite"
								:task-id="taskId"
								:creatable="!authStore.isLinkShareAuth"
								:creation-disabled-message="authStore.isLinkShareAuth ? $t('task.label.linkShareCannotCreate') : ''"
							/>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="priority"
							icon="exclamation-circle"
							:label="$t('task.attributes.priority')"
							:is-set="isSet.priority"
							:can-write="canWrite"
						>
							<PrioritySelect
								:ref="e => setFieldRef('priority', e)"
								v-model="task.priority"
								:disabled="!canWrite"
								@update:modelValue="setPriority"
							/>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="percent-done"
							icon="percent"
							:label="$t('task.attributes.percentDone')"
							:is-set="isSet.percentDone"
							:can-write="canWrite"
						>
							<PercentDoneSelect
								v-model="task.percent_done"
								:disabled="!canWrite"
								@update:modelValue="setPercentDone"
							/>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="due-date"
							icon="calendar"
							:label="$t('task.detail.properties.dueDate')"
							:is-set="isSet.dueDate"
							:can-write="canWrite"
							:remove-label="$t('task.detail.removeDueDate')"
							@remove="task.due_date = ''; saveTask()"
						>
							<Datepicker
								ref="dueDatePicker"
								v-model="dueDateInput"
								:choose-date-label="$t('task.detail.chooseDueDate')"
								:empty-label="$t('misc.notSet')"
								:title="$t('task.detail.properties.dueDate')"
								:disabled="taskBusy || !canWrite"
								@closeOnChange="saveTask()"
							/>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="start-date"
							icon="play"
							:label="$t('task.detail.properties.startDate')"
							:is-set="isSet.startDate"
							:can-write="canWrite"
							:remove-label="$t('task.detail.removeStartDate')"
							@remove="task.start_date = ''; saveTask()"
						>
							<Datepicker
								v-model="startDateInput"
								:choose-date-label="$t('task.detail.chooseStartDate')"
								:empty-label="$t('misc.notSet')"
								:title="$t('task.detail.properties.startDate')"
								:disabled="taskBusy || !canWrite"
								@closeOnChange="saveTask()"
							/>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="end-date"
							icon="stop"
							:label="$t('task.detail.properties.endDate')"
							:is-set="isSet.endDate"
							:can-write="canWrite"
							:remove-label="$t('task.detail.removeEndDate')"
							@remove="task.end_date = ''; saveTask()"
						>
							<Datepicker
								v-model="endDateInput"
								:choose-date-label="$t('task.detail.chooseEndDate')"
								:empty-label="$t('misc.notSet')"
								:title="$t('task.detail.properties.endDate')"
								:disabled="taskBusy || !canWrite"
								@closeOnChange="saveTask()"
							/>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="reminders"
							:icon="['far', 'clock']"
							:label="$t('task.attributes.reminders')"
							:is-set="isSet.reminders"
							:can-write="canWrite"
						>
							<Reminders
								:ref="e => setFieldRef('reminders', e)"
								v-model="task.reminders"
								:default-relative-to="remindersDefaultRelativeTo"
								:disabled="!canWrite"
								@update:modelValue="saveTask()"
							/>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="repeat"
							icon="history"
							:label="$t('task.attributes.repeat')"
							:is-set="isSet.repeatAfter"
							:can-write="canWrite"
							:remove-label="$t('task.detail.removeRepeat')"
							:class="{'is-expanded': repeatEditorOpen}"
							@remove="removeRepeatAfter"
						>
							<BaseButton
								id="task-property-repeat-value"
								aria-labelledby="task-property-repeat-label task-property-repeat-value"
								class="property-value-button"
								:class="{'property-empty': !isSet.repeatAfter}"
								:disabled="!canWrite"
								:aria-expanded="canWrite ? repeatEditorOpen : undefined"
								@click="repeatEditorOpen = !repeatEditorOpen"
							>
								{{ repeatSummary }}
								<Icon
									v-if="canWrite"
									icon="chevron-down"
									class="chevron"
								/>
							</BaseButton>
							<template #editor>
								<div
									v-if="repeatEditorOpen && canWrite"
									class="property-editor"
								>
									<RepeatAfter
										:model-value="task"
										@update:modelValue="saveTask($event)"
									/>
								</div>
							</template>
						</TaskPropertyRow>

						<TaskPropertyRow
							name="color"
							icon="fill-drip"
							:label="$t('task.attributes.color')"
							:is-set="isSet.color"
							:can-write="canWrite"
							:remove-label="$t('task.detail.removeColor')"
							@remove="taskColor = ''; saveTask()"
						>
							<ColorPicker
								v-if="canWrite"
								:ref="e => setFieldRef('color', e)"
								v-model="taskColor"
								:resettable="false"
								@update:modelValue="saveTask()"
							/>
							<ColorBubble
								v-else
								:color="color"
							/>
						</TaskPropertyRow>
					</dl>

					<!-- Created / Updated [by] -->
					<CreatedUpdated :task="task" />
				</aside>
			</div>
		</div>

		<BaseButton
			v-if="showScrollToCommentsButton"
			v-tooltip="$t('task.detail.scrollToBottom')"
			class="scroll-to-comments-button d-print-none"
			:aria-label="$t('task.detail.scrollToBottom')"
			@click="scrollToBottom"
		>
			<Icon icon="chevron-down" />
		</BaseButton>
	</div>
</template>

<script lang="ts" setup>
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {ref, reactive, computed, watch, nextTick, onBeforeUnmount, onMounted, useTemplateRef, type ComponentPublicInstance} from 'vue'
import {useRouter, useRoute, type RouteLocation, onBeforeRouteLeave} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {unrefElement, useDebounceFn, useElementSize, useIntersectionObserver, useMutationObserver} from '@vueuse/core'
import {klona} from 'klona/lite'

import {useTask} from '@/composables/useTask'
import {getHexColor} from '@/helpers/task'
import {createTaskDraft, mergeTask, parseRepeatAfter} from '@/helpers/task'

import type {Task as ITask} from '@/client/generated'
import type {ProjectResponse} from '@/client/queries/projects'

import {PRIORITIES} from '@/constants/priorities'
import {PERMISSIONS} from '@/constants/permissions'
import {PRO_FEATURE} from '@/constants/proFeatures'
import {SHORTCUTS} from '@/constants/shortcuts'

import BaseButton from '@/components/base/BaseButton.vue'

// partials
import Attachments from '@/components/tasks/partials/Attachments.vue'
import TaskTimeTracking from '@/components/time-tracking/TaskTimeTracking.vue'
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
import TaskPropertyRow from '@/components/tasks/partials/TaskPropertyRow.vue'
import TaskSubscription from '@/components/misc/Subscription.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import ColorBubble from '@/components/misc/ColorBubble.vue'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import BucketSelect from '@/components/tasks/partials/BucketSelect.vue'
import Reactions from '@/components/input/Reactions.vue'

import {getProjectTitle} from '@/helpers/getProjectTitle'
import {getDefaultProjectView, getProjectViewId} from '@/helpers/projectView'
import {scrollIntoView} from '@/helpers/scrollIntoView'
import {getScrollParent} from '@/helpers/getScrollParent'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import {playPopSound} from '@/helpers/playPop'
import {isTypingTarget, matchesKey, parseKey} from '@/helpers/shortcut'
import {taskLoadErrorAction} from './taskDetailError'
import TaskInTrashNotice from '@/components/tasks/partials/TaskInTrashNotice.vue'
import ProjectInTrashNotice from '@/components/project/ProjectInTrashNotice.vue'

import {
	useUpdateTaskMutation,
	useDeleteTaskMutation,
	useFavoriteTaskMutation,
	useDuplicateTaskMutation,
	useMarkTaskReadMutation,
} from '@/client/queries/taskMutations'
import {useProjects} from '@/composables/useProjects'
import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'

import {useTitle} from '@/composables/useTitle'
import {useTaskDetailShortcuts} from '@/composables/useTaskDetailShortcuts'
import {useDelayedLoading} from '@/composables/useDelayedLoading'

import {error, success} from '@/message'
import type {Action as MessageAction} from '@/message'
import {useSetTaskSubscriptionMutation} from '@/client/queries/subscriptions'
import {useRestoreTaskMutation} from '@/client/queries/trash'
import {useFocusAfterRemoval} from '@/composables/useFocusAfterRemoval'

const props = defineProps<{
	taskId: number,
	backdropView?: RouteLocation['fullPath'],
}>()

defineEmits<{
	'close': [],
}>()

const router = useRouter()
const route = useRoute()
const {t} = useI18n({useScope: 'global'})

const projectList = useProjects()
const updateTask = useUpdateTaskMutation()
const deleteTaskMutation = useDeleteTaskMutation()
const restoreTaskMutation = useRestoreTaskMutation()
const favoriteTask = useFavoriteTaskMutation()
const duplicateTask = useDuplicateTaskMutation()
const markTaskRead = useMarkTaskReadMutation()
const subscriptionMutation = useSetTaskSubscriptionMutation()
const taskMutating = computed(() => [
	updateTask,
	deleteTaskMutation,
	favoriteTask,
	duplicateTask,
	markTaskRead,
].some(mutation => mutation.isPending.value))
const configStore = useConfigStore()
const authStore = useAuthStore()
const timeTrackingEnabled = computed(() => configStore.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING)
	&& !authStore.isLinkShareAuth)
const baseStore = useBaseStore()

const taskQuery = useTask(
	() => props.taskId ?? 0,
	() => [
		'reactions',
		'is_unread',
		'buckets',
		...(timeTrackingEnabled.value ? ['time_entries_count' as const] : []),
	],
)
const task = ref(createTaskDraft())

// Only fields edited here stay local; the rest follows the cache.
function followServerFields(loaded: ITask) {
	const {priority, percent_done, due_date, start_date, end_date, reminders, repeat_after, repeat_mode} = task.value
	task.value = {
		...createTaskDraft(klona(loaded)),
		priority,
		percent_done,
		due_date,
		start_date,
		end_date,
		reminders,
		repeat_after,
		repeat_mode,
	}
}

function seedTask(loaded: ITask) {
	task.value = createTaskDraft(klona(loaded))
	taskColor.value = task.value.hex_color ?? ''
}

function dateInput(field: 'due_date' | 'start_date' | 'end_date') {
	return computed({
		get: () => parseDateOrNull(task.value[field]),
		set: (date: Date | string | null) => {
			task.value[field] = parseDateOrNull(date)?.toISOString() ?? ''
		},
	})
}
const dueDateInput = dateInput('due_date')
const startDateInput = dateInput('start_date')
const endDateInput = dateInput('end_date')
const hasAttachments = computed(() => (task.value.attachments?.length ?? 0) > 0)
const remindersDefaultRelativeTo = computed(() => {
	if (parseDateOrNull(task.value.due_date)) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE
	}
	if (parseDateOrNull(task.value.start_date)) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE
	}
	if (parseDateOrNull(task.value.end_date)) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE
	}
	return null
})
const taskLoadError = computed(() => taskQuery.error.value ? taskLoadErrorAction(taskQuery.error.value) : null)
const trashedTask = computed(() => taskLoadError.value?.action === 'inTrash'
	? taskLoadError.value.inTrash
	: null)
const trashedProject = computed(() => taskLoadError.value?.action === 'projectInTrash'
	? taskLoadError.value.inTrash
	: null)
const hasTrashNotice = computed(() => trashedTask.value !== null || trashedProject.value !== null)
const taskNotFound = computed(() => hasTrashNotice.value || taskLoadError.value?.action === 'notFound')
const taskTitle = computed(() => {
	if (trashedProject.value) return t('trash.title')
	return trashedTask.value ? t('trash.taskInTrashTitle') : task.value.title
})
useTitle(taskTitle)

const lastProject = computed(() => {
	const backRoute = router.options.history.state?.back
	if (!backRoute || typeof backRoute !== 'string') {
		return null
	}

	const projectMatch = backRoute.match(/\/projects\/(-?\d+)/)
	if (!projectMatch || !projectMatch[1]) {
		return null
	}

	const id = parseInt(projectMatch[1])

	return projectList.projects[id] ?? null
})

const lastProjectOrTaskProject = computed(() => lastProject.value ?? project.value)

onBeforeRouteLeave(async () => {
	if (taskNotFound.value) {
		return
	}

	if (!lastProjectOrTaskProject.value) {
		await new Promise<void>((resolve) => {
			const timeout = setTimeout(() => {
				stop()
				resolve()
			}, 5000) // 5 second timeout

			const stop = watch(lastProjectOrTaskProject, (p) => {
				if (p) {
					clearTimeout(timeout)
					stop()
					resolve()
				}
			})
		})
	}

	if (lastProjectOrTaskProject.value) {
		baseStore.setCurrentProjectIfNotSet(lastProjectOrTaskProject.value)
	}
})

// We doubled the task color property here because verte does not have a real change property, leading
// to the color property change being triggered when the # is removed from it, leading to an update,
// which leads in turn to a change... This creates an infinite loop in which the task is updated, changed,
// updated, changed, updated and so on.
// To prevent this, we put the task color property in a separate value which is set to the task color
// when it is saved and loaded.
const taskColor = ref('')

// Used to avoid flashing of empty elements if the task content is not yet loaded.
const visible = ref(false)

const project = computed(() => projectList.projects[task.value.project_id ?? 0])

const projectRoute = computed(() => ({
	name: 'project.index',
	params: {projectId: task.value.project_id},
	hash: route.hash,
}))

const canWrite = computed(() => (
	(taskQuery.task.value?.max_permission ?? 0) > PERMISSIONS.READ
))

const color = computed(() => getHexColor(task.value.hex_color))

const isModal = computed(() => Boolean(props.backdropView))

const heading = ref<HTMLElement | null>(null)

async function scrollToHeading() {
	scrollIntoView(unrefElement(heading))
}

const attachmentsRef = ref<InstanceType<typeof Attachments> | null>(null)

const taskViewContainer = ref<HTMLElement | null>(null)
useFocusAfterRemoval(() => hasTrashNotice.value, () => taskViewContainer.value)
const scrollContainer = ref<HTMLElement | null>(null)
const contentBottomMarker = ref<HTMLElement | null>(null)
const bottomMarkerVisible = ref(true)
const isScrollable = ref(false)

function resolveScrollContainer() {
	// Null after unmount; deferred callers must not fall back to `document` then.
	if (!taskViewContainer.value) {
		return
	}

	scrollContainer.value = getScrollParent(taskViewContainer.value)
}

function updateScrollable() {
	const scroller = scrollContainer.value
	if (!scroller) {
		isScrollable.value = false
		return
	}

	isScrollable.value = scroller.scrollHeight > scroller.clientHeight + 1
}

const showScrollToCommentsButton = computed(() => {
	return isScrollable.value && !bottomMarkerVisible.value
})

function scrollToBottom() {
	if (!contentBottomMarker.value) {
		return
	}

	contentBottomMarker.value.scrollIntoView({
		behavior: 'smooth',
		block: 'end',
		inline: 'nearest',
	})
}

useIntersectionObserver(
	contentBottomMarker,
	([entry]) => {
		bottomMarkerVisible.value = entry?.isIntersecting ?? true
	},
	{threshold: 0.1},
)

const debouncedMutationHandler = useDebounceFn(async () => {
	await nextTick()
	resolveScrollContainer()
	updateScrollable()
}, 100)

useMutationObserver(
	taskViewContainer,
	debouncedMutationHandler,
	{subtree: true, childList: true},
)
onBeforeUnmount(() => debouncedMutationHandler.cancel())

const {height: scrollContainerHeight} = useElementSize(scrollContainer)
watch(scrollContainerHeight, () => updateScrollable())

onMounted(async () => {
	await nextTick()
	resolveScrollContainer()
	updateScrollable()
})

const taskBusy = useDelayedLoading(() => taskQuery.isFetching.value || taskMutating.value)


type FieldType =
	| 'assignees'
	| 'attachments'
	| 'color'
	| 'dueDate'
	| 'labels'
	| 'moveProject'
	| 'priority'
	| 'relatedTasks'
	| 'reminders'
	| 'timeTracking'

type SectionType = 'attachments' | 'moveProject' | 'relatedTasks' | 'timeTracking'

const activeSections: Record<SectionType, boolean> = reactive({
	attachments: false,
	moveProject: false,
	relatedTasks: false,
	timeTracking: false,
})

// Only reveals sections: a save must not hide one the user just opened.
function revealSectionsWithData() {
	activeSections.attachments ||= hasAttachments.value
	activeSections.timeTracking ||= (task.value.time_entries_count ?? 0) > 0
	activeSections.relatedTasks ||= Object.keys(task.value.related_tasks ?? {}).length > 0
}

function resetSections() {
	for (const section of Object.keys(activeSections) as SectionType[]) {
		activeSections[section] = false
	}
	repeatEditorOpen.value = false
	revealSectionsWithData()
}

const isSet = computed(() => ({
	assignees: (task.value.assignees?.length ?? 0) > 0,
	labels: (task.value.labels?.length ?? 0) > 0,
	priority: task.value.priority !== PRIORITIES.UNSET,
	percentDone: (task.value.percent_done ?? 0) > 0,
	dueDate: Boolean(parseDateOrNull(task.value.due_date)),
	startDate: Boolean(parseDateOrNull(task.value.start_date)),
	endDate: Boolean(parseDateOrNull(task.value.end_date)),
	reminders: (task.value.reminders?.length ?? 0) > 0,
	repeatAfter: (task.value.repeat_after ?? 0) > 0
		|| task.value.repeat_mode !== TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT,
	color: Boolean(color.value),
}))
const hasAnyProperty = computed(() => Object.values(isSet.value).some(Boolean))

const favoriteLabel = computed(() => task.value.is_favorite
	? t('task.detail.actions.unfavorite')
	: t('task.detail.actions.favorite'))

const repeatEditorOpen = ref(false)
const repeatSummary = computed(() => {
	if (!isSet.value.repeatAfter) {
		return t('misc.notSet')
	}
	if (task.value.repeat_mode === TASK_REPEAT_MODES.REPEAT_MODE_MONTH) {
		return t('task.repeat.monthly')
	}
	const fromCurrentDate = task.value.repeat_mode === TASK_REPEAT_MODES.REPEAT_MODE_FROM_CURRENT_DATE
	const {amount, type} = parseRepeatAfter(task.value.repeat_after ?? 0)
	if (amount === 0) {
		return t('task.repeat.fromCurrentDate')
	}
	const interval = t(`task.repeat.summary.${type}`, {count: amount}, amount)
	return fromCurrentDate
		? t('task.repeat.summary.fromCurrentDate', {interval})
		: interval
})

watch(taskQuery.task, async (loaded, previous) => {
	if (!loaded) return
	if (loaded.id !== previous?.id) {
		seedTask(loaded)
		resetSections()
	} else {
		followServerFields(loaded)
		revealSectionsWithData()
	}
	if (loaded.is_unread) markTaskRead.mutateAsync(loaded.id).catch(() => {})
	if (lastProject.value) baseStore.setCurrentProjectIfNotSet(lastProject.value)
	await nextTick()
	if (loaded.id !== previous?.id) scrollToHeading()
	resolveScrollContainer()
	updateScrollable()
	visible.value = true
}, {immediate: true})
watch(taskLoadError, loadError => {
	switch (loadError?.action) {
		case 'notFound':
			router.replace({name: 'not-found'})
			break
		case 'report':
			error(taskQuery.error.value)
			visible.value = true
			break
	}
})

const activeFieldElements: { [id in FieldType]: HTMLElement | null } = reactive({
	assignees: null,
	attachments: null,
	color: null,
	dueDate: null,
	labels: null,
	moveProject: null,
	priority: null,
	relatedTasks: null,
	reminders: null,
	timeTracking: null,
})

const dueDatePicker = useTemplateRef<InstanceType<typeof Datepicker>>('dueDatePicker')

function setFieldRef(name: FieldType, e: Element | ComponentPublicInstance | null) {
	const element = e instanceof Element ? e : e?.$el
	activeFieldElements[name] = element instanceof HTMLElement ? element : null
}

const FOCUSABLE = 'input, select, textarea, button, [tabindex]:not([tabindex="-1"])'

function setFieldActive(fieldName: FieldType) {
	if (fieldName in activeSections) {
		activeSections[fieldName as SectionType] = true
	}
	nextTick(() => {
		const datepicker = fieldName === 'dueDate' ? dueDatePicker.value : null
		const el: HTMLElement | null = datepicker?.$el ?? activeFieldElements[fieldName]

		if (!el) {
			return
		}

		// Multiselect's tabindex=-1 root forwards focus to its search input.
		const focusable = el.matches(`[tabindex], ${FOCUSABLE}`) ? el : el.querySelector<HTMLElement>(FOCUSABLE)
		focusable?.focus()

		// Finish scrolling before the datepicker sheet locks the page.
		scrollIntoView(el, datepicker ? 'instant' : 'smooth')

		datepicker?.open()
	})
}

// v-shortcut only clicks; these need focus, or live in the closed actions menu.
const fieldShortcuts = ([
	[SHORTCUTS.taskDetail.labels, () => setFieldActive('labels')],
	[SHORTCUTS.taskDetail.priority, () => setFieldActive('priority')],
	[SHORTCUTS.taskDetail.color, () => setFieldActive('color')],
	[SHORTCUTS.taskDetail.assignees, () => setFieldActive('assignees')],
	[SHORTCUTS.taskDetail.dueDate, () => setFieldActive('dueDate')],
	[SHORTCUTS.taskDetail.reminder, () => setFieldActive('reminders')],
	[SHORTCUTS.taskDetail.attachments, () => openAttachments()],
	[SHORTCUTS.taskDetail.relatedTasks, () => setRelatedTasksActive()],
	[SHORTCUTS.taskDetail.moveProject, () => setFieldActive('moveProject')],
	[SHORTCUTS.taskDetail.delete, () => deleteTask()],
] as const).map(([key, run]) => [parseKey(key), run] as const)

function handleFieldShortcut(event: KeyboardEvent) {
	if (!canWrite.value || event.defaultPrevented || event.repeat || event.isComposing || isTypingTarget(event)) {
		return
	}
	for (const [parsed, run] of fieldShortcuts) {
		if (matchesKey(event, parsed)) {
			event.preventDefault()
			run()
			return
		}
	}
}

onMounted(() => document.addEventListener('keydown', handleFieldShortcut))
onBeforeUnmount(() => document.removeEventListener('keydown', handleFieldShortcut))

function openAttachments() {
	activeSections.attachments = true
	nextTick(() => {
		const el = activeFieldElements.attachments
		if (el) {
			scrollIntoView(el)
		}
		attachmentsRef.value?.openFilePicker()
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

	currentTask.hex_color = taskColor.value

	// If no end date is being set, but a start date and due date,
	// use the due date as the end date
	if (
		!parseDateOrNull(currentTask.end_date) &&
		Boolean(parseDateOrNull(currentTask.start_date)) &&
		Boolean(parseDateOrNull(currentTask.due_date))
	) {
		currentTask.end_date = currentTask.due_date
	}

	seedTask(mergeTask(task.value, await updateTask.mutateAsync({...currentTask, id: currentTask.id!})))
	revealSectionsWithData()

	let actions: MessageAction[] = []
	if (undoCallback) {
		actions = [{
			title: t('task.undo'),
			callback: undoCallback,
		}]
	}
	success({message: t('task.detail.updateSuccess')}, actions)
}

useTaskDetailShortcuts({
	task: () => task.value,
	taskTitle: () => taskTitle.value,
	onSave: saveTask,
	onDoneAndBack: markDoneAndGoBack,
})

async function deleteTask() {
	const id = task.value.id!
	const projectId = task.value.project_id
	let viewId: number | undefined
	try {
		const destinationProject = await projectList.loadProject(projectId)
		const view = destinationProject.views.find(view => view.id === getProjectViewId(projectId))
			?? getDefaultProjectView(destinationProject.views, authStore.settings.frontend_settings.default_view)
		viewId = view?.id
		await deleteTaskMutation.mutateAsync(id)
	} catch {
		return
	}
	await router.push(viewId
		? {
			name: 'project.view',
			params: {
				projectId,
				viewId,
			},
		}
		: {
			name: 'project.index',
			params: {projectId},
		})
	success({message: t('task.detail.deleteSuccess')}, [{
		title: t('task.undo'),
		callback: () => restoreTaskMutation.mutate(id),
	}], {takeFocus: true})
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

async function markDoneAndGoBack() {
	if (!canWrite.value || task.value.done) {
		return
	}

	// Keydown handlers bypass Vue's errorHandler, so surface save errors here.
	try {
		await toggleTaskDone()
	} catch (e) {
		error(e)
		return
	}
	goBack()
}

function goBack() {
	if (lastProject.value) {
		router.back()
		return
	}
	router.push(projectRoute.value)
}

async function changeProject(project: ProjectResponse | null) {
	if (project === null) {
		return
	}
	await saveTask({
		...task.value,
		project_id: project.id,
	})
	baseStore.setCurrentProject(project)
}

function toggleSubscription(subscribed: boolean) {
	if (subscriptionMutation.isPending.value) return
	subscriptionMutation.mutate({taskId: task.value.id!, subscribed})
}

async function toggleFavorite() {
	await favoriteTask.mutateAsync({...task.value, id: task.value.id!})
}

async function duplicateCurrentTask() {
	const duplicatedTask = await duplicateTask.mutateAsync(task.value.id!)
	if (duplicatedTask) {
		success({message: t('task.detail.duplicateSuccess')})
		router.push({
			name: 'task.detail',
			params: {id: duplicatedTask.id},
		})
	}
}

async function setPriority(priority: number) {
	const newTask: ITask = {
		...task.value,
		priority,
	}

	return saveTask(newTask)
}

async function setPercentDone(percentDone: number) {
	const newTask: ITask = {
		...task.value,
		percent_done: percentDone,
	}

	return saveTask(newTask)
}

async function removeRepeatAfter() {
	task.value.repeat_after = 0
	task.value.repeat_mode = TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT
	await saveTask()
}

function setRelatedTasksActive() {
	setFieldActive('relatedTasks')

	// If the related tasks are already available, show the form again
	const el = activeFieldElements['relatedTasks']
	if (!el) {
		return
	}
	for (const child of Array.from(el.children)) {
		if ((child as HTMLElement).id === 'showRelatedTasksFormButton') {
			(child as HTMLElement).click()
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

h2 .button {
	vertical-align: middle;
}

.icon.is-grey {
	color: var(--grey-400);
}

.details {
	padding-block-end: 0.75rem;
	margin-block-end: 0;
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

// keep the title clear of the modal's fixed close button
.is-modal .heading {
	@media screen and (min-width: $tablet) and (max-width: $desktop) {
		padding-inline-end: 3.5rem;
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

// Bulma stacks columns as blocks on narrow screens, so flex is needed for `order`.
.task-columns {
	@media screen and (max-width: $tablet), print {
		display: flex;
		flex-direction: column;
	}
}

// Not sticky: the list can outgrow the viewport and its popups open downwards.
.task-sidebar {
	container: task-sidebar / inline-size;

	@media screen and (max-width: $tablet), print {
		order: -1;
	}

	@media print {
		inline-size: 100% !important;

		.property:not(.is-set),
		.chevron,
		:deep(.multiselect input),
		:deep(.remove-assignee),
		:deep(.tag .delete),
		:deep(.reminders .remove),
		:deep(.reminders > :not(.reminder-input)) {
			display: none;
		}
	}
}

.sidebar-toolbar {
	display: flex;
	flex-wrap: wrap;
	align-items: stretch;
	gap: .375rem;
	margin-block-end: .75rem;

	.button--mark-done {
		flex: 1 1 10rem;
		line-height: 1.25;
		padding-block: .25rem;
		text-transform: none;
		justify-content: flex-start;
		text-align: start;
		background-color: transparent;
		box-shadow: none;

		// bright brand green with fixed dark text passes contrast in both themes
		&.is-pending {
			background-color: var(--success);
			color: hsl(215, 27.9%, 16.9%);

			&:hover,
			&:focus {
				filter: brightness(1.05);
			}
		}
	}
}

.toolbar-icon-button {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	inline-size: 2.5rem;
	min-block-size: $button-height;
	border-radius: $radius;
	border: 1px solid var(--grey-300);
	color: var(--grey-600);
	background: var(--scheme-main);
	transition: color $transition, background-color $transition, border-color $transition, transform 100ms ease;

	&:hover,
	&:focus-visible {
		color: var(--text);
		border-color: var(--grey-400);
	}

	&:active {
		transform: scale(.95);
	}

	&.is-favorite {
		color: var(--warning);
		border-color: color-mix(in srgb, var(--warning) 40%, transparent);
		background: color-mix(in srgb, var(--warning) 10%, var(--scheme-main));
	}
}

.property-list {
	margin: 0;
	border-block-start: 1px solid var(--grey-200);
	font-size: .9375rem;
}

.property-empty {
	color: var(--grey-600);
	font-style: italic;
}

.property-value-button {
	display: inline-flex;
	align-items: center;
	gap: .375rem;
	flex: 1;
	min-block-size: 2rem;
	padding: .25rem .5rem;
	border-radius: $radius;
	text-align: start;
	color: var(--text);
	transition: background-color $transition;

	.chevron {
		inline-size: .625rem;
		color: var(--grey-400);
		transition: transform 200ms ease;
	}

	&:not(:disabled):hover {
		background: var(--scheme-main);
	}

	.is-expanded & .chevron {
		transform: rotate(180deg);
	}
}

.property-editor {
	grid-column: 1 / -1;
	padding: .25rem 0 .25rem .5rem;
	border-inline-start: 2px solid var(--grey-200);
	margin-block-start: .25rem;

	// Bulma addons don't shrink, which pushes the unit select out of the narrow sidebar.
	:deep(.field.has-addons) {
		min-inline-size: 0;

		.control:first-child {
			flex: 1 1 auto;
			min-inline-size: 0;
		}
	}
}

:deep(.created) {
	margin-block-start: 1rem;
	font-size: .8125rem;
	color: var(--grey-500);
	text-align: start;
}

.scroll-to-comments-button {
	position: fixed;
	// Position above the keyboard shortcuts button (which is at bottom: calc(1rem - 4px))
	inset-block-end: calc(var(--bottom-banners-end-height, 0px) + 2.5rem);
	inset-inline-end: .75rem;
	z-index: 10;
	inline-size: 2rem;
	block-size: 2rem;
	border-radius: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0;
	background-color: var(--site-background);
	border: 1px solid var(--grey-300);
	color: var(--grey-500);
	box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
	transition: all $transition;

	&:hover {
		background-color: var(--grey-100);
		color: var(--grey-700);
	}

	@media screen and (max-width: $tablet) {
		// Hide on mobile since keyboard shortcuts button is also hidden
		display: none;
	}
}
</style>

<style lang="scss">
// global style to override position when the modal task detail is active
.modal-content .scroll-to-comments-button {
	inset-block-end: .75rem;
	inset-inline-end: 1rem;
}

// the task card spans the full width here, so the modal's white close button sits on it instead of the scrim
@media screen and (min-width: $tablet) and (max-width: $desktop) {
	.modal-dialog:has(.task-view-container.is-modal) .modal-container > .close {
		color: var(--text);
	}
}
</style>
