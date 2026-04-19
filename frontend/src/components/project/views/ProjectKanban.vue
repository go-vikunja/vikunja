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
									class="bucket-header kb-col-head"
									@click="() => unCollapseBucket(bucket)"
								>
									<div class="kb-col-title-wrap">
										<div class="kb-col-title">
											<span
												class="kb-col-dot"
												:style="{background: getBucketColor(bucket, bucketIndex)}"
											/>
											<h2
												class="title input"
												:contenteditable="(bucketTitleEditable && canWrite && !collapsedBuckets[bucket.id]) ? true : undefined"
												:spellcheck="false"
												@keydown.enter.prevent.stop="!$event.isComposing && ($event.target as HTMLElement).blur()"
												@keydown.esc.prevent.stop="!$event.isComposing && ($event.target as HTMLElement).blur()"
												@blur="saveBucketTitle(bucket.id, ($event.target as HTMLElement).textContent as string)"
												@click.stop="focusBucketTitle"
											>
												{{ bucket.title }}
											</h2>
											<span
												v-if="bucket.limit > 0 || alwaysShowBucketTaskCount"
												:class="{'is-max': bucket.limit > 0 && bucket.count >= bucket.limit}"
												class="limit kb-col-count"
											>
												{{ bucket.limit > 0 ? `${bucket.count}/${bucket.limit}` : bucket.count }}
											</span>
											<span
												v-if="bucket.id !== 0 && view?.doneBucketId === bucket.id"
												v-tooltip="$t('project.kanban.doneBucketHint')"
												class="kb-done-indicator"
												@click.stop="() => collapseBucket(bucket)"
											>
												<Icon icon="check-double" />
											</span>
										</div>
									</div>
									<Dropdown
										v-if="canWrite && !collapsedBuckets[bucket.id]"
										class="is-right options kb-col-menu-dropdown"
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
													class="input kb-plain-input"
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
									<template #footer>
										<div
											v-if="canCreateTasks"
											class="bucket-footer"
										>
											<div
												v-if="showNewTaskInput === bucket.id"
												class="kb-inline-add"
											>
												<div
													class="control"
													:class="{'is-loading': loading || taskLoading}"
												>
													<input
														v-model="newTaskText"
														v-focus.always
														class="input kb-inline-input"
														:disabled="loading || taskLoading || undefined"
														:placeholder="$t('project.kanban.addTaskPlaceholder')"
														type="text"
														@focusout="toggleShowNewTaskInput(bucket.id)"
														@focusin="() => newTaskInputFocused = true"
														@keyup.enter="addTaskToBucket(bucket.id)"
														@keyup.esc="toggleShowNewTaskInput(bucket.id)"
													>
												</div>
												<p
													v-if="newTaskError[bucket.id] && newTaskText === ''"
													class="help is-danger"
												>
													{{ $t('project.create.addTitleRequired') }}
												</p>
											</div>
											<XButton
												v-else
												v-tooltip="bucket.limit > 0 && bucket.count >= bucket.limit ? $t('project.kanban.bucketLimitReached') : ''"
												class="kb-add-btn"
												:shadow="false"
												icon="plus"
												variant="secondary"
												:disabled="bucket.limit > 0 && bucket.count >= bucket.limit"
												@click="toggleShowNewTaskInput(bucket.id)"
											>
												{{ $t('project.kanban.addTask') }}
											</XButton>
										</div>
									</template>

									<template #item="{element: task}">
										<div
											class="task-item kb-card"
											:data-task-id="task.id"
											@click="openTask(task)"
										>
											<span
												v-if="canWrite && isTouchDevice"
												class="handle"
												@click.stop="openTask(task)"
												@touchstart.passive="onHandleTouchStart"
												@touchmove.passive="onHandleTouchMove"
											/>
											<p class="kb-card-id">
												#{{ task.identifier || task.id }}
											</p>
											<p class="kb-card-title">
												{{ task.title }}
											</p>
											<div class="kb-card-sep" />
											<div class="kb-card-foot">
												<span
													v-for="label in task.labels"
													:key="label.id"
													class="kb-tag"
													:style="getLabelStyle(label)"
												>
													{{ label.title }}
												</span>
												<span
													v-if="task.priority > 0"
													class="kb-tag"
													:class="getPriorityClass(task.priority)"
												>
													{{ getPriorityLabel(task.priority) }}
												</span>
												<div
													v-if="task.assignees.length > 0"
													class="kb-assignee"
												>
													{{ getAssigneeInitials(task.assignees[0]) }}
												</div>
											</div>
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
							class="input kb-plain-input"
							:placeholder="$t('project.kanban.addBucketPlaceholder')"
							type="text"
							@blur="() => showNewBucketInput = false"
							@keyup.enter="createNewBucket"
							@keyup.esc="($event.target as HTMLInputElement).blur()"
						>
						<XButton
							v-else
							:shadow="false"
							class="kb-new-bucket"
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
</template>

