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
					v-if="!isSavedFilterProject(project)"
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
					:class="{ 'is-loading': initialLoading }"
					class="kanban kanban-bucket-container loader-container"
				>
					<draggable
						v-bind="DRAG_OPTIONS"
						:model-value="buckets"
						group="buckets"
						:disabled="!canWrite || newTaskInputFocused"
						tag="ul"
						:item-key="({id}: BucketResponse) => `bucket${id}`"
						:component-data="bucketDraggableComponentData"
						@update:modelValue="updateBuckets"
						@end="updateBucketPosition"
						@start="() => dragBucket = true"
					>
						<template #item="{element: bucket, index: bucketIndex }">
							<li
								class="bucket"
								:class="{'is-collapsed': collapsedBuckets[bucket.id]}"
								:data-bucket-id="bucket.id"
							>
								<div
									class="bucket-header"
									@click="() => unCollapseBucket(bucket)"
								>
									<span
										v-if="bucket.id !== 0 && view?.done_bucket_id === bucket.id"
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
										:trigger-label="$t('project.kanban.bucketOptions')"
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
													:aria-label="$t('misc.save')"
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
												$t('project.kanban.limit', {limit: bucket.limit > 0 ? bucket.limit : $t('misc.notSet')})
											}}
										</DropdownItem>
										<DropdownItem
											v-tooltip="$t('project.kanban.doneBucketHintExtended')"
											:icon-class="{'has-text-success': bucket.id === view?.done_bucket_id}"
											icon="check-double"
											@click.stop="toggleDoneBucket(bucket)"
										>
											{{ $t('project.kanban.doneBucket') }}
										</DropdownItem>
										<DropdownItem
											v-tooltip="$t('project.kanban.defaultBucketHint')"
											:icon-class="{'has-text-primary': bucket.id === view?.default_bucket_id}"
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
									:item-key="(task: TaskResponse) => `bucket${bucket.id}-task${task.id}`"
									:component-data="getTaskDraggableTaskComponentData(bucket)"
									@update:modelValue="(tasks) => updateTasks(bucket.id, tasks)"
									@start="handleTaskDragStart"
									@end="updateTaskPosition"
								>
									<template #footer>
										<li
											v-if="canCreateTasks"
											class="bucket-footer"
										>
											<div
												v-if="showNewTaskInput === bucket.id"
												class="field"
											>
												<div
													class="control"
													:class="{'is-loading': initialLoading || taskLoading}"
												>
													<input
														v-model="newTaskText"
														v-focus.always
														class="input"
														:disabled="initialLoading || taskLoading || undefined"
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
										</li>
									</template>

									<template #item="{element: task}">
										<li
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
												:project-id="projectId"
												@taskCompletedRecurring="handleRecurringTaskCompletion"
											/>
										</li>
									</template>
								</draggable>
							</li>
						</template>
					</draggable>

					<div
						v-if="canWrite && !initialLoading && buckets.length > 0"
						class="bucket new-bucket"
					>
						<input
							v-if="showNewBucketInput"
							v-model="newBucketTitle"
							v-focus.always
							:class="{'is-loading': initialLoading}"
							:disabled="initialLoading || undefined"
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
</template>

<script setup lang="ts">
import {useKanban} from '@/composables/useKanban'
import {bucketHasMore, kanbanKeys, type BucketResponse} from '@/client/queries/kanban'
import {removeTaskFromBoard} from '@/client/queries/taskCache'
import {useUpdateTaskPositionMutation, useMoveTaskMutation} from '@/client/queries/taskMutations'
import {
	useCreateBucketMutation,
	useDeleteBucketMutation,
	useUpdateBucketMutation,
	useLoadBucketPageMutation,
} from '@/client/queries/kanbanMutations'
import {computed, nextTick, ref, watch, toRef} from 'vue'
import {useQuery, useQueryClient} from '@tanstack/vue-query'
import {useRouter} from 'vue-router'
import {useRouteQuery} from '@vueuse/router'
import {useI18n} from 'vue-i18n'
import draggable from 'zhyswan-vuedraggable'

