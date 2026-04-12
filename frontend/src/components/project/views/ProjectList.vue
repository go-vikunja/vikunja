<template>
	<ProjectWrapper
		class="project-list"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #header>
			<div class="filter-container">
				<SortPopup
					v-model="sortByParam"
				/>
				<FilterPopup
					v-if="!isSavedFilter(project)"
					v-model="params"
					:view-id="viewId"
					:project-id="projectId"
					@update:modelValue="loadTasks()"
				/>
			</div>
		</template>

		<template #default>
			<div
				:class="{ 'is-loading': hasBuckets ? bucketStore.isLoading : loading }"
				class="loader-container is-max-width-desktop list-view"
			>
				<Card
					:padding="false"
					:has-content="false"
					class="has-overflow"
				>
					<!-- Sectioned mode (list view with manual bucket configuration) -->
					<template v-if="hasBuckets">
						<AddTask
							v-if="!project?.isArchived && canWrite"
							ref="addTaskRef"
							class="list-view__add-task d-print-none"
							:default-position="firstNewPosition"
							@taskAdded="updateTaskList"
						/>

						<div
							v-for="bucket in buckets"
							:key="bucket.id"
							class="bucket-section"
						>
							<div
								class="bucket-section__header"
								@click="toggleBucketCollapse(bucket.id)"
							>
								<span
									class="icon bucket-section__collapse-icon"
									:class="{'is-collapsed': collapsedBuckets[bucket.id]}"
								>
									<Icon icon="chevron-down" />
								</span>
								<h2
									class="bucket-section__title"
									:contenteditable="(canWrite && !collapsedBuckets[bucket.id]) ? true : undefined"
									:spellcheck="false"
									@keydown.enter.prevent.stop="!$event.isComposing && ($event.target as HTMLElement).blur()"
									@keydown.esc.prevent.stop="!$event.isComposing && ($event.target as HTMLElement).blur()"
									@blur="saveBucketTitle(bucket.id, ($event.target as HTMLElement).textContent as string)"
									@click.stop
								>
									{{ bucket.title }}
								</h2>
								<span
									v-if="bucket.limit > 0 || bucket.count > 0"
									:class="{'is-max': bucket.limit > 0 && bucket.count >= bucket.limit}"
									class="bucket-section__count"
								>
									{{ bucket.limit > 0 ? `${bucket.count}/${bucket.limit}` : bucket.count }}
								</span>
								<Dropdown
									v-if="canWrite && !collapsedBuckets[bucket.id]"
									class="is-right bucket-section__options"
									trigger-icon="ellipsis-v"
									@click.stop
								>
									<DropdownItem
										v-tooltip="$t('project.kanban.defaultBucketHint')"
										:icon-class="{'has-text-primary': bucket.id === currentView?.defaultBucketId}"
										icon="th"
										@click.stop="toggleDefaultBucket(bucket)"
									>
										{{ $t('project.kanban.defaultBucket') }}
									</DropdownItem>
									<DropdownItem
										icon="angles-up"
										@click.stop="toggleBucketCollapse(bucket.id)"
									>
										{{ $t('project.list.collapseSection') }}
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
								v-if="!collapsedBuckets[bucket.id]"
								:model-value="bucket.tasks"
								group="tasks"
								item-key="id"
								tag="ul"
								:component-data="{
									class: {
										tasks: true,
										'dragging-disabled': !canDragTasks || !isPositionSorting,
									},
									type: 'transition-group',
								}"
								:animation="100"
								:handle="dragHandle"
								:delay-on-touch-only="!isTouchDevice"
								:delay="isTouchDevice ? 0 : 1000"
								ghost-class="task-ghost"
								@start="handleDragStart"
								@end="(e) => saveTaskPositionInBucket(e, bucket)"
							>
								<template #item="{element: t}">
									<SingleTaskInProject
										:show-list-color="false"
										:can-mark-as-done="canWrite || isPseudoProject"
										:the-task="t"
										:all-tasks="getAllTasksFromBuckets()"
										@taskUpdated="updateTaskInBuckets"
									>
										<span
											v-if="canDragTasks && isPositionSorting"
											class="icon handle"
										>
											<Icon icon="grip-lines" />
										</span>
									</SingleTaskInProject>
								</template>
							</draggable>

							<!-- Load more -->
							<div
								v-if="!collapsedBuckets[bucket.id] && bucket.tasks.length < bucket.count"
								class="bucket-section__load-more"
							>
								<ButtonLink @click="loadMoreForBucket(bucket.id)">
									{{ $t('project.list.loadMore', {remaining: bucket.count - bucket.tasks.length}) }}
								</ButtonLink>
							</div>
						</div>

						<!-- Add section button -->
						<div
							v-if="canWrite"
							class="bucket-section__add"
						>
							<div
								v-if="showNewSectionInput"
								class="field has-addons bucket-section__add-input"
							>
								<div class="control is-expanded">
									<input
										v-model="newSectionTitle"
										v-focus.always
										class="input"
										:placeholder="$t('project.list.addSectionPlaceholder')"
										type="text"
										@keyup.enter="createNewSection"
										@keyup.esc="showNewSectionInput = false"
									>
								</div>
								<div class="control">
									<XButton
										:shadow="false"
										@click="createNewSection"
									>
										{{ $t('project.list.add') }}
									</XButton>
								</div>
							</div>
							<ButtonLink
								v-else
								@click="showNewSectionInput = true"
							>
								<Icon icon="plus" />
								{{ $t('project.list.addSection') }}
							</ButtonLink>
						</div>

						<!-- Delete section confirmation modal -->
						<Modal
							v-if="showSectionDeleteModal"
							@close="showSectionDeleteModal = false"
							@submit="deleteSection"
						>
							<template #header>
								{{ $t('project.list.deleteSection') }}
							</template>
							<template #text>
								<p>{{ $t('project.list.deleteSectionText1') }}</p>
								<p>{{ $t('project.list.deleteSectionText2') }}</p>
							</template>
						</Modal>
					</template>

					<!-- Flat mode (existing) -->
					<template v-else>
						<AddTask
							v-if="!project?.isArchived && canWrite"
							ref="addTaskRef"
							class="list-view__add-task d-print-none"
							:default-position="firstNewPosition"
							@taskAdded="updateTaskList"
						/>

						<Nothing v-if="ctaVisible && tasks.length === 0 && !loading">
							{{ $t('project.list.empty') }}
							<ButtonLink
								v-if="project?.id > 0 && canWrite"
								@click="focusNewTaskInput()"
							>
								{{ $t('project.list.newTaskCta') }}
							</ButtonLink>
						</Nothing>

						<draggable
							v-if="tasks && tasks.length > 0"
							v-model="tasks"
							:group="{name: 'tasks', put: false}"
							:disabled="!canDragTasks || !isPositionSorting"
							item-key="id"
							tag="ul"
							:component-data="{
								class: {
									tasks: true,
									'dragging-disabled': !canDragTasks || !isPositionSorting
								},
								type: 'transition-group'
							}"
							:animation="100"
							:handle="dragHandle"
							:delay-on-touch-only="!isTouchDevice"
							:delay="isTouchDevice ? 0 : 1000"
							ghost-class="task-ghost"
							@start="handleDragStart"
							@end="saveTaskPosition"
						>
							<template #item="{element: t, index}">
								<SingleTaskInProject
									:ref="(el) => setTaskRef(el, index)"
									:show-list-color="false"
									:can-mark-as-done="canWrite || isPseudoProject"
									:the-task="t"
									:all-tasks="allTasks"
									@taskUpdated="updateTasks"
								>
									<span
										v-if="canDragTasks && isPositionSorting"
										class="icon handle"
									>
										<Icon icon="grip-lines" />
									</span>
								</SingleTaskInProject>
							</template>
						</draggable>

						<Pagination
							:total-pages="totalPages"
							:current-page="currentPage"
						/>
					</template>
				</Card>
			</div>
		</template>
	</ProjectWrapper>
