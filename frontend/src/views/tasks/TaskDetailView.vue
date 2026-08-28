<template>
	<div
		ref="taskViewContainer"
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
			<BaseButton
				v-if="!isModal"
				class="back-button mbs-2"
				@click="lastProject ? router.back() : router.push(projectRoute)"
			>
				<Icon icon="arrow-left" />
				{{ $t('task.detail.back') }}
			</BaseButton>
			<Heading
				ref="heading"
				:task="task"
				:can-write="canWrite"
				:has-close="isModal"
				@update:task="Object.assign(task, $event)"
				@close="$emit('close')"
			/>
			<nav
				v-if="project?.id"
				aria-label="Breadcrumb"
				class="subtitle"
			>
				<template
					v-for="p in projectStore.getAncestors(project)"
					:key="p.id"
				>
					<a
						v-if="router.options.history.state?.back?.includes('/projects/'+p.id+'/') || false"
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
					@update:task="Object.assign(task, $event)"
				/>
			</nav>

			<ChecklistSummary :task="task" />

			<div class="columns task-columns mbs-2">
				<!-- Content -->
				<div class="column is-two-thirds detail-content">
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
						class="details d-print-none"
						:disabled="!canWrite"
					/>

					<!-- Attachments -->
					<div
						v-show="activeFields.attachments || hasAttachments"
						class="content attachments"
					>
						<Attachments
							:ref="e => { setFieldRef('attachments', e); attachmentsRef = e as any }"
							:edit-enabled="canWrite"
							:task="task"
							@taskChanged="({coverImageAttachmentId}) => task.coverImageAttachmentId = coverImageAttachmentId"
							@update:attachments="onAttachmentsUpdated"
						/>
					</div>

					<!-- Time Tracking -->
					<div
						v-if="timeTrackingEnabled && activeFields.timeTracking"
						:ref="e => setFieldRef('timeTracking', e)"
						class="content time-tracking"
					>
						<TaskTimeTracking :task-id="task.id" />
					</div>

					<!-- Related Tasks -->
					<div
						v-if="activeFields.relatedTasks"
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
						:project-id="task.projectId"
						:initial-comments="task.comments"
					/>

					<!-- Marker element for scroll-to-bottom button visibility -->
					<div
						ref="contentBottomMarker"
						class="content-bottom-marker"
					/>
				</div>

				<!-- Sidebar: toolbar + always-visible properties -->
				<aside class="column is-one-third action-buttons task-sidebar d-print-none">
					<div
						v-if="canWrite"
						class="sidebar-toolbar"
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
							v-tooltip="task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')"
							class="toolbar-icon-button"
							:class="{'is-favorite': task.isFavorite}"
							:aria-label="task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')"
							:aria-pressed="task.isFavorite"
							@click="toggleFavorite"
						>
							<Icon :icon="task.isFavorite ? 'star' : ['far', 'star']" />
						</BaseButton>
						<Dropdown
							class="task-actions-menu"
							:trigger-label="$t('task.detail.moreActions')"
						>
							<template #trigger="{toggleOpen, open}">
								<BaseButton
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
										:model-value="task.subscription"
										@update:modelValue="sub => task.subscription = sub"
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
										@click="showDeleteModal = true"
									>
										{{ $t('task.detail.actions.delete') }}
									</DropdownItem>
								</div>
							</template>
						</Dropdown>
					</div>

					<div
						class="sidebar-layout-switch"
						role="radiogroup"
						:aria-label="$t('task.detail.sidebarLayout.label')"
					>
						<BaseButton
							v-for="layout in SIDEBAR_LAYOUTS"
							:key="layout"
							role="radio"
							:aria-checked="sidebarLayout === layout"
							:class="{'is-active': sidebarLayout === layout}"
							@click="sidebarLayout = layout"
						>
							<Icon :icon="layout === 'rows' ? 'list' : 'table-cells-large'" />
							{{ $t(`task.detail.sidebarLayout.${layout}`) }}
						</BaseButton>
					</div>

					<dl
						class="property-list"
						:class="`is-${sidebarLayout}`"
					>
						<h3 class="property-group property-group-schedule">
							{{ $t('task.detail.sidebarLayout.groupSchedule') }}
						</h3>
						<h3 class="property-group property-group-details">
							{{ $t('task.detail.sidebarLayout.groupDetails') }}
						</h3>
						<div
							class="property property-assignees"
							:class="{'is-set': task.assignees.length > 0}"
						>
							<dt>
								<Icon icon="users" />
								{{ $t('task.attributes.assignees') }}
							</dt>
							<dd>
								<EditAssignees
									v-if="canWrite"
									:ref="e => setFieldRef('assignees', e)"
									v-model="task.assignees"
									:project-id="task.projectId"
									:task-id="task.id"
								/>
								<AssigneeList
									v-else-if="task.assignees.length > 0"
									:assignees="task.assignees"
								/>
								<span
									v-else
									class="property-empty"
								>{{ $t('misc.notSet') }}</span>
							</dd>
						</div>

						<div
							class="property property-labels"
							:class="{'is-set': task.labels.length > 0}"
						>
							<dt>
								<Icon icon="tags" />
								{{ $t('task.attributes.labels') }}
							</dt>
							<dd>
								<EditLabels
									v-if="canWrite || task.labels.length > 0"
									:ref="e => setFieldRef('labels', e)"
									v-model="task.labels"
									:disabled="!canWrite"
									:task-id="taskId"
									:creatable="!authStore.isLinkShareAuth"
									:creation-disabled-message="authStore.isLinkShareAuth ? $t('task.label.linkShareCannotCreate') : ''"
								/>
								<span
									v-else
									class="property-empty"
								>{{ $t('misc.notSet') }}</span>
							</dd>
						</div>

						<div
							class="property property-priority"
							:class="{'is-set': task.priority !== PRIORITIES.UNSET}"
						>
							<dt>
								<Icon icon="exclamation-circle" />
								{{ $t('task.attributes.priority') }}
							</dt>
							<dd>
								<PrioritySelect
									:ref="e => setFieldRef('priority', e)"
									v-model="task.priority"
									:disabled="!canWrite"
									@update:modelValue="setPriority"
								/>
							</dd>
						</div>

						<div
							class="property property-percent-done"
							:class="{'is-set': task.percentDone > 0}"
						>
							<dt>
								<Icon icon="percent" />
								{{ $t('task.attributes.percentDone') }}
							</dt>
							<dd>
								<PercentDoneSelect
									:ref="e => setFieldRef('percentDone', e)"
									v-model="task.percentDone"
									:disabled="!canWrite"
									@update:modelValue="setPercentDone"
								/>
							</dd>
						</div>

						<div
							class="property property-due-date"
							:class="{'is-set': task.dueDate !== null}"
						>
							<dt>
								<Icon icon="calendar" />
								{{ $t('task.attributes.dueDate') }}
							</dt>
							<dd class="date-input">
								<Datepicker
									:ref="e => setFieldRef('dueDate', e)"
									v-model="task.dueDate"
									:choose-date-label="$t('task.detail.chooseDueDate')"
									:empty-label="$t('misc.notSet')"
									:disabled="taskService.loading || !canWrite"
									@closeOnChange="saveTask()"
								/>
								<BaseButton
									v-if="task.dueDate && canWrite"
									class="remove"
									:aria-label="$t('task.detail.removeDueDate')"
									@click="() => {task.dueDate = null;saveTask()}"
								>
									<Icon icon="times" />
								</BaseButton>
							</dd>
						</div>

						<div
							class="property property-start-date"
							:class="{'is-set': task.startDate !== null}"
						>
							<dt>
								<Icon icon="play" />
								{{ $t('task.attributes.startDate') }}
							</dt>
							<dd class="date-input">
								<Datepicker
									:ref="e => setFieldRef('startDate', e)"
									v-model="task.startDate"
									:choose-date-label="$t('task.detail.chooseStartDate')"
									:empty-label="$t('misc.notSet')"
									:disabled="taskService.loading || !canWrite"
									@closeOnChange="saveTask()"
								/>
								<BaseButton
									v-if="task.startDate && canWrite"
									class="remove"
									:aria-label="$t('task.detail.removeStartDate')"
									@click="() => {task.startDate = null;saveTask()}"
								>
									<Icon icon="times" />
								</BaseButton>
							</dd>
						</div>

						<div
							class="property property-end-date"
							:class="{'is-set': task.endDate !== null}"
						>
							<dt>
								<Icon icon="stop" />
								{{ $t('task.attributes.endDate') }}
							</dt>
							<dd class="date-input">
								<Datepicker
									:ref="e => setFieldRef('endDate', e)"
									v-model="task.endDate"
									:choose-date-label="$t('task.detail.chooseEndDate')"
									:empty-label="$t('misc.notSet')"
									:disabled="taskService.loading || !canWrite"
									@closeOnChange="saveTask()"
								/>
								<BaseButton
									v-if="task.endDate && canWrite"
									class="remove"
									:aria-label="$t('task.detail.removeEndDate')"
									@click="() => {task.endDate = null;saveTask()}"
								>
									<Icon icon="times" />
								</BaseButton>
							</dd>
						</div>

						<div
							class="property property-reminders"
							:class="{'is-set': task.reminders.length > 0}"
						>
							<dt>
								<Icon :icon="['far', 'clock']" />
								{{ $t('task.attributes.reminders') }}
							</dt>
							<dd>
								<Reminders
									:ref="e => setFieldRef('reminders', e)"
									v-model="task.reminders"
									:default-relative-to="remindersDefaultRelativeTo"
									:disabled="!canWrite"
									@update:modelValue="saveTask()"
								/>
							</dd>
						</div>

						<div
							class="property property-repeat"
							:class="{'is-set': hasRepeat, 'is-expanded': repeatEditorOpen}"
						>
							<dt>
								<Icon icon="history" />
								{{ $t('task.attributes.repeat') }}
							</dt>
							<dd>
								<BaseButton
									:ref="e => setFieldRef('repeatAfter', e)"
									class="property-value-button"
									:class="{'property-empty': !hasRepeat}"
									:disabled="!canWrite"
									:aria-expanded="repeatEditorOpen"
									@click="repeatEditorOpen = !repeatEditorOpen"
								>
									{{ repeatSummary }}
									<Icon
										v-if="canWrite"
										icon="chevron-down"
										class="chevron"
									/>
								</BaseButton>
								<BaseButton
									v-if="hasRepeat && canWrite"
									class="remove"
									:aria-label="$t('task.detail.removeRepeat')"
									@click="removeRepeatAfter"
								>
									<Icon icon="times" />
								</BaseButton>
							</dd>
							<div
								v-if="repeatEditorOpen && canWrite"
								class="property-editor"
							>
								<RepeatAfter
									v-model="task"
									:disabled="!canWrite"
									@update:modelValue="saveTask()"
								/>
							</div>
						</div>

						<div
							class="property property-color"
							:class="{'is-set': taskColor !== ''}"
						>
							<dt>
								<Icon icon="fill-drip" />
								{{ $t('task.attributes.color') }}
							</dt>
							<dd>
								<ColorPicker
									v-if="canWrite"
									:ref="e => setFieldRef('color', e)"
									v-model="taskColor"
									menu-position="bottom"
									@update:modelValue="saveTask()"
								/>
								<ColorBubble
									v-else-if="color"
									:color="color"
								/>
								<span
									v-else
									class="property-empty"
								>{{ $t('misc.notSet') }}</span>
							</dd>
						</div>
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

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteTask()"
		>
			<template #header>
				<span>{{ $t('task.detail.delete.header') }}</span>
			</template>

			<template #text>
				<p class="tw:text-balance">
					{{ $t('task.detail.delete.text1') }}
				</p>
				<p class="tw:text-balance">
					{{ $t('task.detail.delete.text2') }}
				</p>
			</template>
		</Modal>
	</div>