import {PERMISSIONS as Permissions} from '@/constants/permissions'

import {useQuickAddTask} from '@/composables/useQuickAddTask'
import {useTaskDragState} from '@/composables/useTaskDragState'
import {useAuthStore} from '@/stores/auth'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import KanbanCard from '@/components/tasks/partials/KanbanCard.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'

import {
	type CollapsedBuckets,
	getCollapsedBucketState,
	saveCollapsedBucketState,
} from '@/helpers/saveCollapsedBucketState'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'

import {getSavedFilterIdFromProjectId, isSavedFilterProject} from '@/client/queries/projects'
import {savedFilterQuery} from '@/client/queries/savedFilters'
import {useCurrentProject} from '@/composables/useCurrentProject'
import {useTaskDragToProject} from '@/composables/useTaskDragToProject'
import type {TaskFilterParams, TaskResponse} from '@/client/queries/tasks'
import type {ProjectView} from '@/client/generated'
import {createProjectViewUpdate, useUpdateProjectViewMutation} from '@/client/queries/projectViews'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	viewId: number,
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

const MIN_SCROLL_HEIGHT_PERCENT = 0.25

const {t} = useI18n({useScope: 'global'})
const isCurrentProject = ({projectId: id}: {projectId: number}) => projectId.value === id
const updateDefaultBucket = useUpdateProjectViewMutation(t('project.kanban.defaultBucketSavedSuccess'), isCurrentProject)
const updateDoneBucket = useUpdateProjectViewMutation(t('project.kanban.doneBucketSavedSuccess'), isCurrentProject)

const {createNewTask, isLoading: quickAddLoading} = useQuickAddTask()
const {setDraggedTask} = useTaskDragState()
const authStore = useAuthStore()

const alwaysShowBucketTaskCount = computed(() => authStore.settings.frontendSettings.alwaysShowBucketTaskCount)
const {handleTaskDropToProject} = useTaskDragToProject()
const positionMutation = useUpdateTaskPositionMutation()
const moveMutation = useMoveTaskMutation()

const savedFilter = useQuery(computed(() => savedFilterQuery(getSavedFilterIdFromProjectId(projectId.value)))).data

const taskContainerRefs = ref<{ [id: number]: HTMLElement }>({})
const bucketLimitInputRef = ref<HTMLInputElement | null>(null)

const drag = ref(false)
const dragBucket = ref(false)
const sourceBucket = ref(0)

const showBucketDeleteModal = ref(false)
const bucketToDelete = ref(0)
const bucketTitleEditable = ref(false)

const newTaskText = ref('')
const showNewTaskInput = ref<number | null>(null)

const newBucketTitle = ref('')
const showNewBucketInput = ref(false)
const newTaskError = ref<{ [id: number]: boolean }>({})
const newTaskInputFocused = ref(false)

const showSetLimitInput = ref(false)
const collapsedBuckets = ref<CollapsedBuckets>({})

// URL-synchronized filter parameters
const filter = useRouteQuery('filter')
const s = useRouteQuery('s')

const params = ref<TaskFilterParams>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	q: '',
})

watch([filter, s], ([filterValue, sValue]) => {
	params.value.filter = String(filterValue ?? '')
	params.value.q = String(sValue ?? '')
}, { immediate: true })

function updateFilters(newParams: TaskFilterParams) {
	// Update all params
	params.value = { ...newParams }
	
	// Sync only filter and s to URL
	filter.value = newParams.filter || undefined
	s.value = newParams.q || undefined
}

