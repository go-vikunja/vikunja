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
									>
										<Icon icon="check-double" />
									</span>
									<h2
										class="title input"
										:contenteditable="(bucketTitleEditable && canWrite && !collapsedBuckets[bucket.id]) ? true : undefined"
										:spellcheck="false"
										@keydown.enter.prevent.stop="($event.target as HTMLElement).blur()"
										@keydown.esc.prevent.stop="($event.target as HTMLElement).blur()"
										@blur="saveBucketTitle(bucket.id, ($event.target as HTMLElement).textContent as string)"
										@click="focusBucketTitle"
									>
										{{ bucket.title }}
									</h2>
									<span
										v-if="bucket.limit > 0"
										:class="{'is-max': bucket.count >= bucket.limit}"
										class="limit"
									>
										{{ bucket.count }}/{{ bucket.limit }}
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

								<draggable
									v-bind="DRAG_OPTIONS"
									:model-value="bucket.tasks"
									:group="{name: 'tasks', put: shouldAcceptDrop(bucket) && !dragBucket}"
									:disabled="!canWrite"
									:data-bucket-index="bucketIndex"
									tag="ul"
									:item-key="(task: ITask) => `bucket${bucket.id}-task${task.id}`"
									:component-data="getTaskDraggableTaskComponentData(bucket)"
									@update:modelValue="(tasks) => updateTasks(bucket.id, tasks)"
									@start="() => dragstart(bucket)"
									@end="updateTaskPosition"
								>
									<template #footer>
										<div
											v-if="canCreateTasks"
											class="bucket-footer"
										>
											<div
												v-if="showNewTaskInput[bucket.id]"
												class="field"
											>
												<div
													class="control"
													:class="{'is-loading': loading || taskLoading}"
												>
													<input
														v-model="newTaskText"
														v-focus.always
														class="input"
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
									</template>

									<template #item="{element: task}">
										<div class="task-item">
											<KanbanCard
												class="kanban-card"
												:task="task"
												:loading="taskUpdating[task.id] ?? false"
												:project-id="projectId"
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
</template>

<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import draggable from 'zhyswan-vuedraggable'
import {klona} from 'klona/lite'

import {PERMISSIONS as Permissions} from '@/constants/permissions'
import BucketModel from '@/models/bucket'

import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'

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

import {isSavedFilter} from '@/services/savedFilter'
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

const DRAG_OPTIONS = {
	// sortable options
	animation: 150,
	ghostClass: 'ghost',
	dragClass: 'task-dragging',
	delayOnTouchOnly: true,
	delay: 150,
} as const

const MIN_SCROLL_HEIGHT_PERCENT = 0.25

const {t} = useI18n({useScope: 'global'})

const baseStore = useBaseStore()
const kanbanStore = useKanbanStore()
const taskStore = useTaskStore()
const projectStore = useProjectStore()
const taskPositionService = ref(new TaskPositionService())
const taskBucketService = ref(new TaskBucketService())

const taskContainerRefs = ref<{ [id: IBucket['id']]: HTMLElement }>({})
const bucketLimitInputRef = ref<HTMLInputElement | null>(null)

const drag = ref(false)
const dragBucket = ref(false)
const sourceBucket = ref(0)

const showBucketDeleteModal = ref(false)
const bucketToDelete = ref(0)
const bucketTitleEditable = ref(false)

const newTaskText = ref('')
const showNewTaskInput = ref<{ [id: IBucket['id']]: boolean }>({})

const newBucketTitle = ref('')
const showNewBucketInput = ref(false)
const newTaskError = ref<{ [id: IBucket['id']]: boolean }>({})
const newTaskInputFocused = ref(false)

const showSetLimitInput = ref(false)
const collapsedBuckets = ref<CollapsedBuckets>({})

// We're using this to show the loading animation only at the task when updating it
const taskUpdating = ref<{ [id: ITask['id']]: boolean }>({})
const oneTaskUpdating = ref(false)

const params = ref<TaskFilterParams>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	s: '',
})

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
const project = computed(() => props.projectId ? projectStore.projects[props.projectId] : null)
const view = computed(() => project.value?.views.find(v => v.id === props.viewId) as IProjectView || null)
const canWrite = computed(() => baseStore.currentProject?.maxPermission > Permissions.READ && view.value.bucketConfigurationMode === 'manual')
const canCreateTasks = computed(() => canWrite.value && props.projectId > 0)
const buckets = computed(() => kanbanStore.buckets)
const loading = computed(() => kanbanStore.isLoading)

const taskLoading = computed(() => taskStore.isLoading || taskPositionService.value.loading)

watch(
	() => ({
		params: params.value,
		projectId: props.projectId,
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
		props.projectId,
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
	const oldBucket = buckets.value.find(b => b.id === task.bucketId)
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
				projectId: project.value.id,
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
	showNewTaskInput.value[bucketId] = !showNewTaskInput.value[bucketId]
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
		projectId: project.value.id,
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
		projectId: project.value.id,
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
				projectId: project.value.id,
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
		projectId: props.projectId,
	})
	success({message: i18n.global.t('project.kanban.bucketTitleSavedSuccess')})
	bucketTitleEditable.value = false
}

function updateBuckets(value: IBucket[]) {
	// (1) buckets get updated in store and tasks positions get invalidated
	kanbanStore.setBuckets(value)
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
		projectId: props.projectId,
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
		projectId: props.projectId,
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
	saveCollapsedBucketState(project.value.id, collapsedBuckets.value)
}

function unCollapseBucket(bucket: IBucket) {
	if (!collapsedBuckets.value[bucket.id]) {
		return
	}

	collapsedBuckets.value[bucket.id] = false
	saveCollapsedBucketState(project.value.id, collapsedBuckets.value)
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

			&:first-of-type {
				padding-block-start: .5rem;
			}

			&:last-of-type {
				padding-block-end: .5rem;
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

	.bucket-footer {
		position: sticky;
		inset-block-end: 0;
		block-size: min-content;
		padding: .5rem;
		background-color: var(--grey-100);
		border-end-start-radius: $radius;
		border-end-end-radius: $radius;
		transform: none;

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