</template>

<script lang="ts" setup>
import {ref, reactive, shallowReactive, computed, watch, nextTick, onMounted, onBeforeUnmount} from 'vue'
import {useRouter, useRoute, type RouteLocation, onBeforeRouteLeave} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {unrefElement, useDebounceFn, useElementSize, useIntersectionObserver, useMutationObserver, useStorage} from '@vueuse/core'
import {klona} from 'klona/lite'

import TaskService from '@/services/task'
import TaskModel from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {IProject} from '@/modelTypes/IProject'

import {PRIORITIES, type Priority} from '@/constants/priorities'
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
import TaskSubscription from '@/components/misc/Subscription.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import ColorBubble from '@/components/misc/ColorBubble.vue'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import BucketSelect from '@/components/tasks/partials/BucketSelect.vue'
import Reactions from '@/components/input/Reactions.vue'

import {uploadFile} from '@/helpers/attachments'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {scrollIntoView} from '@/helpers/scrollIntoView'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import {playPopSound} from '@/helpers/playPop'
import {isFormField, matchesKey, parseKey} from '@/helpers/shortcut'

import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'

import {useTitle} from '@/composables/useTitle'
import {useTaskDetailShortcuts} from '@/composables/useTaskDetailShortcuts'

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
const route = useRoute()
const {t} = useI18n({useScope: 'global'})

