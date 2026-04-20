<template>
	<ProjectWrapper
		class="project-kanban"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #header>
			<div class="filter-container">
				<FilterPopup
					v-if="!isSavedFilter(project)"
					v-model="params"
					:view-id="viewId"
					:project-id="projectId"
					@update:modelValue="updateFilters"
				/>
			</div>
		</template>

		<template #default>
			<div class="kanban-view">
				<div
					:class="{ 'is-loading': loading && !oneTaskUpdating}"
					class="kanban kanban-bucket-container loader-container"
				>
					<draggable
						v-bind="DRAG_OPTIONS"
						:model-value="buckets"
						group="buckets"
						:disabled="!canWrite || newTaskInputFocused"
						tag="ul"
						:item-key="({id}: IBucket) => `bucket${id}`"
						:component-data="bucketDraggableComponentData"
						@update:modelValue="updateBuckets"
						@end="updateBucketPosition"
						@start="() => dragBucket = true"
					>
						<template #item="{element: bucket, index: bucketIndex }">
							<div
								class="bucket"
								:class="{'is-collapsed': collapsedBuckets[bucket.id]}"
							>
								<div
									class="bucket-header"
									@click="() => unCollapseBucket(bucket)"
								>
									<span
										v-if="bucket.id !== 0 && view?.doneBucketId === bucket.id"
										v-tooltip="$t('project.kanban.doneBucketHint')"
										class="icon is-small has-text-success mie-2"
										@click.stop="() => collapseBucket(bucket)"
									>
										<Icon icon="check-double" />
									</span>
									<h2
										class="title input"
										:contenteditable="(bucketTitleEditable && canWrite && !collapsedBuckets[bucket.id]) ? true : undefined"
										:spellcheck="false"
										@keydown.enter.prevent.stop="!$event.isComposing && ($event.target as HTMLElement).blur()"
										@keydown.esc.prevent.stop="!$event.isComposing && ($event.target as HTMLElement).blur()"
										@blur="saveBucketTitle(bucket.id, ($event.target as HTMLElement).textContent as string)"
										@click="focusBucketTitle"
									>
										{{ bucket.title }}
									</h2>
									<span
										v-if="bucket.limit > 0 || alwaysShowBucketTaskCount"
										:class="{'is-max': bucket.limit > 0 && bucket.count >= bucket.limit}"
										class="limit"
									>
										{{ bucket.limit > 0 ? `${bucket.count}/${bucket.limit}` : bucket.count }}
									</span>
									<Dropdown
										v-if="canWrite && !collapsedBuckets[bucket.id]"
										class="is-right options"
										trigger-icon="ellipsis-v"
										@close="() => showSetLimitInput = false"
									>
										<div
											v-if="showSetLimitInput"
											class="field has-addons"
										>
											<div class="control">
												<input
													ref="bucketLimitInputRef"
													v-focus.always
													:value="bucket.limit"
													class="input"
													type="number"
													min="0"
													@keyup.esc="() => showSetLimitInput = false"
													@keyup.enter="() => {setBucketLimit(bucket.id, true); showSetLimitInput = false}"
													@input="setBucketLimit(bucket.id)"
												>
											</div>
											<div class="control">
												<XButton
													v-cy="'setBucketLimit'"
													:disabled="bucket.limit < 0"
													:icon="['far', 'save']"
													:shadow="false"
													@click="() => {setBucketLimit(bucket.id, true); showSetLimitInput = false}"
												/>
											</div>
										</div>
										<DropdownItem
											v-else
											@click.stop="showSetLimitInput = true"
										>
											{{
												$t('project.kanban.limit', {limit: bucket.limit > 0 ? bucket.limit : $t('project.kanban.noLimit')})
											}}
										</DropdownItem>
										<DropdownItem
											v-tooltip="$t('project.kanban.doneBucketHintExtended')"
											:icon-class="{'has-text-success': bucket.id === view?.doneBucketId}"
											icon="check-double"
											@click.stop="toggleDoneBucket(bucket)"
										>
											{{ $t('project.kanban.doneBucket') }}
										</DropdownItem>
										<DropdownItem
											v-tooltip="$t('project.kanban.defaultBucketHint')"
											:icon-class="{'has-text-primary': bucket.id === view?.defaultBucketId}"
											icon="th"
											@click.stop="toggleDefaultBucket(bucket)"
										>
											{{ $t('project.kanban.defaultBucket') }}
										</DropdownItem>
										<DropdownItem
											icon="angles-up"
											@click.stop="() => collapseBucket(bucket)"
										>
											{{ $t('project.kanban.collapse') }}
										</DropdownItem>
										<DropdownItem
											v-tooltip="buckets.length <= 1 ? $t('project.kanban.deleteLast') : ''"
											class="has-text-danger"
											:class="{'is-disabled': buckets.length <= 1}"
											icon-class="has-text-danger"
											icon="trash-alt"
											@click.stop="() => deleteBucketModal(bucket.id)"
										>
											{{ $t('misc.delete') }}
										</DropdownItem>
									</Dropdown>
								</div>

								<div
									v-if="canCreateTasks"
									class="bucket-top"
								>
									<div
										v-if="showNewTaskInput === bucket.id"
										class="field add-task-inline"
										:class="{'has-task-color': newTaskColor !== '', 'has-light-text': newTaskColor !== '' && !colorIsDark(newTaskColor)}"
										:style="newTaskColor ? {'--task-color': newTaskColor} : undefined"
										@focusout="handleAddTaskFocusOut($event, bucket.id)"
									>
										<div
											class="control add-task-inline__control"
											:class="{'is-loading': loading || taskLoading}"
										>
											<input
												v-model="newTaskText"
												v-focus.always
												class="input"
												:disabled="loading || taskLoading || undefined"
												:placeholder="$t('project.kanban.addTaskPlaceholder')"
												type="text"
												@focusin="() => newTaskInputFocused = true"
												@keyup.enter="addTaskToBucket(bucket.id)"
												@keyup.esc="toggleShowNewTaskInput(bucket.id)"
											>
											<button
												v-if="newTaskText.trim() !== ''"
												type="button"
												class="add-task-inline__submit"
												:disabled="loading || taskLoading || undefined"
												:title="$t('misc.save')"
												@click="addTaskToBucket(bucket.id)"
											>
												<Icon icon="check" />
											</button>
										</div>
										<p
											v-if="newTaskError[bucket.id] && newTaskText === ''"
											class="help is-danger"
										>
											{{ $t('project.create.addTitleRequired') }}
										</p>
										<div
											v-if="hasAnyInlineQuickAddField"
											class="inline-quick-add-chip-bar"
										>
											<button
												v-for="chip in inlineChips"
												:key="chip.field"
												type="button"
												class="inline-quick-add-chip"
												:class="[`inline-quick-add-chip--${chip.modifier}`, {'is-set': chip.isSet}]"
												:disabled="loading || taskLoading || undefined"
												@click.stop="toggleInlinePopup(chip.popup, $event)"
											>
												<span
													v-if="chip.colorValue"
													class="inline-quick-add-chip__swatch"
													:style="{background: chip.colorValue}"
												/>
												<Icon
													v-else
													:icon="chip.icon"
													class="inline-quick-add-chip__icon"
													:class="`inline-quick-add-chip__icon--${chip.modifier}`"
												/>
												<span>{{ chip.label }}</span>
												<span
													v-if="chip.isSet"
													class="inline-quick-add-chip__clear"
													@click.stop="clearInlineField(chip.field)"
												>
													<Icon icon="times" />
												</span>
											</button>
										</div>
									</div>
									<XButton
										v-else
										v-tooltip="bucket.limit > 0 && bucket.count >= bucket.limit ? $t('project.kanban.bucketLimitReached') : ''"
										class="is-fullwidth has-text-centered"
										:shadow="false"
										icon="plus"
										variant="secondary"
										:disabled="bucket.limit > 0 && bucket.count >= bucket.limit"
										@click="toggleShowNewTaskInput(bucket.id)"
									>
										{{
											bucket.tasks.length === 0 ? $t('project.kanban.addTask') : $t('project.kanban.addAnotherTask')
										}}
									</XButton>
								</div>

								<draggable
									v-bind="DRAG_OPTIONS"
									:handle="taskDragHandle"
									:delay="isTouchDevice ? 300 : 1000"
									:model-value="bucket.tasks"
									:group="{name: 'tasks', put: shouldAcceptDrop(bucket) && !dragBucket}"
									:disabled="!canWrite"
									:data-bucket-index="bucketIndex"
									tag="ul"
									:item-key="(task: ITask) => `bucket${bucket.id}-task${task.id}`"
									:component-data="getTaskDraggableTaskComponentData(bucket)"
									@update:modelValue="(tasks) => updateTasks(bucket.id, tasks)"
									@start="handleTaskDragStart"
									@end="updateTaskPosition"
								>
									<template #item="{element: task}">
										<div
											class="task-item"
											:data-task-id="task.id"
										>
											<span
												v-if="canWrite && isTouchDevice"
												class="handle"
												@click="openTask(task)"
												@touchstart.passive="onHandleTouchStart"
												@touchmove.passive="onHandleTouchMove"
											/>
											<KanbanCard
												class="kanban-card"
												:task="task"
												:loading="taskUpdating[task.id] ?? false"
												:project-id="projectId"
												@taskCompletedRecurring="handleRecurringTaskCompletion"
											/>
										</div>
									</template>
								</draggable>
							</div>
						</template>
					</draggable>

					<div
						v-if="canWrite && !loading && buckets.length > 0"
						class="bucket new-bucket"
					>
						<input
							v-if="showNewBucketInput"
							v-model="newBucketTitle"
							v-focus.always
							:class="{'is-loading': loading}"
							:disabled="loading || undefined"
							class="input"
							:placeholder="$t('project.kanban.addBucketPlaceholder')"
							type="text"
							@blur="() => showNewBucketInput = false"
							@keyup.enter="createNewBucket"
							@keyup.esc="($event.target as HTMLInputElement).blur()"
						>
						<XButton
							v-else
							:shadow="false"
							class="is-transparent is-fullwidth has-text-centered"
							variant="secondary"
							icon="plus"
							@click="() => showNewBucketInput = true"
						>
							{{ $t('project.kanban.addBucket') }}
						</XButton>
					</div>
				</div>

				<Modal
					:enabled="showBucketDeleteModal"
					@close="showBucketDeleteModal = false"
					@submit="deleteBucket()"
				>
					<template #header>
						<span>{{ $t('project.kanban.deleteHeaderBucket') }}</span>
					</template>

					<template #text>
						<p>
							{{ $t('project.kanban.deleteBucketText1') }}<br>
							{{ $t('project.kanban.deleteBucketText2') }}
						</p>
					</template>
				</Modal>
			</div>
		</template>
	</ProjectWrapper>
	<Teleport to="body">
		<div
			v-if="openInlinePopup !== null"
			ref="inlinePopupRef"
			class="inline-quick-add-popup"
			:class="[
				`inline-quick-add-popup--${popupVariant}`,
				openInlinePopup === 'reminder' ? 'inline-quick-add-popup--wide' : null,
				isPopupReady ? null : 'inline-quick-add-popup--measuring',
			]"
			:style="{top: `${inlinePopupPosition.top}px`, left: `${inlinePopupPosition.left}px`}"
		>
			<DatepickerInline
				v-if="openInlinePopup === 'due'"
				v-model="newTaskDueDate"
			/>
			<DatepickerInline
				v-else-if="openInlinePopup === 'start'"
				v-model="newTaskStartDate"
			/>
			<ul
				v-else-if="openInlinePopup === 'priority'"
				class="inline-quick-add-priority-options"
			>
				<li
					v-for="option in PRIORITY_OPTIONS"
					:key="option.value"
				>
					<button
						type="button"
						class="inline-quick-add-priority-option"
						:class="{'is-active': newTaskPriority === option.value}"
						@click="selectPriority(option.value)"
					>
						{{ $t(option.labelKey) }}
					</button>
				</li>
			</ul>
			<EditAssignees
				v-else-if="openInlinePopup === 'assignee'"
				v-model="newTaskAssignees"
				:task-id="0"
				:project-id="projectIdWithFallback"
			/>
			<EditLabels
				v-else-if="openInlinePopup === 'labels'"
				v-model="newTaskLabels"
				:task-id="0"
				:creatable="false"
			/>
			<Reminders
				v-else-if="openInlinePopup === 'reminder'"
				v-model="newTaskReminders"
				:default-relative-to="reminderDefaultRelativeTo"
			/>
			<DatepickerInline
				v-else-if="openInlinePopup === 'endDate'"
				v-model="newTaskEndDate"
			/>
			<ColorPicker
				v-else-if="openInlinePopup === 'color'"
				v-model="newTaskColor"
			/>
			<div
				v-else-if="openInlinePopup === 'percentDone'"
				class="inline-quick-add-percent-done"
			>
				<input
					v-model.number="newTaskPercentDone"
					type="range"
					min="0"
					max="100"
					step="10"
					class="inline-quick-add-percent-done__slider"
				>
				<span class="inline-quick-add-percent-done__label">{{ newTaskPercentDone }}%</span>
			</div>
			<XButton
				v-if="openInlinePopup !== 'priority'"
				class="inline-quick-add-popup__confirm"
				:shadow="false"
				@click="openInlinePopup = null"
			>
				{{ $t('misc.confirm') }}
			</XButton>
		</div>
	</Teleport>