</template>


<script setup lang="ts">
import {ref, computed, nextTick, onMounted, onBeforeUnmount, watch, toRef} from 'vue'
import {useI18n} from 'vue-i18n'
import draggable from 'zhyswan-vuedraggable'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import ButtonLink from '@/components/misc/ButtonLink.vue'
import AddTask from '@/components/tasks/AddTask.vue'
import SingleTaskInProject from '@/components/tasks/partials/SingleTaskInProject.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import Nothing from '@/components/misc/Nothing.vue'
import Pagination from '@/components/misc/Pagination.vue'
import SortPopup from '@/components/project/partials/SortPopup.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import Modal from '@/components/misc/Modal.vue'

import {useTaskList} from '@/composables/useTaskList'
import {useTaskDragToProject} from '@/composables/useTaskDragToProject'
import {shouldShowTaskInListView} from '@/composables/useTaskListFiltering'
import {PERMISSIONS as Permissions} from '@/constants/permissions'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import {
	type CollapsedBuckets,
	getCollapsedBucketState,
	saveCollapsedBucketState,
} from '@/helpers/saveCollapsedBucketState'
import type {ITask} from '@/modelTypes/ITask'
import type {IBucket} from '@/modelTypes/IBucket'
import {isSavedFilter, useSavedFilter} from '@/services/savedFilter'
import {success} from '@/message'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'
import {useBucketStore} from '@/stores/buckets'
import {useProjectStore} from '@/stores/projects'