const projectStore = useProjectStore()
const taskStore = useTaskStore()
const configStore = useConfigStore()
const timeTrackingEnabled = computed(() => configStore.isProFeatureEnabled(PRO_FEATURE.TIME_TRACKING))
const kanbanStore = useKanbanStore()
const authStore = useAuthStore()
const baseStore = useBaseStore()

const task = ref<ITask>(new TaskModel())
const hasAttachments = computed(() => (task.value.attachments?.length ?? 0) > 0)
const remindersDefaultRelativeTo = computed(() => {
	if (task.value.dueDate) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE
	}
	if (task.value.startDate) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE
	}
	if (task.value.endDate) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE
	}
	return null
})
const taskNotFound = ref(false)
const taskTitle = computed(() => task.value.title)
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

	return projectStore.projects[id] ?? null
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
		await baseStore.handleSetCurrentProjectIfNotSet(lastProjectOrTaskProject.value)
	}
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

const projectRoute = computed(() => ({
	name: 'project.index',
	params: {projectId: task.value.projectId},
	hash: route.hash,
}))

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

async function attachmentUpload(file: File, onSuccess?: (url: string) => void) {
	const uploaded = await uploadFile(props.taskId, file, onSuccess)
	if (uploaded.length > 0) {
		onAttachmentsUpdated([...task.value.attachments, ...uploaded])
	}
	return uploaded
}