<script setup lang="ts">
import {computed, nextTick, ref, watch, toRef} from 'vue'
import {useRouter} from 'vue-router'
import {useRouteQuery} from '@vueuse/router'
import {useI18n} from 'vue-i18n'
import draggable from 'zhyswan-vuedraggable'
import {klona} from 'klona/lite'

import {PERMISSIONS as Permissions} from '@/constants/permissions'
import {PRIORITIES} from '@/constants/priorities'
import BucketModel from '@/models/bucket'

import type {IBucket} from '@/modelTypes/IBucket'
import type {ILabel} from '@/modelTypes/ILabel'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'
import {useAuthStore} from '@/stores/auth'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
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

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	viewId: IProjectView['id'],
}>()

const projectId = toRef(props, 'projectId')

const DRAG_OPTIONS = {
	// sortable options
	animation: 150,
	ghostClass: 'ghost',
	dragClass: 'task-dragging',
	delayOnTouchOnly: true,
	delay: 1000,
} as const

const BUCKET_COLORS = ['#6c63f5', '#f59e0b', '#10b981', '#ef4444', '#06b6d4', '#f97316']

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

function getBucketColor(bucket: IBucket, bucketIndex: number): string {
	if (bucket.id !== 0 && view.value?.doneBucketId === bucket.id) {
		return '#10b981'
	}

	if (bucket.limit > 0 && bucket.count >= bucket.limit) {
		return '#ef4444'
	}

	return BUCKET_COLORS[bucketIndex % BUCKET_COLORS.length]
}

function getLabelStyle(label: ILabel) {
	return {
		borderColor: `${label.hexColor}66`,
		color: label.textColor || label.hexColor,
		background: `${label.hexColor}2b`,
	}
}

function getPriorityClass(priority: ITask['priority']) {
	if (priority >= PRIORITIES.HIGH) {
		return 'tag--high'
	}

	if (priority === PRIORITIES.MEDIUM) {
		return 'tag--med'
	}

	return 'tag--low'
}

function getPriorityLabel(priority: ITask['priority']) {
	if (priority >= PRIORITIES.HIGH) {
		return t('task.priority.high')
	}

	if (priority === PRIORITIES.MEDIUM) {
		return t('task.priority.medium')
	}

	return t('task.priority.low')
}

function getAssigneeInitials(user: IUser) {
	const source = user.name || user.username || user.email || ''
	const parts = source.trim().split(/\s+/).filter(Boolean)

	if (parts.length === 0) {
		return '?'
	}

	if (parts.length === 1) {
		return parts[0].slice(0, 2).toUpperCase()
	}

	return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
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
			newTask.bucketId = updatedTaskBucket.bucketId
			if (updatedTaskBucket.bucketId !== newTask.bucketId) {
				kanbanStore.moveTaskToBucket(newTask, updatedTaskBucket.bucketId)
			}
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
}