import type {IProject} from '@/modelTypes/IProject'
import type {IProjectView} from '@/modelTypes/IProjectView'
import TaskPositionService from '@/services/taskPosition'
import TaskPositionModel from '@/models/taskPosition'
import TaskBucketService from '@/services/taskBucket'
import TaskBucketModel from '@/models/taskBucket'
import BucketModel from '@/models/bucket'
import ProjectViewService from '@/services/projectViews'
import ProjectViewModel from '@/models/projectView'

const props = defineProps<{
        isLoadingProject: boolean,
        projectId: IProject['id'],
        viewId: IProjectView['id'],
}>()

const projectId = toRef(props, 'projectId')

defineOptions({name: 'List'})

const {t} = useI18n({useScope: 'global'})

const ctaVisible = ref(false)

const drag = ref(false)

const {
	tasks: allTasks,
	loading,
	totalPages,
	currentPage,
	loadTasks,
	params,
	sortByParam,
} = useTaskList(
	() => projectId.value,
	() => props.viewId,
	{position: 'asc'},
	() => projectId.value === -1
		? ['comment_count', 'is_unread']
		: ['subtasks', 'comment_count', 'is_unread'],
)

const taskPositionService = ref(new TaskPositionService())

// Saved filter composable for accessing filter data
const _savedFilter = useSavedFilter(() => isSavedFilter({id: projectId.value}) ? projectId.value : undefined).filter

const tasks = ref<ITask[]>([])
watch(
	allTasks,
	() => {
		const isFiltered = isSavedFilter({id: projectId.value})
		tasks.value = ([...allTasks.value]).filter(t => shouldShowTaskInListView(t, allTasks.value, isFiltered))
	},
)

const isPositionSorting = computed(() => 'position' in sortByParam.value)

const firstNewPosition = computed(() => {
	if (hasBuckets.value) {
		const defaultBucket = buckets.value.find(b => b.id === currentView.value?.defaultBucketId) || buckets.value[0]
		if (defaultBucket?.tasks?.length > 0) {
			return calculateItemPosition(null, defaultBucket.tasks[0].position)
		}
		return 0
	}

	if (tasks.value.length === 0) {
		return 0
	}

	return calculateItemPosition(null, tasks.value[0].position)
})

const baseStore = useBaseStore()
const taskStore = useTaskStore()
const bucketStore = useBucketStore()
const projectStore = useProjectStore()
const {handleTaskDropToProject} = useTaskDragToProject()
const project = computed(() => baseStore.currentProject)

const canWrite = computed(() => {
	return project.value?.maxPermission > Permissions.READ && project.value?.id > 0
})