</template>

<script setup lang="ts">
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch, toRef} from 'vue'
import {useRouter} from 'vue-router'
import {useRouteQuery} from '@vueuse/router'
import {useI18n} from 'vue-i18n'
import draggable from 'zhyswan-vuedraggable'
import {klona} from 'klona/lite'

import {PERMISSIONS as Permissions} from '@/constants/permissions'
import BucketModel from '@/models/bucket'

import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'
import type {ILabel} from '@/modelTypes/ILabel'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import type {IReminderPeriodRelativeTo} from '@/types/IReminderPeriodRelativeTo'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'
import {useAuthStore} from '@/stores/auth'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import KanbanCard from '@/components/tasks/partials/KanbanCard.vue'
import DatepickerInline from '@/components/input/DatepickerInline.vue'
import EditAssignees from '@/components/tasks/partials/EditAssignees.vue'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import Reminders from '@/components/tasks/partials/Reminders.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import {DEFAULT_INLINE_QUICK_ADD_FIELDS} from '@/modelTypes/IUserSettings'
import {formatDateShort} from '@/helpers/time/formatDate'
import {closeWhenClickedOutside} from '@/helpers/closeWhenClickedOutside'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	viewId: IProjectView['id'],
}>()
const PRIORITY_LABEL_KEYS: Record<number, string> = {
	1: 'low',
	2: 'medium',
	3: 'high',
	4: 'urgent',
	5: 'doNow',
}
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'

import {
	type CollapsedBuckets,
	getCollapsedBucketState,
	saveCollapsedBucketState,
} from '@/helpers/saveCollapsedBucketState'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'

import {isSavedFilter, useSavedFilter} from '@/services/savedFilter'
import {useTaskDragToProject} from '@/composables/useTaskDragToProject'
import {success} from '@/message'
import {useProjectStore} from '@/stores/projects'
import type {TaskFilterParams} from '@/services/taskCollection'
import type {IProjectView} from '@/modelTypes/IProjectView'
import TaskPositionService from '@/services/taskPosition'
import TaskPositionModel from '@/models/taskPosition'
import {i18n} from '@/i18n'
import ProjectViewService from '@/services/projectViews'
import ProjectViewModel from '@/models/projectView'
import TaskBucketService from '@/services/taskBucket'
import TaskBucketModel from '@/models/taskBucket'

const projectId = toRef(props, 'projectId')