async function addTaskToBucket(bucketId: IBucket['id']) {
	if (newTaskText.value === '') {
		newTaskError.value[bucketId] = true
		return
	}
	newTaskError.value[bucketId] = false

	const task = await taskStore.createNewTask({
		title: newTaskText.value,
		bucketId,
		projectId: projectIdWithFallback.value,
	})
	newTaskText.value = ''
	kanbanStore.addTaskToBucket(task)
	scrollTaskContainerToTop(bucketId)

	const bucket = kanbanStore.getBucketById(bucketId)
	if (bucket && bucket.limit && bucket.count >= bucket.limit) {
		toggleShowNewTaskInput(bucketId)
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
</style>


<style lang="scss">
$ease-out: all .3s cubic-bezier(0.23, 1, 0.32, 1);
$bucket-width: 256px;
$bucket-header-height: auto;
$bucket-right-margin: 16px;
$crazy-height-calculation: '100vh - 4.5rem - 1.5rem - 1rem - 1.5rem - 11px';
$filter-container-height: '1rem - #{$switch-view-height}';

.kanban {
	overflow-x: auto;
	overflow-y: hidden;
	block-size: calc(#{$crazy-height-calculation});
	margin: 0;
	padding: 20px 24px;
	align-items: flex-start;
	gap: $bucket-right-margin;
	background: #0f1016;

	&:focus, .bucket .tasks:focus {
		box-shadow: none;
	}

	@media screen and (max-width: $tablet) {
		block-size: calc(#{$crazy-height-calculation} - #{$filter-container-height} + 9px);
		scroll-snap-type: x mandatory;
		margin: 0;
		padding: 16px 12px;
	}

	&-bucket-container {
		display: flex;
		align-items: flex-start;
		gap: $bucket-right-margin;
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
		position: relative;
		margin: 0;
		max-block-size: 100%;
		inline-size: $bucket-width;
		display: flex;
		flex-direction: column;
		gap: 10px;
		flex-shrink: 0;
		overflow: visible;

		@media screen and (max-width: $tablet) {
			scroll-snap-align: center;
		}

		.tasks {
			overflow: hidden auto;
			block-size: 100%;
			display: flex;
			flex-direction: column;
			gap: 10px;
			padding: 0;
			margin: 0;
			list-style: none;
		}

		.task-item {
			position: relative;
			padding: 13px 14px;
			margin: 0;
			background: #13141a;
			border: .5px solid #1e1f28;
			border-radius: 10px;
			cursor: pointer;
			transition: border-color .15s, background .15s;

			&:hover {
				border-color: #2a2b35;
				background: #161720;
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
			font-size: 12.5px;
			margin: 0;
			font-weight: 500 !important;
		}

		&.new-bucket {
			min-inline-size: auto;
			background: transparent;

			.button {
				inline-size: auto;
			}
		}

		&.is-collapsed {
			align-self: flex-start;
			transform: rotate(90deg) translateY(-100%);
			transform-origin: top left;
			margin-inline-end: calc((#{$bucket-width} - #{$bucket-header-height} - #{$bucket-right-margin}) * -1);
			cursor: pointer;

			.tasks, .bucket-footer {
				display: none;
			}
		}
	}

	.bucket-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 2px;
		block-size: $bucket-header-height;
		background: transparent;

		.icon.has-text-success {
			cursor: pointer;
		}

		.limit {
			padding: 1px 7px;
			font-weight: 500;

			&.is-max {
				color: #f87171;
			}
		}

		.title.input {
			block-size: auto;
			padding: 0;
			display: inline-block;
			cursor: pointer;
			background: transparent;
			border: 0;
			color: #8a8a9a;
		}
	}

	:deep(.dropdown-trigger) {
		padding: 0;
	}

	:deep(.dropdown-trigger .button) {
		inline-size: 26px;
		block-size: 26px;
		padding: 0;
		border: 0;
		border-radius: 6px;
		background: transparent;
		color: #3a3b48;
		box-shadow: none;
		transition: all .12s;

		&:hover,
		&:focus-visible {
			background: #1a1b22;
			color: #8a8a9a;
		}
	}

	.bucket-footer {
		position: static;
		block-size: min-content;
		padding: 0;
		background: transparent;
		transform: none;

		.button {
			background-color: transparent;
		}
	}
}

.kanban-view {
	background: #0f1016;
	border-radius: 12px;
}

.kb-col-title-wrap {
	min-inline-size: 0;
	flex: 1;
}

.kb-col-title {
	display: flex;
	align-items: center;
	gap: 8px;
	font-size: 12.5px;
	font-weight: 500;
	color: #8a8a9a;
	min-inline-size: 0;
}

.kb-col-dot {
	inline-size: 7px;
	block-size: 7px;
	border-radius: 50%;
	flex-shrink: 0;
}

.kb-col-count {
	font-size: 11px;
	color: #3a3b48;
	background: #1a1b22;
	border-radius: 10px;
	line-height: 1.4;
}

.kb-col-menu-dropdown {
	flex-shrink: 0;
}

.kb-done-indicator {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	color: #34d399;
	font-size: 12px;
	flex-shrink: 0;
}

.kb-card {
	box-sizing: border-box;
}

.kb-card-id {
	font-size: 10.5px;
	color: #4a4b57;
	margin: 0 0 6px;
}

.kb-card-title {
	font-size: 13px;
	color: #c0bdb8;
	line-height: 1.45;
	margin: 0 0 11px;
	font-weight: 400;
}

.kb-card-sep {
	block-size: .5px;
	background: #1a1b22;
	margin-block-end: 10px;
}

.kb-card-foot {
	display: flex;
	align-items: center;
	gap: 5px;
	flex-wrap: wrap;
}

.kb-tag {
	font-size: 10.5px;
	padding: 2px 8px;
	border-radius: 20px;
	border: .5px solid;
	white-space: nowrap;
	text-transform: capitalize;
	line-height: 1.4;
}

.tag--high {
	border-color: rgb(224 72 72 / 25%);
	color: #f87171;
	background: rgb(224 72 72 / 8%);
}

.tag--med {
	border-color: rgb(245 158 11 / 25%);
	color: #fbbf24;
	background: rgb(245 158 11 / 8%);
}

.tag--low {
	border-color: rgb(16 185 129 / 25%);
	color: #34d399;
	background: rgb(16 185 129 / 8%);
}

.kb-assignee {
	inline-size: 20px;
	block-size: 20px;
	border-radius: 50%;
	background: #6c63f5;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 9px;
	color: #fff;
	font-weight: 600;
	margin-inline-start: auto;
	flex-shrink: 0;
	text-transform: uppercase;
}

.kb-inline-add {
	background: #13141a;
	border: .5px solid #6c63f5;
	border-radius: 10px;
	padding: 10px 12px;
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.kb-inline-input {
	background: transparent;
	border: 0;
	outline: none;
	font-size: 13px;
	color: #c0bdb8;
	font-family: inherit;
	inline-size: 100%;
	box-shadow: none;
	padding: 0;

	&::placeholder {
		color: #3a3b48;
	}
}

.kb-plain-input {
	background: #13141a;
	border: .5px solid #2a2b35;
	border-radius: 8px;
	font-size: 13px;
	color: #c0bdb8;
	font-family: inherit;
	box-shadow: none;
	padding: 8px 10px;

	&::placeholder {
		color: #3a3b48;
	}

	&:focus {
		border-color: #6c63f5;
		box-shadow: none;
	}
}

.kb-inline-actions {
	display: flex;
	gap: 6px;
}

.kb-btn-confirm,
.kb-btn-cancel {
	border-radius: 6px;
	font-size: 12px;
	font-family: inherit;
	cursor: pointer;
	padding: 4px 12px;
	transition: all .12s;
}

.kb-btn-confirm {
	background: #6c63f5;
	color: #fff;
	border: 0;
	padding-inline: 14px;

	&:hover:not(:disabled) {
		opacity: .85;
	}

	&:disabled {
		opacity: .5;
		cursor: not-allowed;
	}
}

.kb-btn-cancel {
	background: transparent;
	color: #8a8a9a;
	border: .5px solid #2a2b35;

	&:hover {
		background: #1e1f28;
	}
}

.kb-add-btn,
.kb-new-bucket {
	border: .5px dashed #2a2b35 !important;
	background: transparent !important;
	color: #3a3b48 !important;
	border-radius: 8px;
	transition: all .12s;
	box-shadow: none !important;
	justify-content: flex-start;
}

.kb-add-btn {
	display: flex;
	align-items: center;
	gap: 7px;
	padding: 8px 10px;
	font-size: 12.5px;
	font-family: inherit;
	inline-size: 100%;
	margin-block-start: 2px;

	&:hover:not(:disabled) {
		border-color: #6c63f5 !important;
		color: #8a8a9a !important;
		background: #13141a !important;
	}
}

.kb-new-bucket {
	display: flex;
	align-items: center;
	gap: 8px;
	padding: 10px 14px;
	font-size: 12.5px;
	font-family: inherit;
	height: 42px;
	white-space: nowrap;
	flex-shrink: 0;

	&:hover {
		border-color: #6c63f5 !important;
		color: #a78bfa !important;
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