const isPseudoProject = computed(() => (project.value && isSavedFilter(project.value)) || project.value?.id === -1)

onMounted(async () => {
	await nextTick()
	ctaVisible.value = true
})

const canDragTasks = computed(() => canWrite.value || isSavedFilter(project.value))

const isTouchDevice = ref(false)
if (typeof window !== 'undefined') {
	isTouchDevice.value = !window.matchMedia('(hover: hover) and (pointer: fine)').matches
}
const dragHandle = computed(() => isTouchDevice.value ? '.handle' : undefined)

const addTaskRef = ref<typeof AddTask | null>(null)

function focusNewTaskInput() {
	addTaskRef.value?.focusTaskInput()
}

// ==========================================
// Bucket/Section mode
// ==========================================

const currentView = computed(() => {
	return project.value?.views?.find(v => v.id === props.viewId) as IProjectView || null
})

const hasBuckets = computed(() => {
	return currentView.value?.bucketConfigurationMode !== 'none'
		&& currentView.value?.bucketConfigurationMode !== undefined
})

const buckets = computed(() => bucketStore.buckets)

const collapsedBuckets = ref<CollapsedBuckets>({})

watch(
	[() => props.projectId, () => props.viewId, hasBuckets],
	async ([pId, vId, bucketed]) => {
		if (!bucketed) return
		await bucketStore.loadBucketsForProject(pId, vId, params.value)
		collapsedBuckets.value = getCollapsedBucketState(pId)
	},
	{immediate: true},
)

function toggleBucketCollapse(bucketId: IBucket['id']) {
	collapsedBuckets.value = {
		...collapsedBuckets.value,
		[bucketId]: !collapsedBuckets.value[bucketId],
	}
	saveCollapsedBucketState(props.projectId, collapsedBuckets.value)
}

function getAllTasksFromBuckets(): ITask[] {
	return buckets.value.flatMap(b => b.tasks)
}

function updateTaskInBuckets(updatedTask: ITask) {
	bucketStore.setTaskInBucket(updatedTask)
}

async function saveBucketTitle(bucketId: IBucket['id'], bucketTitle: string) {
	const bucket = bucketStore.getBucketById(bucketId)
	if (bucket?.title === bucketTitle) {
		return
	}

	await bucketStore.updateBucket({
		id: bucketId,
		title: bucketTitle,
		projectId: projectId.value,
	})
	success({message: t('project.list.sectionTitleSavedSuccess')})
}

const sectionToDelete = ref<IBucket['id']>(0)
const showSectionDeleteModal = ref(false)

function deleteBucketModal(bucketId: IBucket['id']) {
	if (buckets.value.length <= 1) {
		return
	}

	sectionToDelete.value = bucketId
	showSectionDeleteModal.value = true
}

async function deleteSection() {
	try {
		await bucketStore.deleteBucket({
			bucket: new BucketModel({
				id: sectionToDelete.value,
				projectId: projectId.value,
				projectViewId: props.viewId,
			}),
			params: params.value,
		})
		success({message: t('project.list.deleteSectionSuccess')})
	} finally {
		showSectionDeleteModal.value = false
	}
}

const newSectionTitle = ref('')
const showNewSectionInput = ref(false)

async function createNewSection() {
	if (newSectionTitle.value === '') {
		return
	}

	await bucketStore.createBucket(new BucketModel({
		title: newSectionTitle.value,
		projectId: projectId.value,
		projectViewId: props.viewId,
	}))
	newSectionTitle.value = ''
	showNewSectionInput.value = false
}

async function toggleDefaultBucket(bucket: IBucket) {
	const defaultBucketId = currentView.value?.defaultBucketId === bucket.id
		? 0
		: bucket.id

	const projectViewService = new ProjectViewService()
	const updatedView = await projectViewService.update(new ProjectViewModel({
		...currentView.value,
		defaultBucketId,
	}))

	const views = project.value.views.map(v => v.id === currentView.value?.id ? updatedView : v)
	const updatedProject = {
		...project.value,
		views,
	}

	projectStore.setProject(updatedProject)

	success({message: t('project.kanban.defaultBucketSavedSuccess')})
}