const DRAG_OPTIONS = {
	// sortable options
	animation: 150,
	ghostClass: 'ghost',
	dragClass: 'task-dragging',
	delayOnTouchOnly: true,
	delay: 1000,
} as const

const MIN_SCROLL_HEIGHT_PERCENT = 0.25

const {t} = useI18n({useScope: 'global'})

const baseStore = useBaseStore()
const kanbanStore = useKanbanStore()
const taskStore = useTaskStore()
const projectStore = useProjectStore()
const authStore = useAuthStore()

const alwaysShowBucketTaskCount = computed(() => authStore.settings.frontendSettings.alwaysShowBucketTaskCount)
const {handleTaskDropToProject} = useTaskDragToProject()
const taskPositionService = ref(new TaskPositionService())
const taskBucketService = ref(new TaskBucketService())

// Saved filter composable for accessing filter data
const savedFilter = useSavedFilter(() => isSavedFilter({id: projectId.value}) ? projectId.value : undefined).filter

const taskContainerRefs = ref<{ [id: IBucket['id']]: HTMLElement }>({})
const bucketLimitInputRef = ref<HTMLInputElement | null>(null)

const drag = ref(false)
const dragBucket = ref(false)
const sourceBucket = ref(0)

const showBucketDeleteModal = ref(false)
const bucketToDelete = ref(0)
const bucketTitleEditable = ref(false)

const newTaskText = ref('')
const showNewTaskInput = ref<IBucket['id'] | null>(null)

const newBucketTitle = ref('')
const showNewBucketInput = ref(false)
const newTaskError = ref<{ [id: IBucket['id']]: boolean }>({})
const newTaskInputFocused = ref(false)

// Inline quick-add field state. Only one bucket's add-task form is open at
// a time, so a single value per field is enough.
const newTaskDueDate = ref<Date | null>(null)
const newTaskStartDate = ref<Date | null>(null)
const newTaskEndDate = ref<Date | null>(null)
const newTaskPriority = ref<number>(0)
const newTaskAssignees = ref<IUser[]>([])
const newTaskLabels = ref<ILabel[]>([])
const newTaskReminders = ref<ITaskReminder[]>([])
const newTaskColor = ref<string>('')
const newTaskPercentDone = ref<number>(0)

type InlinePopup = 'due' | 'start' | 'endDate' | 'assignee' | 'labels' | 'reminder' | 'priority' | 'color' | 'percentDone' | null
const openInlinePopup = ref<InlinePopup>(null)
const inlinePopupRef = ref<HTMLElement | null>(null)
const inlinePopupPosition = ref<{top: number, left: number}>({top: 0, left: 0})
// Hides the popup for its first render frame so the position clamp can
// run before the user sees it. Without this, the popup briefly flashes
// at the un-clamped position and then jumps into place.
const isPopupReady = ref(false)

const PRIORITY_OPTIONS = [
	{value: 0, labelKey: 'task.priority.unset'},
	{value: 1, labelKey: 'task.priority.low'},
	{value: 2, labelKey: 'task.priority.medium'},
	{value: 3, labelKey: 'task.priority.high'},
	{value: 4, labelKey: 'task.priority.urgent'},
	{value: 5, labelKey: 'task.priority.doNow'},
] as const

function selectPriority(value: number) {
	newTaskPriority.value = value
	openInlinePopup.value = null
}

// Dates render the two-column shortcuts+calendar layout; the pickers
// (assignee/labels/reminder) use the simpler padded card layout. With the
// add-task form moved to the top of the bucket, both have plenty of room
// to expand downward, so a viewport-scale fallback is no longer needed.
const popupVariant = computed(() => {
	if (openInlinePopup.value === 'due' || openInlinePopup.value === 'start' || openInlinePopup.value === 'endDate') {
		return 'date'
	}
	return 'picker'
})

// Default new reminders relative to the due date if one is set; otherwise
// leave absolute so the reminder detail input picks its own default.
const reminderDefaultRelativeTo = computed<IReminderPeriodRelativeTo | null>(
	() => newTaskDueDate.value !== null ? REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE : null,
)

const anchorChipRect = ref<DOMRect | null>(null)
let popupResizeObserver: ResizeObserver | null = null

function toggleInlinePopup(which: Exclude<InlinePopup, null>, event: MouseEvent) {
	if (openInlinePopup.value === which) {
		openInlinePopup.value = null
		return
	}
	const chip = event.currentTarget as HTMLElement
	const rect = chip.getBoundingClientRect()
	anchorChipRect.value = rect
	// Initial placement; clampInlinePopupToViewport refines it once the
	// popup has rendered and we know its actual dimensions.
	inlinePopupPosition.value = {
		top: rect.bottom + 4,
		left: rect.left,
	}
	isPopupReady.value = false
	openInlinePopup.value = which
	nextTick(() => {
		clampInlinePopupToViewport()
		// Reveal after the clamp has written the final position so the
		// popup never renders at the wrong spot.
		isPopupReady.value = true
		observePopupResize()
	})
}

function clampInlinePopupToViewport() {
	const popup = inlinePopupRef.value
	const chipRect = anchorChipRect.value
	if (!popup || !chipRect) {
		return
	}
	const margin = 8
	const popupRect = popup.getBoundingClientRect()
	const top = chipRect.bottom + 4
	let left = chipRect.left

	// Shift left so the right edge stays inside the viewport. If the popup
	// is wider than the viewport itself (rare — very narrow windows), pin
	// it to the left margin; CSS max-inline-size keeps it from overflowing
	// the right edge in that case.
	if (left + popupRect.width + margin > window.innerWidth) {
		left = Math.max(margin, window.innerWidth - popupRect.width - margin)
	}

	inlinePopupPosition.value = {top, left}
}

function observePopupResize() {
	disconnectPopupResize()
	const popup = inlinePopupRef.value
	if (!popup || typeof ResizeObserver === 'undefined') {
		return
	}
	// The popup can grow after its initial render (Multiselect dropdown
	// opens, ReminderDetail switches to the date+time form, etc.). Re-clamp
	// on every resize so horizontal overflow is always corrected.
	// Defer the position update to the next animation frame so we never
	// write to layout inside the ResizeObserver callback itself — doing so
	// triggers "ResizeObserver loop completed with undelivered notifications".
	popupResizeObserver = new ResizeObserver(() => {
		requestAnimationFrame(() => clampInlinePopupToViewport())
	})
	popupResizeObserver.observe(popup)
}

function disconnectPopupResize() {
	if (popupResizeObserver) {
		popupResizeObserver.disconnect()
		popupResizeObserver = null
	}
}

watch(openInlinePopup, (value) => {
	if (value === null) {
		disconnectPopupResize()
		anchorChipRect.value = null
		isPopupReady.value = false
	}
})

// Auto-close the assignee picker after a selection is made. Picking an
// assignee is usually a single action; leaving the picker open makes the
// form feel stuck. Users who want more assignees can reopen the chip.
watch(() => newTaskAssignees.value.length, (newLen, oldLen) => {
	if (newLen > oldLen && openInlinePopup.value === 'assignee') {
		openInlinePopup.value = null
	}
})

function onDocumentClickForInlinePopup(e: MouseEvent) {
	if (openInlinePopup.value !== null && inlinePopupRef.value) {
		closeWhenClickedOutside(e, inlinePopupRef.value, () => {
			openInlinePopup.value = null
		})
	}
}

onMounted(() => document.addEventListener('click', onDocumentClickForInlinePopup))
onBeforeUnmount(() => {
	document.removeEventListener('click', onDocumentClickForInlinePopup)
	disconnectPopupResize()
})