function onAttachmentsUpdated(attachments: IAttachment[]) {
	task.value.attachments = attachments
	kanbanStore.setTaskInBucket({
		...task.value,
		attachments,
	})
}

const heading = ref<HTMLElement | null>(null)

async function scrollToHeading() {
	scrollIntoView(unrefElement(heading))
}

const attachmentsRef = ref<InstanceType<typeof Attachments> | null>(null)

const taskViewContainer = ref<HTMLElement | null>(null)
const scrollContainer = ref<HTMLElement | null>(null)
const contentBottomMarker = ref<HTMLElement | null>(null)
const bottomMarkerVisible = ref(true)
const isScrollable = ref(false)

function resolveScrollContainer() {
	let el = taskViewContainer.value

	while (el) {
		const overflowY = getComputedStyle(el).overflowY
		if (['auto', 'scroll', 'overlay'].includes(overflowY)) {
			scrollContainer.value = el
			return
		}
		el = el.parentElement
	}

	scrollContainer.value = (document.scrollingElement as HTMLElement | null) ?? document.documentElement
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

const {height: scrollContainerHeight} = useElementSize(scrollContainer)
watch(scrollContainerHeight, () => updateScrollable())

onMounted(async () => {
	await nextTick()
	resolveScrollContainer()
	updateScrollable()
})

const taskService = shallowReactive(new TaskService())

// load task
watch(
	() => props.taskId,
	async (id) => {
		if (id === undefined) {
			return
		}

		try {
			const expand = ['reactions', 'comments', 'is_unread', 'buckets']
			if (timeTrackingEnabled.value) {
				// Only request the (server-computed) count when the feature is on.
				expand.push('time_entries_count')
			}
			const loaded = await taskService.get({id}, {expand})
			Object.assign(task.value, loaded)
			taskColor.value = task.value.hexColor
			setActiveFields()

			if (task.value.isUnread) {
				await taskStore.markTaskAsRead(task.value.id)
				task.value.isUnread = false
			}

			if (lastProject.value) {
				await baseStore.handleSetCurrentProjectIfNotSet(lastProject.value)
			}
		} catch (e) {
			// 403 means the task exists but is not visible to us; treat it like
			// a 404 so we route away instead of rendering an empty task shell.
			if (e?.response?.status === 404 || e?.response?.status === 403) {
				taskNotFound.value = true
				router.replace({name: 'not-found'})
				return
			}

			throw e
		} finally {
			await nextTick()
			scrollToHeading()
			resolveScrollContainer()
			updateScrollable()
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
	| 'timeTracking'

// Sections in the content column that are hidden until they have data or the user opens them.
type SectionType = 'attachments' | 'moveProject' | 'relatedTasks' | 'timeTracking'

const activeFields: { [type in SectionType]: boolean } = reactive({
	attachments: false,
	moveProject: false,
	relatedTasks: false,
	timeTracking: false,
})

function setActiveFields() {
	activeFields.attachments = task.value.attachments.length > 0
	activeFields.timeTracking = (task.value.timeEntriesCount ?? 0) > 0
	activeFields.relatedTasks = Object.keys(task.value.relatedTasks).length > 0
}

const SIDEBAR_LAYOUTS = ['rows', 'cards'] as const
const sidebarLayout = useStorage<typeof SIDEBAR_LAYOUTS[number]>('taskDetailSidebarLayout', 'rows')

const hasRepeat = computed(() =>
	(task.value.repeatAfter?.amount ?? 0) > 0 ||
	task.value.repeatMode !== TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT,
)
const repeatEditorOpen = ref(false)
const repeatSummary = computed(() => {
	if (!hasRepeat.value) {
		return t('misc.notSet')
	}
	if (task.value.repeatMode === TASK_REPEAT_MODES.REPEAT_MODE_MONTH) {
		return t('task.repeat.monthly')
	}
	const {amount, type} = typeof task.value.repeatAfter === 'number'
		? {amount: task.value.repeatAfter, type: 'seconds'}
		: task.value.repeatAfter
	const every = amount === 1 && type === 'days'
		? t('task.repeat.everyDay')
		: amount === 1 && type === 'weeks'
			? t('task.repeat.everyWeek')
			: t('task.repeat.every', {amount, unit: t(`task.repeat.${type}`).toLowerCase()})
	return task.value.repeatMode === TASK_REPEAT_MODES.REPEAT_MODE_FROM_CURRENT_DATE
		? `${every} · ${t('task.repeat.fromCurrentDate')}`
		: every
})

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
	timeTracking: null,
})

function setFieldRef(name, e) {
	activeFieldElements[name] = unrefElement(e)
}

const FOCUSABLE = 'input, select, textarea, button, [tabindex]:not([tabindex="-1"])'
const DATE_FIELDS: FieldType[] = ['dueDate', 'startDate', 'endDate']

function setFieldActive(fieldName: FieldType) {
	if (fieldName in activeFields) {
		activeFields[fieldName as SectionType] = true
	}
	if (fieldName === 'repeatAfter') {
		repeatEditorOpen.value = true
	}
	nextTick(() => {
		const el = activeFieldElements[fieldName]

		if (!el) {
			return
		}

		const focusable = el.matches(FOCUSABLE) ? el : el.querySelector<HTMLElement>(FOCUSABLE)
		focusable?.focus()
		// Date fields are buttons that open a popup, so a shortcut should open it right away.
		if (DATE_FIELDS.includes(fieldName)) {
			focusable?.click()
		}

		scrollIntoView(el)
	})
}

// Property rows and menu items have no always-mounted button to hang v-shortcut on,
// so their shortcuts are dispatched here.
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
	[SHORTCUTS.taskDetail.delete, () => showDeleteModal.value = true],
] as Array<[string, () => void]>).map(([key, fn]) => [parseKey(key), fn] as const)