async function loadMoreForBucket(bucketId: IBucket['id']) {
	await bucketStore.loadNextTasksForBucket(
		props.projectId,
		props.viewId,
		params.value,
		bucketId,
	)
}

async function saveTaskPositionInBucket(
	e: {originalEvent?: MouseEvent, to: HTMLElement, from: HTMLElement, newIndex: number, oldIndex: number},
	targetBucket: IBucket,
) {
	drag.value = false

	const task = targetBucket.tasks[e.newIndex]
	if (!task) return

	const taskBefore = targetBucket.tasks[e.newIndex - 1] ?? null
	const taskAfter = targetBucket.tasks[e.newIndex + 1] ?? null

	const position = calculateItemPosition(
		taskBefore?.position ?? null,
		taskAfter?.position ?? null,
	)

	await taskPositionService.value.update(new TaskPositionModel({
		position,
		projectViewId: props.viewId,
		taskId: task.id,
	}))

	// If bucket changed, update bucket assignment
	if (e.to !== e.from) {
		const taskBucketService = new TaskBucketService()
		await taskBucketService.update(new TaskBucketModel({
			taskId: task.id,
			bucketId: targetBucket.id,
			projectViewId: props.viewId,
		}))
	}
}

// ==========================================
// Flat mode (existing logic)
// ==========================================

function updateTaskList(task: ITask) {
	if (hasBuckets.value) {
		bucketStore.addTaskToBucket(task)
		baseStore.setHasTasks(true)
		return
	}

	if (!isPositionSorting.value) {
		// reload tasks with current filter and sorting
		loadTasks()
	} else {
		allTasks.value = [
			task,
			...allTasks.value,
		]
	}

	baseStore.setHasTasks(true)
}

function updateTasks(updatedTask: ITask) {
	if (projectId.value < 0) {
		// Reload tasks to keep saved filter results in sync
		loadTasks(false)
		return
	}

	for (const t in tasks.value) {
		if (tasks.value[t].id === updatedTask.id) {
			tasks.value[t] = updatedTask
			break
		}
	}
}

function handleDragStart(e: { item: HTMLElement }) {
	drag.value = true
	const taskId = parseInt(e.item.dataset.taskId ?? '', 10)
	const allAvailableTasks = hasBuckets.value ? getAllTasksFromBuckets() : tasks.value
	const task = allAvailableTasks.find(t => t.id === taskId)

	if (task) {
		taskStore.setDraggedTask(task)
	}
}

async function saveTaskPosition(e: { originalEvent?: MouseEvent, to: HTMLElement, from: HTMLElement, newIndex: number }) {
	drag.value = false

	// Check if dropped on a sidebar project
	const {moved} = await handleTaskDropToProject(e, (task) => {
		tasks.value = tasks.value.filter(t => t.id !== task.id)
	})

	if (moved) {
		return
	}

	// If dropped outside this list
	if (e.to !== e.from) {
		return
	}

	const task = tasks.value[e.newIndex]
	const taskBefore = tasks.value[e.newIndex - 1] ?? null
	const taskAfter = tasks.value[e.newIndex + 1] ?? null

	const position = calculateItemPosition(taskBefore !== null ? taskBefore.position : null, taskAfter !== null ? taskAfter.position : null)

	await taskPositionService.value.update(new TaskPositionModel({
		position,
		projectViewId: props.viewId,
		taskId: task.id,
	}))
	tasks.value[e.newIndex] = {
		...task,
		position,
	}
}

const taskRefs = ref<(InstanceType<typeof SingleTaskInProject> | null)[]>([])
const focusedIndex = ref(-1)

function setTaskRef(el: InstanceType<typeof SingleTaskInProject> | null, index: number) {
	if (el === null) {
		delete taskRefs.value[index]
	} else {
		taskRefs.value[index] = el
	}
}