const enabledInlineQuickAddFields = computed(
	() => authStore.settings.frontendSettings.inlineQuickAddFields ?? DEFAULT_INLINE_QUICK_ADD_FIELDS,
)
const showInlineAssignee = computed(() => enabledInlineQuickAddFields.value.includes('assignee'))
const showInlineDueDate = computed(() => enabledInlineQuickAddFields.value.includes('dueDate'))
const showInlineStartDate = computed(() => enabledInlineQuickAddFields.value.includes('startDate'))
const showInlinePriority = computed(() => enabledInlineQuickAddFields.value.includes('priority'))
const showInlineLabels = computed(() => enabledInlineQuickAddFields.value.includes('labels'))
const showInlineReminder = computed(() => enabledInlineQuickAddFields.value.includes('reminder'))
const showInlineEndDate = computed(() => enabledInlineQuickAddFields.value.includes('endDate'))
const showInlineColor = computed(() => enabledInlineQuickAddFields.value.includes('color'))
const showInlinePercentDone = computed(() => enabledInlineQuickAddFields.value.includes('percentDone'))
const hasAnyInlineQuickAddField = computed(() => enabledInlineQuickAddFields.value.length > 0)

const assigneeChipLabel = computed(() => {
	const count = newTaskAssignees.value.length
	if (count === 0) return t('task.attributes.assignees')
	if (count === 1) return newTaskAssignees.value[0].name || newTaskAssignees.value[0].username
	return t('task.attributes.assigneesN', count)
})

const labelsChipLabel = computed(() => {
	const count = newTaskLabels.value.length
	if (count === 0) return t('task.attributes.labels')
	if (count === 1) return newTaskLabels.value[0].title
	return t('task.attributes.labelsN', count)
})

const reminderChipLabel = computed(() => {
	const count = newTaskReminders.value.length
	if (count === 0) return t('task.attributes.reminders')
	return t('task.attributes.remindersN', count)
})

type InlineChip = {
	field: string
	modifier: string
	icon: string
	popup: Exclude<InlinePopup, null>
	isSet: boolean
	label: string
	colorValue?: string
}

const CHIP_CONFIG: Record<string, {modifier: string, icon: string, popup: Exclude<InlinePopup, null>}> = {
	assignee: {modifier: 'assignee', icon: 'user', popup: 'assignee'},
	dueDate: {modifier: 'due', icon: 'calendar', popup: 'due'},
	startDate: {modifier: 'start', icon: 'play', popup: 'start'},
	endDate: {modifier: 'end', icon: 'stop', popup: 'endDate'},
	priority: {modifier: 'priority', icon: 'exclamation', popup: 'priority'},
	labels: {modifier: 'labels', icon: 'tags', popup: 'labels'},
	reminder: {modifier: 'reminder', icon: 'bell', popup: 'reminder'},
	color: {modifier: 'color', icon: 'fill-drip', popup: 'color'},
	percentDone: {modifier: 'percent', icon: 'percent', popup: 'percentDone'},
}

const inlineChips = computed<InlineChip[]>(() => {
	const chipLabel: Record<string, () => string> = {
		assignee: () => assigneeChipLabel.value,
		dueDate: () => newTaskDueDate.value !== null ? formatDateShort(newTaskDueDate.value) : t('task.attributes.dueDate'),
		startDate: () => newTaskStartDate.value !== null ? formatDateShort(newTaskStartDate.value) : t('task.attributes.startDate'),
		endDate: () => newTaskEndDate.value !== null ? formatDateShort(newTaskEndDate.value) : t('task.attributes.endDate'),
		priority: () => newTaskPriority.value !== 0 ? t(`task.priority.${PRIORITY_LABEL_KEYS[newTaskPriority.value]}`) : t('task.attributes.priority'),
		labels: () => labelsChipLabel.value,
		reminder: () => reminderChipLabel.value,
		color: () => t('task.attributes.color'),
		percentDone: () => newTaskPercentDone.value > 0 ? `${newTaskPercentDone.value}%` : t('task.attributes.percentDone'),
	}
	const chipIsSet: Record<string, () => boolean> = {
		assignee: () => newTaskAssignees.value.length > 0,
		dueDate: () => newTaskDueDate.value !== null,
		startDate: () => newTaskStartDate.value !== null,
		endDate: () => newTaskEndDate.value !== null,
		priority: () => newTaskPriority.value !== 0,
		labels: () => newTaskLabels.value.length > 0,
		reminder: () => newTaskReminders.value.length > 0,
		color: () => newTaskColor.value !== '',
		percentDone: () => newTaskPercentDone.value > 0,
	}

	return enabledInlineQuickAddFields.value.map(field => {
		const cfg = CHIP_CONFIG[field]
		return {
			field,
			modifier: cfg.modifier,
			icon: cfg.icon,
			popup: cfg.popup,
			isSet: chipIsSet[field](),
			label: chipLabel[field](),
			colorValue: field === 'color' && newTaskColor.value ? newTaskColor.value : undefined,
		}
	})
})

function clearInlineField(field: string) {
	const clearMap: Record<string, () => void> = {
		assignee: () => { newTaskAssignees.value = [] },
		dueDate: () => { newTaskDueDate.value = null },
		startDate: () => { newTaskStartDate.value = null },
		endDate: () => { newTaskEndDate.value = null },
		priority: () => { newTaskPriority.value = 0 },
		labels: () => { newTaskLabels.value = [] },
		reminder: () => { newTaskReminders.value = [] },
		color: () => { newTaskColor.value = '' },
		percentDone: () => { newTaskPercentDone.value = 0 },
	}
	clearMap[field]?.()
}

function resetNewTaskInlineFields() {
	newTaskDueDate.value = null
	newTaskStartDate.value = null
	newTaskEndDate.value = null
	newTaskPriority.value = 0
	newTaskAssignees.value = []
	newTaskLabels.value = []
	newTaskReminders.value = []
	newTaskColor.value = ''
	newTaskPercentDone.value = 0
	openInlinePopup.value = null
}

const showSetLimitInput = ref(false)
const collapsedBuckets = ref<CollapsedBuckets>({})

// We're using this to show the loading animation only at the task when updating it
const taskUpdating = ref<{ [id: ITask['id']]: boolean }>({})
const oneTaskUpdating = ref(false)

// URL-synchronized filter parameters
const filter = useRouteQuery('filter')
const s = useRouteQuery('s')

const params = ref<TaskFilterParams>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	s: '',
})

watch([filter, s], ([filterValue, sValue]) => {
	params.value.filter = filterValue ?? ''
	params.value.s = sValue ?? ''
}, { immediate: true })

function updateFilters(newParams: TaskFilterParams) {
	// Update all params
	params.value = { ...newParams }
	
	// Sync only filter and s to URL
	filter.value = newParams.filter || undefined
	s.value = newParams.s || undefined
}

const getTaskDraggableTaskComponentData = computed(() => (bucket: IBucket) => {
	return {
		ref: (el: HTMLElement) => setTaskContainerRef(bucket.id, el),
		onScroll: (event: Event) => handleTaskContainerScroll(bucket.id, event.target as HTMLElement),
		type: 'transition-group',
		name: !drag.value ? 'move-card' : null,
		class: [
			'tasks',
			{'dragging-disabled': !canWrite.value},
		],
	}
})

const bucketDraggableComponentData = computed(() => ({
	type: 'transition-group',
	name: !dragBucket.value ? 'move-bucket' : null,
	class: [
		'kanban-bucket-container',
		{'dragging-disabled': !canWrite.value},
	],
}))
const project = computed(() => projectId.value ? projectStore.projects[projectId.value] : null)
const view = computed(() => project.value?.views.find(v => v.id === props.viewId) as IProjectView || null)
const canWrite = computed(() => baseStore.currentProject?.maxPermission > Permissions.READ && view.value.bucketConfigurationMode === 'manual')
const canCreateTasks = computed(() => canWrite.value && projectId.value > 0)

const isTouchDevice = ref(false)
if (typeof window !== 'undefined') {
	isTouchDevice.value = !window.matchMedia('(hover: hover) and (pointer: fine)').matches
}
const taskDragHandle = computed(() => isTouchDevice.value ? '.handle' : undefined)

const router = useRouter()
const touchStartY = ref(0)