function handleFieldShortcut(event: KeyboardEvent) {
	if (!canWrite.value || event.defaultPrevented || event.repeat || isFormField(event.target)) {
		return
	}
	for (const [parsed, fn] of fieldShortcuts) {
		if (matchesKey(event, parsed)) {
			event.preventDefault()
			fn()
			return
		}
	}
}

onMounted(() => document.addEventListener('keydown', handleFieldShortcut))
onBeforeUnmount(() => document.removeEventListener('keydown', handleFieldShortcut))

function openAttachments() {
	activeFields.attachments = true
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

useTaskDetailShortcuts({
	task: () => task.value,
	taskTitle: () => taskTitle.value,
	onSave: saveTask,
})

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

async function changeProject(project: IProject | null) {
	if (project === null) {
		return
	}
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

async function duplicateCurrentTask() {
	const duplicatedTask = await taskStore.duplicateTask(task.value.id)
	if (duplicatedTask) {
		success({message: t('task.detail.duplicateSuccess')})
		router.push({
			name: 'task.detail',
			params: {id: duplicatedTask.id},
		})
	}
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

.checklist-summary {
	padding-inline-start: .25rem;
}

.detail-content {
	@media print {
		inline-size: 100% !important;
	}
}

// --- Sidebar ---------------------------------------------------------------

// On narrow screens the properties belong above the description, not below the
// comments. Bulma stacks columns as blocks there, so flex is needed for `order`.
.task-columns {
	@media screen and (max-width: $tablet) {
		display: flex;
		flex-direction: column;
	}
}

// Not sticky: the property list can be taller than the viewport and its popups
// open downwards, so a stuck sidebar would leave them unreachable.
.task-sidebar {
	@media screen and (max-width: $tablet) {
		order: -1;
	}
}

.sidebar-toolbar {
	display: flex;
	align-items: stretch;
	gap: .375rem;
	margin-block-end: .75rem;

	.button--mark-done {
		flex: 1;
		justify-content: flex-start;
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

.task-actions-menu {
	display: inline-flex;

	.toolbar-icon-button {
		block-size: 100%;
	}
}

// --- Layout switch ---------------------------------------------------------

.sidebar-layout-switch {
	display: inline-flex;
	gap: .125rem;
	padding: .125rem;
	margin-block-end: .5rem;
	border-radius: $radius;
	background: var(--grey-100);
	font-size: .75rem;

	.base-button {
		display: inline-flex;
		align-items: center;
		gap: .375rem;
		padding: .25rem .625rem;
		border-radius: calc(#{$radius} - 2px);
		color: var(--grey-500);
		transition: color $transition, background-color $transition, box-shadow $transition;

		svg {
			inline-size: .75rem;
		}

		&:hover {
			color: var(--text);
		}

		&.is-active {
			color: var(--text);
			background: var(--scheme-main);
			box-shadow: var(--shadow-xs);
		}
	}
}

// --- Property list ---------------------------------------------------------

.property-list {
	margin: 0;
	border-block-start: 1px solid var(--grey-200);
	font-size: .9375rem;
}

.property-group {
	display: none;
}

// Cards layout: label above value, short fields two per row, grouped under headings.
// Same markup as the rows layout; grid `order` regroups the fields.
.property-list.is-cards {
	display: grid;
	grid-template-columns: repeat(2, minmax(0, 1fr));
	gap: .5rem;
	border-block-start: none;

	.property-group {
		display: block;
		grid-column: 1 / -1;
		margin: .5rem 0 0;
		font-size: .8125rem;
		font-weight: 600;
		color: var(--grey-500);
	}

	.property-group-schedule {
		order: 0;
		margin-block-start: 0;
	}

	.property-due-date { order: 1; }
	.property-start-date { order: 2; }
	.property-end-date { order: 3; }
	.property-repeat { order: 4; }
	.property-reminders { order: 5; }
	.property-group-details { order: 6; }
	.property-priority { order: 7; }
	.property-percent-done { order: 8; }
	.property-color { order: 9; }
	.property-labels { order: 10; }
	.property-assignees { order: 11; }

	.property {
		display: flex;
		flex-direction: column;
		align-items: stretch;
		gap: .125rem;
		padding: .5rem .625rem;
		border: 1px solid var(--grey-200);
		border-radius: $radius;
		background: var(--scheme-main);

		&:hover,
		&:focus-within {
			border-color: var(--grey-300);
		}

		dt {
			min-block-size: 0;
			font-size: .75rem;
			font-weight: 600;
			letter-spacing: .01em;
		}

		dd {
			flex-wrap: wrap;
		}

		&.is-set {
			border-inline-start: 3px solid var(--primary);
		}

		// Editors are flat on the card, so the card itself is the affordance.
		:deep(.select select),
		:deep(.datepicker .show),
		.property-value-button {
			padding-inline-start: 0;

			&:hover {
				background: transparent;
			}
		}

		:deep(.multiselect .input-wrapper) {
			padding-inline-start: 0;
		}
	}

	.property-assignees,
	.property-labels,
	.property-reminders,
	.property-repeat.is-expanded {
		grid-column: 1 / -1;
	}

	.property-editor {
		margin-block-start: .5rem;
		padding-inline-start: 0;
		border-inline-start: none;
	}
}

.property {
	display: grid;
	grid-template-columns: minmax(6rem, 7.5rem) 1fr;
	column-gap: .5rem;
	align-items: start;
	padding-block: .375rem;
	border-block-end: 1px solid var(--grey-200);
	transition: background-color $transition;

	dt {
		display: flex;
		align-items: center;
		gap: .5rem;
		// Line up the label with the value's first line, whose control has its own vertical padding.
		min-block-size: 2rem;
		color: var(--grey-500);
		font-weight: 500;
		transition: color $transition;

		svg {
			inline-size: .875rem;
			flex-shrink: 0;
			color: var(--grey-400);
			transition: color $transition;
		}
	}

	dd {
		min-inline-size: 0;
		margin: 0;
		display: flex;
		align-items: center;
		gap: .25rem;

		> * {
			min-inline-size: 0;
		}
	}

	// A filled row reads darker than an empty one so the state is visible at a glance.
	&.is-set dt {
		color: var(--grey-700);

		svg {
			color: var(--primary);
		}
	}

	.remove {
		color: var(--grey-400);
		line-height: 1;
		padding: .25rem;
		border-radius: $radius;
		opacity: 0;
		transition: opacity $transition, color $transition;

		&:hover,
		&:focus-visible {
			color: var(--danger);
		}
	}

	&:hover .remove,
	&:focus-within .remove {
		opacity: 1;
	}
}

.property-empty {
	color: var(--grey-400);
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
}

// Controls sit flat on the row and only reveal their chrome on interaction.
.property-list .property {
	:deep(.select) {
		inline-size: 100%;

		select {
			inline-size: 100%;
			background: transparent;
			border-color: transparent;
			box-shadow: none;
			block-size: 2rem;
			padding-block: 0;
			padding-inline-start: .5rem;
			line-height: 2rem;
		}

		&:not(.has-defaults):after {
			opacity: 0;
		}

		&:hover:not(.has-defaults):after {
			opacity: 1;
		}

		select:not(:disabled) {
			cursor: pointer;

			&:hover,
			&:focus {
				background: var(--scheme-main);
				border-color: var(--border);
			}
		}
	}

	:deep(.multiselect) {
		inline-size: 100%;
		min-inline-size: 0;

		// Native inputs have an intrinsic width that would push the row past the sidebar.
		input {
			min-inline-size: 0;
			inline-size: 100%;
		}

		.input-wrapper {
			&:not(:focus-within, :hover) {
				background: transparent;
				border-color: transparent;
			}
		}
	}

	:deep(.datepicker) {
		flex: 1;
		min-inline-size: 0;
		position: relative;

		// The sidebar is narrower than the popup, so anchor it to the row's end edge.
		.datepicker-popup {
			inset-inline-end: 0;
			inset-inline-start: auto;
		}

		.show {
			color: var(--text);
			padding: .25rem .5rem;
			transition: background-color $transition;
			border-radius: $radius;
			display: block;
			inline-size: 100%;
			min-block-size: 2rem;
			text-align: start;

			i {
				color: var(--grey-400);
			}

			&:hover {
				background: var(--scheme-main);
			}
		}

		&.disabled .show:hover {
			background: transparent;
		}
	}

	:deep(.reminders) {
		inline-size: 100%;
		padding-block-start: .25rem;

		> * {
			position: relative;
		}

		.popup {
			inset-inline-end: 0;
			inset-inline-start: auto;
		}
	}

	:deep(.color-picker-container) {
		justify-content: flex-start;
		min-block-size: 2rem;
		padding-inline-start: .5rem;

		.button {
			text-transform: none;
			font-size: .8125rem;
			font-weight: 500;
			color: var(--grey-500);
			border-color: transparent;
			padding: .25rem .5rem;

			&:hover {
				color: var(--text);
				border-color: var(--grey-300);
			}
		}
	}

	:deep(.assignees-list) {
		padding-block-start: .375rem;
	}
}

.date-input {
	display: flex;
	align-items: center;
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
	inset-block-end: 2.5rem;
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
</style>