function focusTask(index: number) {
	if (index < 0 || index >= tasks.value.length) {
		return
	}

	const taskRef = taskRefs.value[index]

	focusedIndex.value = index
	taskRef?.focus()
}

function handleListNavigation(e: KeyboardEvent) {
	if (e.target instanceof HTMLElement && (e.target.closest('input, textarea, select, [contenteditable="true"]'))) {
		return
	}

	if (e.code === 'KeyJ') {
		e.preventDefault()
		focusTask(Math.min(focusedIndex.value + 1, tasks.value.length - 1))
		return
	}

	if (e.code === 'KeyK') {
		e.preventDefault()
		if (focusedIndex.value === -1) {
			focusTask(tasks.value.length - 1)
			return
		}

		if (focusedIndex.value === 0) {
			addTaskRef.value?.focusTaskInput()
			focusedIndex.value = -1
			return
		}

		focusTask(Math.max(focusedIndex.value - 1, 0))
		return
	}

	if (e.code === 'Enter') {
		if (e.isComposing) {
			return
		}
		e.preventDefault()
		taskRefs.value[focusedIndex.value]?.click(e)
	}
}

onMounted(() => {
	document.addEventListener('keydown', handleListNavigation)
})

onBeforeUnmount(() => {
	document.removeEventListener('keydown', handleListNavigation)
})
</script>

<style lang="scss" scoped>
.filter-container {
	display: flex;
	align-items: center;
	gap: .5rem;

	:deep(.popup) {
		inset-block-start: 3rem;
		inset-inline-end: 0;
		max-inline-size: 300px;
	}
}

.tasks {
	padding: .5rem;
}

.task-ghost {
	border-radius: $radius;
	background: var(--grey-100);
	border: 2px dashed var(--grey-300);

	* {
		opacity: 0;
	}
}

.list-view__add-task {
	padding: 1rem 1rem 0;
}

.link-share-view .card {
	border: none;
	box-shadow: none;
}

:deep(.single-task .handle) {
	cursor: grab;
	margin-inline-end: .25rem;
	color: var(--grey-400);
}

@media (hover: hover) and (pointer: fine) {
	:deep(.single-task .handle) {
		display: none;
	}
}

:deep(.tasks:not(.dragging-disabled) .single-task) {
	cursor: grab;
	-webkit-touch-callout: none;
	user-select: none;
	touch-action: manipulation;

	&:active {
		cursor: grabbing;
	}
}

.list-view {
	padding-block-end: 1rem;

	:deep(.card) {
		margin-block-end: 0;
	}
}

// Bucket section styles
.bucket-section {
	&:not(:first-child) {
		border-block-start: 1px solid var(--grey-200);
	}
}

.bucket-section__header {
	display: flex;
	align-items: center;
	gap: .5rem;
	padding: .75rem 1rem;
	cursor: pointer;
	user-select: none;
	background: var(--grey-50);
	border-block-end: 1px solid var(--grey-100);

	&:hover {
		background: var(--grey-100);
	}
}

.bucket-section__collapse-icon {
	transition: transform 150ms ease;
	color: var(--grey-500);

	&.is-collapsed {
		transform: rotate(-90deg);
	}
}

.bucket-section__title {
	flex: 1;
	font-size: 1rem;
	font-weight: 600;
	margin: 0;
	padding: .125rem .25rem;
	border-radius: $radius;
	min-inline-size: 0;

	&[contenteditable='true'] {
		cursor: text;

		&:focus {
			outline: 2px solid var(--primary);
			outline-offset: 1px;
		}
	}
}

.bucket-section__count {
	font-size: .85rem;
	color: var(--grey-500);
	font-weight: 500;

	&.is-max {
		color: var(--danger);
	}
}

.bucket-section__options {
	margin-inline-start: auto;
}

.bucket-section__load-more {
	padding: .5rem 1rem;
	text-align: center;
	color: var(--grey-500);
}

.bucket-section__add {
	padding: 1rem;
	text-align: center;
}

.bucket-section__add-input {
	max-inline-size: 400px;
	margin-inline: auto;
}
</style>