function openTask(task: ITask) {
	router.push({
		name: 'task.detail',
		params: {id: task.id},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

function onHandleTouchStart(e: TouchEvent) {
	touchStartY.value = e.touches[0].clientY
}

function onHandleTouchMove(e: TouchEvent) {
	if (drag.value) return

	const currentY = e.touches[0].clientY
	const deltaY = touchStartY.value - currentY
	const scrollContainer = (e.target as HTMLElement).closest('.tasks') as HTMLElement | null
	if (scrollContainer) {
		scrollContainer.scrollTop += deltaY
		touchStartY.value = currentY
	}
}

const buckets = computed(() => kanbanStore.buckets)
const loading = computed(() => kanbanStore.isLoading)
const projectIdWithFallback = computed<number>(() => project.value?.id || projectId.value)

const taskLoading = computed(() => taskStore.isLoading || taskPositionService.value.loading)

watch(
	() => ({
		params: params.value,
		projectId: projectId.value,
		viewId: props.viewId,
	}),
	({params, projectId, viewId}) => {
		if (projectId === undefined || Number(projectId) === 0) {
			return
		}
		collapsedBuckets.value = getCollapsedBucketState(projectId)
		kanbanStore.loadBucketsForProject(projectId, viewId, params)
	},
	{
		immediate: true,
		deep: true,
	},
)

function setTaskContainerRef(id: IBucket['id'], el: HTMLElement) {
	if (!el) return
	taskContainerRefs.value[id] = el
}

function handleTaskContainerScroll(id: IBucket['id'], el: HTMLElement) {
	if (!el) {
		return
	}
	const scrollTopMax = el.scrollHeight - el.clientHeight
	const threshold = el.scrollTop + el.scrollTop * MIN_SCROLL_HEIGHT_PERCENT
	if (scrollTopMax > threshold) {
		return
	}

	kanbanStore.loadNextTasksForBucket(
		projectId.value,
		props.viewId,
		params.value,
		id,
	)
}

function updateTasks(bucketId: IBucket['id'], tasks: IBucket['tasks']) {
	const bucket = kanbanStore.getBucketById(bucketId)

	if (bucket === undefined) {
		return
	}

	kanbanStore.setBucketById({
		...bucket,
		tasks,
	})
}

async function updateTaskPosition(e) {
	drag.value = false

	// Check if dropped on a sidebar project
	const {moved} = await handleTaskDropToProject(e, (task) => {
		kanbanStore.removeTaskInBucket(task)
	})

	if (moved) {
		return
	}

	// If dropped outside kanban
	if (!e.to.dataset.bucketIndex) {
		return
	}

	// While we could just pass the bucket index in through the function call, this would not give us the
	// new bucket id when a task has been moved between buckets, only the new bucket. Using the data-bucket-id
	// of the drop target works all the time.
	const bucketIndex = parseInt(e.to.dataset.bucketIndex)

	const newBucket = buckets.value[bucketIndex]

	// HACK:
	// this is a hacky workaround for a known problem of vue.draggable.next when using the footer slot
	// the problem: https://github.com/SortableJS/vue.draggable.next/issues/108
	// This hack doesn't remove the problem that the ghost item is still displayed below the footer
	// It just makes releasing the item possible.

	// The newIndex of the event doesn't count in the elements of the footer slot.
	// This is why in case the length of the tasks is identical with the newIndex
	// we have to remove 1 to get the correct index.
	const newTaskIndex = newBucket.tasks.length === e.newIndex
		? e.newIndex - 1
		: e.newIndex

	const task = newBucket.tasks[newTaskIndex]
	const oldBucket = buckets.value.find(b => b.id === sourceBucket.value)
	const taskBefore = newBucket.tasks[newTaskIndex - 1] ?? null
	const taskAfter = newBucket.tasks[newTaskIndex + 1] ?? null
	taskUpdating.value[task.id] = true

	const newTask = klona(task) // cloning the task to avoid pinia store manipulation
	newTask.bucketId = newBucket.id
	const position = calculateItemPosition(
		taskBefore !== null ? taskBefore.position : null,
		taskAfter !== null ? taskAfter.position : null,
	)
	
	let bucketHasChanged = false
	if (
		oldBucket !== undefined && // This shouldn't actually be `undefined`, but let's play it safe.
		newBucket.id !== oldBucket.id
	) {
		kanbanStore.setBucketById({
			...oldBucket,
			count: oldBucket.count - 1,
		})
		kanbanStore.setBucketById({
			...newBucket,
			count: newBucket.count + 1,
		})
		bucketHasChanged = true
	}

	try {
		const newPosition = new TaskPositionModel({
			position,
			projectViewId: props.viewId,
			taskId: newTask.id,
		})
		await taskPositionService.value.update(newPosition)
		newTask.position = position
		
		if(bucketHasChanged) {
			const updatedTaskBucket = await taskBucketService.value.update(new TaskBucketModel({
				taskId: newTask.id,
				bucketId: newTask.bucketId,
				projectViewId: props.viewId,
				projectId: projectIdWithFallback.value,
			}))
			Object.assign(newTask, updatedTaskBucket.task)
			if (updatedTaskBucket.bucketId !== newTask.bucketId) {
				kanbanStore.moveTaskToBucket(newTask, updatedTaskBucket.bucketId)
			}
			newTask.bucketId = updatedTaskBucket.bucketId
			if (updatedTaskBucket.bucket) {
				kanbanStore.setBucketById(updatedTaskBucket.bucket, false)
			}
		}
		kanbanStore.setTaskInBucket(newTask)

		// Make sure the first and second task don't both get position 0 assigned
		if (newTaskIndex === 0 && taskAfter !== null && taskAfter.position === 0) {
			const taskAfterAfter = newBucket.tasks[newTaskIndex + 2] ?? null
			const newTaskAfter = klona(taskAfter) // cloning the task to avoid pinia store manipulation
			newTaskAfter.bucketId = newBucket.id
			newTaskAfter.position = calculateItemPosition(
				0,
				taskAfterAfter !== null ? taskAfterAfter.position : null,
			)

			await taskStore.update(newTaskAfter)
		}
	} finally {
		taskUpdating.value[task.id] = false
		oneTaskUpdating.value = false
	}
}

function toggleShowNewTaskInput(bucketId: IBucket['id']) {
	if (loading.value || taskLoading.value) {
		return
	}
	showNewTaskInput.value = showNewTaskInput.value === bucketId
		? null
		: bucketId
	newTaskInputFocused.value = false
	// Whether the user is opening or closing the form, start from a clean
	// slate. This also fixes a stale-title bug where dismissing the form
	// (by clicking outside) left newTaskText around, so the next opening
	// pre-filled the old draft.
	newTaskText.value = ''
	newTaskError.value[bucketId] = false
	resetNewTaskInlineFields()
}

function handleAddTaskFocusOut(event: FocusEvent, bucketId: IBucket['id']) {
	// Inline quick-add popups are teleported to <body>, so focus legitimately
	// moves outside this container while one is open. Treat the popup as part
	// of the add-task workflow and keep the form alive until the popup closes.
	if (openInlinePopup.value !== null) {
		return
	}
	const container = event.currentTarget as HTMLElement | null
	const nextFocus = event.relatedTarget as HTMLElement | null
	if (container && nextFocus && container.contains(nextFocus)) {
		return
	}
	if (nextFocus && inlinePopupRef.value?.contains(nextFocus)) {
		return
	}
	toggleShowNewTaskInput(bucketId)
}

async function addTaskToBucket(bucketId: IBucket['id']) {
	if (newTaskText.value === '') {
		newTaskError.value[bucketId] = true
		return
	}
	newTaskError.value[bucketId] = false

	// Capture inline field values before resetting so the follow-up attach
	// calls aren't affected by the UI state being cleared.
	const capturedAssignees = showInlineAssignee.value ? [...newTaskAssignees.value] : []
	const capturedLabels = showInlineLabels.value ? [...newTaskLabels.value] : []
	const capturedReminders = showInlineReminder.value ? [...newTaskReminders.value] : []

	const task = await taskStore.createNewTask({
		title: newTaskText.value,
		bucketId,
		projectId: projectIdWithFallback.value,
		dueDate: showInlineDueDate.value && newTaskDueDate.value !== null
			? newTaskDueDate.value
			: undefined,
		startDate: showInlineStartDate.value && newTaskStartDate.value !== null
			? newTaskStartDate.value
			: undefined,
		endDate: showInlineEndDate.value && newTaskEndDate.value !== null
			? newTaskEndDate.value
			: undefined,
		priority: showInlinePriority.value && newTaskPriority.value !== 0
			? newTaskPriority.value
			: undefined,
		hexColor: showInlineColor.value && newTaskColor.value !== ''
			? newTaskColor.value
			: undefined,
		percentDone: showInlinePercentDone.value && newTaskPercentDone.value > 0
			? newTaskPercentDone.value
			: undefined,
	})
	kanbanStore.addTaskToBucket(task)
	scrollTaskContainerToTop(bucketId)
	// Close the add-task form after a successful submission. Upstream keeps
	// it open for rapid entry; we intentionally diverge because a new task
	// popping into the bucket while the user is still looking at the form
	// was jarring — each submit should be an explicit action.
	// Also clears newTaskText and the inline quick-add fields via the
	// shared reset path.
	if (showNewTaskInput.value === bucketId) {
		toggleShowNewTaskInput(bucketId)
	}

	// Attach multi-valued fields that the create endpoint doesn't accept
	// directly. Sequential to keep ordering stable and errors isolated; the
	// task itself is already saved, so a failed extra attach should not
	// block the rest.
	for (const user of capturedAssignees) {
		await taskStore.addAssignee({user, taskId: task.id})
	}
	for (const label of capturedLabels) {
		await taskStore.addLabel({label, taskId: task.id})
	}
	if (capturedReminders.length > 0) {
		await taskStore.update({...task, reminders: capturedReminders})
	}
}

function scrollTaskContainerToTop(bucketId: IBucket['id']) {
	const bucketEl = taskContainerRefs.value[bucketId]
	if (!bucketEl) {
		return
	}
	bucketEl.scrollTop = 0
}

async function createNewBucket() {
	if (newBucketTitle.value === '') {
		return
	}

	await kanbanStore.createBucket(new BucketModel({
		title: newBucketTitle.value,
		projectId: projectIdWithFallback.value,
		projectViewId: props.viewId,
	}))
	newBucketTitle.value = ''
}

function deleteBucketModal(bucketId: IBucket['id']) {
	if (buckets.value.length <= 1) {
		return
	}

	bucketToDelete.value = bucketId
	showBucketDeleteModal.value = true
}

async function deleteBucket() {
	try {
		await kanbanStore.deleteBucket({
			bucket: new BucketModel({
				id: bucketToDelete.value,
				projectId: projectIdWithFallback.value,
				projectViewId: props.viewId,
			}),
			params: params.value,
		})
		success({message: t('project.kanban.deleteBucketSuccess')})
	} finally {
		showBucketDeleteModal.value = false
	}
}

/** This little helper allows us to drag a bucket around at the title without focusing on it right away. */
async function focusBucketTitle(e: Event) {
	bucketTitleEditable.value = true
	await nextTick()
	const target = e.target as HTMLInputElement
	target.focus()
}

async function saveBucketTitle(bucketId: IBucket['id'], bucketTitle: string) {
	
	const bucket = kanbanStore.getBucketById(bucketId)
	if (bucket?.title === bucketTitle) {
		bucketTitleEditable.value = false
		return
	}
	
	await kanbanStore.updateBucket({
		id: bucketId,
		title: bucketTitle,
		projectId: projectId.value,
	})
	success({message: i18n.global.t('project.kanban.bucketTitleSavedSuccess')})
	bucketTitleEditable.value = false
}

function updateBuckets(value: IBucket[]) {
	// (1) buckets get updated in store and tasks positions get invalidated
	kanbanStore.setBuckets(value)
}

function handleRecurringTaskCompletion() {
	// Only reload if we're in a saved filter and the filter contains date fields
	if (!isSavedFilter(project.value)) {
		return
	}

	const filterContainsDateFields = savedFilter.value?.filters?.filter?.includes('due_date') ||
		savedFilter.value?.filters?.filter?.includes('start_date') ||
		savedFilter.value?.filters?.filter?.includes('end_date')
		
	if (filterContainsDateFields) {
		// Reload the kanban board to refresh tasks that now match/don't match the filter
		kanbanStore.loadBucketsForProject(projectId.value, props.viewId, params.value)
	}
}

// TODO: fix type
function updateBucketPosition(e: { newIndex: number }) {
	// (2) bucket positon is changed
	dragBucket.value = false

	const bucket = buckets.value[e.newIndex]
	const bucketBefore = buckets.value[e.newIndex - 1] ?? null
	const bucketAfter = buckets.value[e.newIndex + 1] ?? null

	kanbanStore.updateBucket({
		id: bucket.id,
		projectId: projectId.value,
		position: calculateItemPosition(
			bucketBefore !== null ? bucketBefore.position : null,
			bucketAfter !== null ? bucketAfter.position : null,
		),
	})
}

async function saveBucketLimit(bucketId: IBucket['id'], limit: number) {
	if (limit < 0) {
		return
	}

	await kanbanStore.updateBucket({
		...kanbanStore.getBucketById(bucketId),
		projectId: projectId.value,
		limit,
	})
	success({message: t('project.kanban.bucketLimitSavedSuccess')})
}

const setBucketLimitCancel = ref<number | null>(null)

async function setBucketLimit(bucketId: IBucket['id'], now: boolean = false) {
	const limit = parseInt(bucketLimitInputRef.value?.value || '')

	if (setBucketLimitCancel.value !== null) {
		clearTimeout(setBucketLimitCancel.value)
	}

	if (now) {
		return saveBucketLimit(bucketId, limit)
	}

	setBucketLimitCancel.value = setTimeout(saveBucketLimit, 2500, bucketId, limit)
}

function shouldAcceptDrop(bucket: IBucket) {
	return (
		// When dragging from a bucket who has its limit reached, dragging should still be possible
		bucket.id === sourceBucket.value ||
		// If there is no limit set, dragging & dropping should always work
		bucket.limit === 0 ||
		// Disallow dropping to buckets which have their limit reached
		bucket.count < bucket.limit
	)
}

function dragstart(bucket: IBucket) {
	drag.value = true
	sourceBucket.value = bucket.id
}

function handleTaskDragStart(e) {
	const taskId = parseInt(e.item.dataset.taskId, 10)
	const bucketIndex = parseInt(e.from.dataset.bucketIndex, 10)
	const bucket = buckets.value[bucketIndex]
	const task = bucket?.tasks.find(t => t.id === taskId)

	if (task) {
		taskStore.setDraggedTask(task)
	}
	dragstart(bucket)
}

async function toggleDefaultBucket(bucket: IBucket) {
	const defaultBucketId = view.value?.defaultBucketId === bucket.id
		? 0
		: bucket.id

	const projectViewService = new ProjectViewService()
	const updatedView = await projectViewService.update(new ProjectViewModel({
		...view.value,
		defaultBucketId,
	}))

	const views = project.value.views.map(v => v.id === view.value?.id ? updatedView : v)
	const updatedProject = {
		...project.value,
		views,
	}

	projectStore.setProject(updatedProject)

	success({message: t('project.kanban.defaultBucketSavedSuccess')})
}

async function toggleDoneBucket(bucket: IBucket) {
	const doneBucketId = view.value?.doneBucketId === bucket.id
		? 0
		: bucket.id
	
	const projectViewService = new ProjectViewService()
	const updatedView = await projectViewService.update(new ProjectViewModel({
		...view.value,
		doneBucketId,
	}))

	const views = project.value.views.map(v => v.id === view.value?.id ? updatedView : v)
	const updatedProject = {
		...project.value,
		views,
	}
	
	projectStore.setProject(updatedProject)
	
	success({message: t('project.kanban.doneBucketSavedSuccess')})
}

function collapseBucket(bucket: IBucket) {
	collapsedBuckets.value[bucket.id] = true
	saveCollapsedBucketState(projectIdWithFallback.value, collapsedBuckets.value)
}

function unCollapseBucket(bucket: IBucket) {
	if (!collapsedBuckets.value[bucket.id]) {
		return
	}

	collapsedBuckets.value[bucket.id] = false
	saveCollapsedBucketState(projectIdWithFallback.value, collapsedBuckets.value)
}
</script>

<style lang="scss" scoped>
.control.is-loading {
  &::after {
    inset-block-start: 30%;
    inset-inline-end: 50%;
    transform: translate(-50%, 0);

	--loader-border-color: var(--grey-500);
  }
}

.add-task-inline {
	padding: .5rem;
	background: var(--grey-200);
	border-radius: $radius;
	transition: background-color $transition;

	&.has-task-color {
		background-color: var(--task-color);

		.inline-quick-add-chip,
		.inline-quick-add-chip__icon,
		.add-task-inline__submit {
			color: rgba(0, 0, 0, .7);
		}

		.inline-quick-add-chip.is-set {
			background: rgba(255, 255, 255, .35);
			color: rgba(0, 0, 0, .8);
		}

		:deep(.input) {
			color: rgba(0, 0, 0, .85);

			&::placeholder {
				color: rgba(0, 0, 0, .45);
			}
		}
	}

	&.has-light-text {
		.inline-quick-add-chip,
		.inline-quick-add-chip__icon,
		.add-task-inline__submit {
			color: rgba(255, 255, 255, .85);
		}

		.inline-quick-add-chip.is-set {
			background: rgba(255, 255, 255, .2);
			color: #ffffff;
		}

		:deep(.input) {
			color: #ffffff;

			&::placeholder {
				color: rgba(255, 255, 255, .6);
			}
		}
	}

	// Let the title input blend into the container rather than stand on its own.
	:deep(.input) {
		background: transparent;
		border-color: transparent;
		box-shadow: none;
		// Reserve space on the right for the inline submit button.
		padding-inline-end: 2.25rem;

		&:focus {
			border-color: var(--primary);
		}
	}
}

.add-task-inline__control {
	position: relative;
}

.add-task-inline__submit {
	position: absolute;
	inset-block-start: 50%;
	inset-inline-end: .25rem;
	transform: translateY(-50%);
	display: inline-flex;
	align-items: center;
	justify-content: center;
	inline-size: 1.75rem;
	block-size: 1.75rem;
	padding: 0;
	border: 0;
	border-radius: $radius;
	background: transparent;
	color: var(--success);
	cursor: pointer;
	transition: background-color $transition, color $transition;

	&:hover:not(:disabled) {
		background: var(--success-light);
		color: var(--success-dark);
	}

	&:disabled {
		cursor: not-allowed;
		opacity: .5;
	}
}

.inline-quick-add-chip-bar {
	// Two-column grid with equal-width tracks. Chips stretch to fill their
	// column so full rows occupy the whole bar (not floating as an island),
	// and a partial last row lands in column 1 — left-aligned half-width.
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: .375rem;
	margin-block-start: .5rem;
}

.inline-quick-add-chip {
	position: relative;
	display: inline-flex;
	align-items: center;
	gap: .4rem;
	padding: .3rem .65rem;
	border: 1px solid transparent;
	border-radius: $radius;
	background: transparent;
	color: var(--grey-700);
	font-size: .8rem;
	font-weight: 500;
	line-height: 1.2;
	cursor: pointer;
	transition: background-color $transition, color $transition, border-color $transition, box-shadow $transition;

	&:hover:not(:disabled) {
		background: var(--white);
		color: var(--grey-900);
		box-shadow: 0 1px 3px hsla(var(--grey-900-hsl), .12);
	}

	&:focus-visible {
		outline: none;
		box-shadow: 0 0 0 2px var(--primary-light);
	}

	&:disabled {
		cursor: not-allowed;
		opacity: .5;
	}
}

// Semantic icon tint, applied whether the chip is set or not.
.inline-quick-add-chip__icon {
	font-size: .85rem;

	&--due {
		color: var(--danger);
	}

	&--start {
		color: var(--success);
	}

	&--priority {
		color: var(--warning);
	}

	&--assignee {
		color: var(--primary);
	}

	&--labels {
		color: var(--primary);
	}

	&--reminder {
		color: var(--primary);
	}

	&--end {
		color: var(--grey-500);
	}

	&--color {
		color: var(--grey-500);
	}

	&--percent {
		color: var(--grey-500);
	}
}

.inline-quick-add-chip__swatch {
	display: inline-block;
	inline-size: .85rem;
	block-size: .85rem;
	border-radius: .2rem;
	border: 1px solid var(--grey-300);
	flex-shrink: 0;
}

.inline-quick-add-chip__clear {
	display: none;
	align-items: center;
	justify-content: center;
	position: absolute;
	inset-inline-end: 0;
	inset-block-start: 0;
	block-size: 100%;
	padding-inline: .25rem;
	font-size: .65rem;
	background: transparent;
	border-radius: 0 $radius $radius 0;
	z-index: 1;
	color: inherit;

	&:hover {
		color: var(--danger);
	}
}

.inline-quick-add-chip:hover .inline-quick-add-chip__clear {
	display: flex;
}

// Set state: chip fills with the field's semantic light background and
// text shifts to the matching dark tone. No border — the color does the work.
.inline-quick-add-chip--due.is-set {
	background: var(--danger-light);
	color: var(--danger-dark);
}

.inline-quick-add-chip--start.is-set {
	background: var(--success-light);
	color: var(--success-dark);
}

.inline-quick-add-chip--priority.is-set {
	background: var(--warning-light);
	color: var(--warning-dark);
}

.inline-quick-add-chip--assignee.is-set,
.inline-quick-add-chip--labels.is-set,
.inline-quick-add-chip--reminder.is-set,
.inline-quick-add-chip--percent.is-set {
	background: var(--primary-light);
	color: var(--primary-dark);
}

.inline-quick-add-chip--end.is-set {
	background: var(--success-light);
	color: var(--success-dark);
}

.inline-quick-add-chip--color.is-set {
	background: var(--primary-light);
	color: var(--primary-dark);
}

.inline-quick-add-popup {
	position: fixed;
	z-index: 50;
	padding: .5rem;
	background: var(--white);
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	box-shadow: var(--shadow-md);
}

// Picker variant hosts assignee/label/reminder/priority components. The
// default compact width suits assignee, labels and priority. Reminder
// opts in to the wider variant so its nested two-column date form fits.
.inline-quick-add-popup--picker.inline-quick-add-popup--wide {
	inline-size: min(28rem, calc(100vw - 2rem));
}

// Keep the popup in the DOM but invisible until the clamp has positioned
// it. Using visibility (not display) preserves getBoundingClientRect so
// the clamp can measure the real size.
.inline-quick-add-popup--measuring {
	visibility: hidden;
}

.inline-quick-add-priority-options {
	display: flex;
	flex-direction: column;
	gap: .125rem;
	margin: 0;
	padding: 0;
	list-style: none;
}

.inline-quick-add-priority-option {
	inline-size: 100%;
	padding: .5rem .75rem;
	border: 0;
	border-radius: $radius;
	background: transparent;
	color: var(--text);
	text-align: start;
	font-size: .9rem;
	cursor: pointer;

	&:hover {
		background: var(--primary-light);
		color: var(--primary-dark);
	}

	&.is-active {
		background: var(--primary-light);
		color: var(--primary-dark);
		font-weight: 600;
	}
}

.inline-quick-add-percent-done {
	display: flex;
	align-items: center;
	gap: .75rem;
	padding: .5rem .25rem;

	&__slider {
		flex: 1;
		accent-color: var(--primary);
	}

	&__label {
		min-inline-size: 3rem;
		text-align: end;
		font-weight: 600;
		font-size: .9rem;
	}
}

.inline-quick-add-popup--picker {
	inline-size: min(18rem, calc(100vw - 2rem));

	:deep(.color-picker-container) {
		justify-content: start;
	}

	// The nested reminder Popup's .popup container defaults to natural
	// inline position (right-of-trigger), which can push the date form
	// past the outer popup's right edge. Pin it to the outer popup's
	// left edge instead so it stays within our already-clamped bounds.
	// !important is needed because Popup.vue's scoped .popup style wins
	// the specificity tie otherwise.
	:deep(.popup) {
		inset-inline-start: 0 !important;
		inset-inline-end: auto !important;
		inline-size: 100%;
	}

	// Constrain the nested card to the outer popup's content width so it
	// cannot overflow horizontally. The card's own scoped 310px width
	// (from ReminderDetail) is overridden here.
	:deep(.reminder-options-popup) {
		inline-size: 100% !important;
		max-inline-size: 100%;
	}

	:deep(.reminder-options-popup .datepicker-inline) {
		flex-direction: row;
		gap: .75rem;
		align-items: stretch;
	}

	:deep(.reminder-options-popup .datepicker-inline__shortcuts) {
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
	}

	:deep(.reminder-options-popup .datepicker-inline__shortcuts .datepicker__quick-select-date) {
		flex: 1 1 auto;
		block-size: auto;
	}

	:deep(.reminder-options-popup .flatpickr-container) {
		flex: 0 1 auto;
	}

	:deep(.reminder-options-popup .flatpickr-container > input) {
		display: none;
	}

	@media (width <= 520px) {
		:deep(.reminder-options-popup .datepicker-inline) {
			flex-direction: column;
		}
	}
}

.inline-quick-add-popup--date {
	display: flex;
	flex-direction: column;
	max-inline-size: calc(100vw - 1rem);
}

// Two-column layout: shortcuts (auto width) on the left, calendar on the
// right. flex: 1 on each shortcut makes them share the full calendar
// height, so the shortcut column stretches exactly as tall as the
// calendar+time regardless of whether 4, 5 or 6 shortcuts are rendered.
.inline-quick-add-popup--date :deep(.datepicker-inline) {
	flex-direction: row;
	gap: .75rem;
	align-items: stretch;
}

.inline-quick-add-popup--date :deep(.datepicker-inline__shortcuts) {
	display: flex;
	flex-direction: column;
	flex-shrink: 0;
}

.inline-quick-add-popup--date :deep(.datepicker-inline__shortcuts .datepicker__quick-select-date) {
	flex: 1 1 auto;
	block-size: auto;
}

.inline-quick-add-popup--date :deep(.flatpickr-container) {
	flex: 0 1 auto;
}

// Hide only flatpickr's top-level text input (duplicates the picker
// value as plain text). Time hour/minute inputs live inside
// .flatpickr-calendar and must stay visible — hence the direct-child
// selector.
.inline-quick-add-popup--date :deep(.flatpickr-container > input) {
	display: none;
}

.inline-quick-add-popup__confirm {
	inline-size: 100%;
	margin-block-start: .5rem;
}

// Fall back to vertical (original) layout when the viewport is too narrow
// to reasonably fit two columns.
@media (width <= 520px) {
	.inline-quick-add-popup--date :deep(.datepicker-inline) {
		flex-direction: column;
	}
}
</style>


<style lang="scss">
$ease-out: all .3s cubic-bezier(0.23, 1, 0.32, 1);
$bucket-width: 300px;
$bucket-header-height: 60px;
$bucket-right-margin: 1rem;
$crazy-height-calculation: '100vh - 4.5rem - 1.5rem - 1rem - 1.5rem - 11px';
$crazy-height-calculation-tasks: '#{$crazy-height-calculation} - 1rem - 2.5rem - 2rem - #{$button-height} - 1rem';
$filter-container-height: '1rem - #{$switch-view-height}';

.kanban {
	overflow-x: auto;
	overflow-y: hidden;
	block-size: calc(#{$crazy-height-calculation});
	margin: 0 -1.5rem;
	padding: 0 1.5rem;

	&:focus, .bucket .tasks:focus {
		box-shadow: none;
	}

	@media screen and (max-width: $tablet) {
		block-size: calc(#{$crazy-height-calculation} - #{$filter-container-height} + 9px);
		scroll-snap-type: x mandatory;
		margin: 0 -0.5rem;
	}

	&-bucket-container {
		display: flex;
	}

	.ghost {
		position: relative;

		* {
			opacity: 0;
		}

		&::after {
			content: '';
			position: absolute;
			display: block;
			inset-block-start: 0.25rem;
			inset-inline-end: 0.5rem;
			inset-block-end: 0.25rem;
			inset-inline-start: 0.5rem;
			border: 3px dashed var(--grey-300);
			border-radius: $radius;
		}
	}

	.bucket {
		border-radius: $radius;
		position: relative;

		margin: 0 $bucket-right-margin 0 0;
		max-block-size: calc(100% - 1rem); // 1rem spacing to the bottom
		min-block-size: 20px;
		inline-size: $bucket-width;
		display: flex;
		flex-direction: column;
		overflow: hidden; // Make sure the edges are always rounded

		@media screen and (max-width: $tablet) {
			scroll-snap-align: center;
		}

		.tasks {
			overflow: hidden auto;
			block-size: 100%;
		}

		.task-item {
			background-color: var(--grey-100);
			padding: .25rem .5rem;
			position: relative;

			&:first-of-type {
				padding-block-start: .5rem;
			}

			&:last-of-type {
				padding-block-end: .5rem;
			}

			.handle {
				position: absolute;
				inset: 0;
				z-index: 1;
				opacity: 0;
				touch-action: none;
				-webkit-touch-callout: none;
				user-select: none;
			}
		}

		.no-move {
			transition: transform 0s;
		}

		h2 {
			font-size: 1rem;
			margin: 0;
			font-weight: 600 !important;
		}

		&.new-bucket {
			// Because of reasons, this button ignores the margin we gave it to the right.
			// To make it still look like it has some, we modify the container to have a padding of 1rem,
			// which is the same as the margin it should have. Then we make the container itself bigger
			// to hide the fact we just made the button smaller.
			min-inline-size: calc(#{$bucket-width} + 1rem);
			background: transparent;

			.button {
				background: var(--grey-100);
				inline-size: 100%;
			}
		}

		&.is-collapsed {
			align-self: flex-start;
			transform: rotate(90deg) translateY(-100%);
			transform-origin: top left;
			// Using negative margins instead of translateY here to make all other buckets fill the empty space
			margin-inline-end: calc((#{$bucket-width} - #{$bucket-header-height} - #{$bucket-right-margin}) * -1);
			cursor: pointer;

			.tasks, .bucket-top {
				display: none;
			}
		}
	}

	.bucket-header {
		background-color: var(--grey-100);
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: .5rem;
		block-size: $bucket-header-height;

		.icon.has-text-success {
			cursor: pointer;
		}

		.limit {
			padding: 0 .5rem;
			font-weight: bold;

			&.is-max {
				color: var(--danger);
			}
		}

		.title.input {
			block-size: auto;
			padding: .4rem .5rem;
			display: inline-block;
			cursor: pointer;
		}
	}

	:deep(.dropdown-trigger) {
		padding: .5rem;
	}

	.bucket-top {
		position: sticky;
		inset-block-start: 0;
		z-index: 2;
		block-size: min-content;
		padding: .5rem;
		background-color: var(--grey-100);

		.button {
			background-color: transparent;

			&:hover {
				background-color: var(--white);
			}
		}
	}
}

// FIXME: This does not seem to work
.task-dragging {
	transform: rotateZ(3deg);
	transition: transform 0.18s ease;
}

.move-card-move {
	transform: rotateZ(3deg);
	transition: transform $transition-duration;
}

.move-card-leave-from,
.move-card-leave-to,
.move-card-leave-active {
	display: none;
}
</style>