const getTaskDraggableTaskComponentData = computed(() => (bucket: BucketResponse) => {
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
const {currentProject: project} = useCurrentProject()
const view = computed(() => project.value?.views.find(view => view.id === props.viewId) as ProjectView || null)
const canWrite = computed(() =>
	typeof project.value?.max_permission === 'number' &&
	project.value.max_permission > Permissions.READ &&
	view.value?.bucket_configuration_mode === 'manual',
)
const canCreateTasks = computed(() => canWrite.value && projectId.value > 0)

const isTouchDevice = ref(false)
if (typeof window !== 'undefined') {
	isTouchDevice.value = !window.matchMedia('(hover: hover) and (pointer: fine)').matches
}
const taskDragHandle = computed(() => isTouchDevice.value ? '.handle' : undefined)

const router = useRouter()
const touchStartY = ref(0)

function openTask(task: TaskResponse) {
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

const boardParams = computed(() => ({...params.value, filter_timezone: authStore.settings.timezone}))
const board = useKanban(projectId, () => props.viewId, boardParams)
const queryClient = useQueryClient()
const buckets = board.buckets
const createBucketMutation = useCreateBucketMutation()
const deleteBucketMutation = useDeleteBucketMutation(t('project.kanban.deleteBucketSuccess'))
const updateBucketMutation = useUpdateBucketMutation()
const saveBucketTitleMutation = useUpdateBucketMutation(t('project.kanban.bucketTitleSavedSuccess'))
const saveBucketLimitMutation = useUpdateBucketMutation(t('project.kanban.bucketLimitSavedSuccess'))
const bucketPageMutation = useLoadBucketPageMutation()
const getBucketById = (id: number) => buckets.value.find(bucket => bucket.id === id)

type BucketPatch = {id: number} & Partial<Pick<BucketResponse, 'title' | 'limit' | 'position'>>

// The update is a PUT: the fields the caller does not touch come from the cached bucket.
function updateBucket(patch: BucketPatch, mutation = updateBucketMutation) {
	const cached = getBucketById(patch.id)
	if (!cached) return
	return mutation.mutateAsync({
		project: projectId.value,
		view: props.viewId,
		bucket: {title: cached.title, limit: cached.limit, position: cached.position, ...patch},
	})
}
const initialLoading = board.isLoading
const projectIdWithFallback = computed<number>(() => project.value?.id || projectId.value)

const taskLoading = computed(() => quickAddLoading.value || positionMutation.isPending.value)

watch(
	projectId,
	id => {
		if (!id) {
			return
		}
		collapsedBuckets.value = getCollapsedBucketState(id)
	},
	{immediate: true},
)

function setTaskContainerRef(id: number, el: HTMLElement) {
	if (!el) return
	taskContainerRefs.value[id] = el
}

function handleTaskContainerScroll(id: number, el: HTMLElement) {
	if (!el) {
		return
	}
	const scrollTopMax = el.scrollHeight - el.clientHeight
	const threshold = el.scrollTop + el.scrollTop * MIN_SCROLL_HEIGHT_PERCENT
	if (scrollTopMax > threshold) {
		return
	}

	// Mid-drag the copy in `buckets` is one task short, which would look like a missing page.
	const bucket = board.data.value?.buckets.find(item => item.id === id)
	if (bucketPageMutation.isPending.value || !bucket || !bucketHasMore(bucket)) return
	bucketPageMutation.mutate({
		project: projectId.value,
		view: props.viewId,
		params: boardParams.value,
		bucket: id!,
		page: (board.data.value?.pages[id!] ?? 1) + 1,
	})
}

function updateTasks(bucketId: number, tasks: BucketResponse['tasks']) {
	if (!board.dragBuckets.value) board.startDrag()
	board.dragBuckets.value = board.dragBuckets.value!.map(bucket => bucket.id === bucketId
		? {...bucket, tasks}
		: bucket)
}

async function updateTaskPosition(e) {
	const project = projectId.value
	const view = props.viewId
	drag.value = false
	try {
		const {moved} = await handleTaskDropToProject(e, task => {
			// A moved task stays in a pseudo-project board (favorites, saved filters) until the board is re-read.
			if (project < 0) {
				removeTaskFromBoard(queryClient, kanbanKeys.board(project, view, boardParams.value), task.id)
			}
		})
		if (moved) return
		const bucket = buckets.value[Number(e.to.dataset.bucketIndex)]
		const index = bucket?.tasks.findIndex(task => task.id === Number(e.item.dataset.taskId)) ?? -1
		if (!bucket || index < 0) return
		const task = bucket.tasks[index]
		const before = bucket.tasks[index - 1]
		const after = bucket.tasks[index + 1]
		if (bucket.id !== sourceBucket.value) {
			const result = await moveMutation.mutateAsync({project, view, bucket: bucket.id, task})
			if (result.bucket_id !== undefined && result.bucket_id !== bucket.id) return
		}
		await positionMutation.mutateAsync({
			taskId: task.id,
			project_view_id: view,
			position: calculateItemPosition(before?.position ?? null, after?.position ?? null),
		})

		// Dropping at the top gives position 0, which the next task may already have.
		if (index === 0 && after?.position === 0) {
			const afterAfter = bucket.tasks[index + 2]
			await positionMutation.mutateAsync({
				taskId: after.id,
				project_view_id: view,
				position: calculateItemPosition(0, afterAfter?.position ?? null),
			})
		}
	} catch { return } finally {
		board.endDrag()
	}
}

function toggleShowNewTaskInput(bucketId: number) {
	if (initialLoading.value || taskLoading.value) {
		return
	}
	showNewTaskInput.value = showNewTaskInput.value === bucketId 
		? null
		: bucketId
	newTaskInputFocused.value = false
}

async function addTaskToBucket(bucketId: number) {
	if (newTaskText.value === '') {
		newTaskError.value[bucketId] = true
		return
	}
	newTaskError.value[bucketId] = false

	await createNewTask({
		title: newTaskText.value,
		bucket_id: bucketId,
		project_id: projectIdWithFallback.value,
	})
	newTaskText.value = ''
	scrollTaskContainerToTop(bucketId)

	const bucket = getBucketById(bucketId)
	if (bucket && bucket.limit && bucket.count >= bucket.limit) {
		toggleShowNewTaskInput(bucketId)
	}
}

function scrollTaskContainerToTop(bucketId: number) {
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

	await createBucketMutation.mutateAsync({
		project: projectId.value,
		view: props.viewId,
		bucket: {title: newBucketTitle.value},
	})
	newBucketTitle.value = ''
}

function deleteBucketModal(bucketId: number) {
	if (buckets.value.length <= 1) {
		return
	}

	bucketToDelete.value = bucketId
	showBucketDeleteModal.value = true
}

async function deleteBucket() {
	try {
		await deleteBucketMutation.mutateAsync({
			project: projectId.value,
			view: props.viewId,
			bucket: bucketToDelete.value,
		})
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

async function saveBucketTitle(bucketId: number, bucketTitle: string) {
	
	const bucket = getBucketById(bucketId)
	if (bucket?.title === bucketTitle) {
		bucketTitleEditable.value = false
		return
	}
	
	await updateBucket({
		id: bucketId,
		title: bucketTitle,
	}, saveBucketTitleMutation)
	bucketTitleEditable.value = false
}

function updateBuckets(value: BucketResponse[]) {
	// (1) buckets get updated in store and tasks positions get invalidated
	board.dragBuckets.value = value
}

function handleRecurringTaskCompletion() {
	// Only reload if we're in a saved filter and the filter contains date fields
	if (!isSavedFilterProject(project.value)) {
		return
	}

	const savedFilterQueryString = savedFilter.value?.filters.filter ?? ''
	const filterContainsDateFields = savedFilterQueryString.includes('due_date') ||
		savedFilterQueryString.includes('start_date') ||
		savedFilterQueryString.includes('end_date')
		
	if (filterContainsDateFields) {
		// Reload the kanban board to refresh tasks that now match/don't match the filter
		board.refetch()
	}
}

// TODO: fix type
async function updateBucketPosition(e: { item: HTMLElement }) {
	// (2) bucket positon is changed
	dragBucket.value = false

	// Sortable reports a DOM index which can point past the last bucket, for example while a
	// deleted bucket is still leaving the transition group. The buckets are already updated here.
	const movedBucketId = parseInt(e.item.dataset.bucketId ?? '', 10)
	const bucketIndex = buckets.value.findIndex(b => b.id === movedBucketId)

	if (bucketIndex === -1) {
		board.endDrag()
		return
	}

	const bucketBefore = buckets.value[bucketIndex - 1] ?? null
	const bucketAfter = buckets.value[bucketIndex + 1] ?? null

	try {
		await updateBucket({
			id: movedBucketId,
			position: calculateItemPosition(
				bucketBefore !== null ? bucketBefore.position : null,
				bucketAfter !== null ? bucketAfter.position : null,
			),
		})
	} catch { /* Mutation reports the error. */ } finally { board.endDrag() }
}

async function saveBucketLimit(bucketId: number, limit: number) {
	if (limit < 0) {
		return
	}

	await updateBucket({
		id: bucketId,
		limit,
	}, saveBucketLimitMutation)
}

const setBucketLimitCancel = ref<number | null>(null)

async function setBucketLimit(bucketId: number, now: boolean = false) {
	const limit = parseInt(bucketLimitInputRef.value?.value || '')

	if (setBucketLimitCancel.value !== null) {
		clearTimeout(setBucketLimitCancel.value)
	}

	if (now) {
		return saveBucketLimit(bucketId, limit)
	}

	setBucketLimitCancel.value = setTimeout(saveBucketLimit, 2500, bucketId, limit)
}

function shouldAcceptDrop(bucket: BucketResponse) {
	return (
		// When dragging from a bucket who has its limit reached, dragging should still be possible
		bucket.id === sourceBucket.value ||
		// If there is no limit set, dragging & dropping should always work
		bucket.limit === 0 ||
		// Disallow dropping to buckets which have their limit reached
		bucket.count < bucket.limit
	)
}

function dragstart(bucket: BucketResponse) {
	drag.value = true
	sourceBucket.value = bucket.id
}

function handleTaskDragStart(e) {
	const taskId = parseInt(e.item.dataset.taskId, 10)
	const bucketIndex = parseInt(e.from.dataset.bucketIndex, 10)
	const bucket = buckets.value[bucketIndex]
	const task = bucket?.tasks.find(t => t.id === taskId)

	if (task) {
		board.startDrag()
		setDraggedTask(task)
	}
	dragstart(bucket)
}

function toggleDefaultBucket(bucket: BucketResponse) {
	const currentView = view.value
	if (!currentView?.id) {
		return
	}
	const defaultBucketId = currentView.default_bucket_id === bucket.id
		? 0
		: bucket.id

	updateDefaultBucket.mutate({
		projectId: projectId.value,
		viewId: currentView.id,
		view: createProjectViewUpdate({...currentView, default_bucket_id: defaultBucketId}),
	})
}

function toggleDoneBucket(bucket: BucketResponse) {
	const currentView = view.value
	if (!currentView?.id) {
		return
	}
	const doneBucketId = currentView.done_bucket_id === bucket.id
		? 0
		: bucket.id

	updateDoneBucket.mutate({
		projectId: projectId.value,
		viewId: currentView.id,
		view: createProjectViewUpdate({...currentView, done_bucket_id: doneBucketId}),
	})
}

function collapseBucket(bucket: BucketResponse) {
	collapsedBuckets.value[bucket.id] = true
	saveCollapsedBucketState(projectIdWithFallback.value, collapsedBuckets.value)
}

function unCollapseBucket(bucket: BucketResponse) {
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
		list-style: none;
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
			list-style: none;
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

			.tasks, .bucket-footer {
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
				color: var(--danger-text);
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

	.bucket-footer {
		position: sticky;
		inset-block-end: 0;
		z-index: 2;
		block-size: min-content;
		padding: .5rem;
		background-color: var(--grey-100);
		border-end-start-radius: $radius;
		border-end-end-radius: $radius;
		transform: none;
		// At fractional device pixel ratios the scroll clip ends below the sticky footer, showing a sliver of tasks
		box-shadow: 0 1px 0 var(--grey-100);

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
