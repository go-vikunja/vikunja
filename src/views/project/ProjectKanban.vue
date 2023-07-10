<template>
	<ProjectWrapper
		class="project-kanban"
		:project-id="project.id"
		viewName="kanban"
	>
		<template #header>
			<div class="filter-container" v-if="!isSavedFilter(project)">
				<div class="items">
					<filter-popup v-model="params" />
				</div>
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
				:modelValue="buckets"
				@update:modelValue="updateBuckets"
				@end="updateBucketPosition"
				@start="() => dragBucket = true"
				group="buckets"
				:disabled="!canWrite || newTaskInputFocused"
				tag="ul"
				:item-key="({id}: IBucket) => `bucket${id}`"
				:component-data="bucketDraggableComponentData"
			>
				<template #item="{element: bucket, index: bucketIndex }">
					<div
						class="bucket"
						:class="{'is-collapsed': collapsedBuckets[bucket.id]}"
					>
						<div class="bucket-header" @click="() => unCollapseBucket(bucket)">
							<span
								v-if="bucket.isDoneBucket"
								class="icon is-small has-text-success mr-2"
								v-tooltip="$t('project.kanban.doneBucketHint')"
							>
								<icon icon="check-double"/>
							</span>
							<h2
								@keydown.enter.prevent.stop="($event.target as HTMLElement).blur()"
								@keydown.esc.prevent.stop="($event.target as HTMLElement).blur()"
								@blur="saveBucketTitle(bucket.id, ($event.target as HTMLElement).textContent as string)"
								@click="focusBucketTitle"
								class="title input"
								:contenteditable="(bucketTitleEditable && canWrite && !collapsedBuckets[bucket.id]) ? true : undefined"
								:spellcheck="false">{{ bucket.title }}</h2>
							<span
								:class="{'is-max': bucket.count >= bucket.limit}"
								class="limit"
								v-if="bucket.limit > 0">
								{{ bucket.count }}/{{ bucket.limit }}
							</span>
							<dropdown
								class="is-right options"
								v-if="canWrite && !collapsedBuckets[bucket.id]"
								trigger-icon="ellipsis-v"
								@close="() => showSetLimitInput = false"
							>
								<dropdown-item
									@click.stop="showSetLimitInput = true"
								>
									<div class="field has-addons" v-if="showSetLimitInput">
										<div class="control">
											<input
												@keyup.esc="() => showSetLimitInput = false"
												@keyup.enter="() => showSetLimitInput = false"
												:value="bucket.limit"
												@input="(event) => setBucketLimit(bucket.id, parseInt((event.target as HTMLInputElement).value))"
												class="input"
												type="number"
												min="0"
												v-focus.always
											/>
										</div>
										<div class="control">
											<x-button
												:disabled="bucket.limit < 0"
												:icon="['far', 'save']"
												:shadow="false"
												v-cy="'setBucketLimit'"
											/>
										</div>
									</div>
									<template v-else>
										{{
											$t('project.kanban.limit', {limit: bucket.limit > 0 ? bucket.limit : $t('project.kanban.noLimit')})
										}}
									</template>
								</dropdown-item>
								<dropdown-item
									@click.stop="toggleDoneBucket(bucket)"
									v-tooltip="$t('project.kanban.doneBucketHintExtended')"
								>
									<span class="icon is-small" :class="{'has-text-success': bucket.isDoneBucket}">
										<icon icon="check-double"/>
									</span>
									{{ $t('project.kanban.doneBucket') }}
								</dropdown-item>
								<dropdown-item
									@click.stop="() => collapseBucket(bucket)"
								>
									{{ $t('project.kanban.collapse') }}
								</dropdown-item>
								<dropdown-item
									:class="{'is-disabled': buckets.length <= 1}"
									@click.stop="() => deleteBucketModal(bucket.id)"
									class="has-text-danger"
									v-tooltip="buckets.length <= 1 ? $t('project.kanban.deleteLast') : ''"
								>
									<span class="icon is-small">
										<icon icon="trash-alt"/>
									</span>
									{{ $t('misc.delete') }}
								</dropdown-item>
							</dropdown>
						</div>

						<draggable
							v-bind="DRAG_OPTIONS"
							:modelValue="bucket.tasks"
							@update:modelValue="(tasks) => updateTasks(bucket.id, tasks)"
							@start="() => dragstart(bucket)"
							@end="updateTaskPosition"
							:group="{name: 'tasks', put: shouldAcceptDrop(bucket) && !dragBucket}"
							:disabled="!canWrite"
							:data-bucket-index="bucketIndex"
							tag="ul"
							:item-key="(task: ITask) => `bucket${bucket.id}-task${task.id}`"
							:component-data="getTaskDraggableTaskComponentData(bucket)"
						>
							<template #footer>
								<div class="bucket-footer" v-if="canWrite">
									<div class="field" v-if="showNewTaskInput[bucket.id]">
										<div class="control" :class="{'is-loading': loading || taskLoading}">
											<input
												class="input"
												:disabled="loading || taskLoading || undefined"
												@focusout="toggleShowNewTaskInput(bucket.id)"
												@focusin="() => newTaskInputFocused = true"
												@keyup.enter="addTaskToBucket(bucket.id)"
												@keyup.esc="toggleShowNewTaskInput(bucket.id)"
												:placeholder="$t('project.kanban.addTaskPlaceholder')"
												type="text"
												v-focus.always
												v-model="newTaskText"
											/>
										</div>
										<p class="help is-danger" v-if="newTaskError[bucket.id] && newTaskText === ''">
											{{ $t('project.create.addTitleRequired') }}
										</p>
									</div>
									<x-button
										@click="toggleShowNewTaskInput(bucket.id)"
										class="is-fullwidth has-text-centered"
										:shadow="false"
										v-else
										icon="plus"
										variant="secondary"
									>
										{{ bucket.tasks.length === 0 ? $t('project.kanban.addTask') : $t('project.kanban.addAnotherTask') }}
									</x-button>
								</div>
							</template>

							<template #item="{element: task}">
								<div class="task-item">
									<kanban-card class="kanban-card" :task="task" :loading="taskUpdating[task.id] ?? false"/>
								</div>
							</template>
						</draggable>
					</div>
				</template>
			</draggable>

			<div class="bucket new-bucket" v-if="canWrite && !loading && buckets.length > 0">
				<input
					:class="{'is-loading': loading}"
					:disabled="loading || undefined"
					@blur="() => showNewBucketInput = false"
					@keyup.enter="createNewBucket"
					@keyup.esc="($event.target as HTMLInputElement).blur()"
					class="input"
					:placeholder="$t('project.kanban.addBucketPlaceholder')"
					type="text"
					v-focus.always
					v-if="showNewBucketInput"
					v-model="newBucketTitle"
				/>
				<x-button
					v-else
					@click="() => showNewBucketInput = true"
					:shadow="false"
					class="is-transparent is-fullwidth has-text-centered"
					variant="secondary"
					icon="plus"
				>
					{{ $t('project.kanban.addBucket') }}
				</x-button>
			</div>
		</div>

		<modal
			:enabled="showBucketDeleteModal"
			@close="showBucketDeleteModal = false"
			@submit="deleteBucket()"
		>
			<template #header><span>{{ $t('project.kanban.deleteHeaderBucket') }}</span></template>

			<template #text>
				<p>{{ $t('project.kanban.deleteBucketText1') }}<br/>
					{{ $t('project.kanban.deleteBucketText2') }}</p>
			</template>
		</modal>
		</div>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, nextTick, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import draggable from 'zhyswan-vuedraggable'
import {klona} from 'klona/lite'

import {RIGHTS as Rights} from '@/constants/rights'
import BucketModel from '@/models/bucket'

import type {IBucket} from '@/modelTypes/IBucket'
import type {IProject} from '@/modelTypes/IProject'
import type {ITask} from '@/modelTypes/ITask'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import FilterPopup from '@/components/project/partials/filter-popup.vue'
import KanbanCard from '@/components/tasks/partials/kanban-card.vue'
import Dropdown from '@/components/misc/dropdown.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'

import {getCollapsedBucketState, saveCollapsedBucketState, type CollapsedBuckets} from '@/helpers/saveCollapsedBucketState'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'

import {isSavedFilter} from '@/services/savedFilter'
import {success} from '@/message'

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

const taskContainerRefs = ref<{[id: IBucket['id']]: HTMLElement}>({})

const drag = ref(false)
const dragBucket = ref(false)
const sourceBucket = ref(0)

const showBucketDeleteModal = ref(false)
const bucketToDelete = ref(0)
const bucketTitleEditable = ref(false)

const newTaskText = ref('')
const showNewTaskInput = ref<{[id: IBucket['id']]: boolean}>({})

const newBucketTitle = ref('')
const showNewBucketInput = ref(false)
const newTaskError = ref<{[id: IBucket['id']]: boolean}>({})
const newTaskInputFocused = ref(false)

const showSetLimitInput = ref(false)
const collapsedBuckets = ref<CollapsedBuckets>({})

// We're using this to show the loading animation only at the task when updating it
const taskUpdating = ref<{[id: ITask['id']]: boolean}>({})
const oneTaskUpdating = ref(false)

const params = ref({
	filter_by: [],
	filter_value: [],
	filter_comparator: [],
	filter_concat: 'and',
})

const getTaskDraggableTaskComponentData = computed(() => (bucket: IBucket) => {
	return {
		ref: (el: HTMLElement) => setTaskContainerRef(bucket.id, el),
		onScroll: (event: Event) => handleTaskContainerScroll(bucket.id, bucket.projectId, event.target as HTMLElement),
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

const canWrite = computed(() => baseStore.currentProject?.maxRight > Rights.READ)
const project = computed(() => baseStore.currentProject)

const buckets = computed(() => kanbanStore.buckets)
const loading = computed(() => kanbanStore.isLoading)

const taskLoading = computed(() => taskStore.isLoading)

watch(
	() => ({
		params: params.value,
		project: project.value,
	}),
	({params, project}) => {
		const projectId = project.id
		if (projectId === undefined) {
			return
		}
		collapsedBuckets.value = getCollapsedBucketState(projectId)
		kanbanStore.loadBucketsForProject({projectId, params})
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

function handleTaskContainerScroll(id: IBucket['id'], projectId: IProject['id'], el: HTMLElement) {
	if (!el) {
		return
	}
	const scrollTopMax = el.scrollHeight - el.clientHeight
	const threshold = el.scrollTop + el.scrollTop * MIN_SCROLL_HEIGHT_PERCENT
	if (scrollTopMax > threshold) {
		return
	}

	kanbanStore.loadNextTasksForBucket({
		projectId: projectId,
		params: params.value,
		bucketId: id,
	})
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
	newTask.kanbanPosition = calculateItemPosition(
		taskBefore !== null ? taskBefore.kanbanPosition : null,
		taskAfter !== null ? taskAfter.kanbanPosition : null,
	)
	if (
		oldBucket !== undefined && // This shouldn't actually be `undefined`, but let's play it safe.
		newBucket.id !== oldBucket.id &&
		newBucket.isDoneBucket !== oldBucket.isDoneBucket
	) {
		newTask.done = newBucket.isDoneBucket
	}
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
	}

	try {
		await taskStore.update(newTask)
		
		// Make sure the first and second task don't both get position 0 assigned
		if(newTaskIndex === 0 && taskAfter !== null && taskAfter.kanbanPosition === 0) {
			const taskAfterAfter = newBucket.tasks[newTaskIndex + 2] ?? null
			const newTaskAfter = klona(taskAfter) // cloning the task to avoid pinia store manipulation
			newTaskAfter.bucketId = newBucket.id
			newTaskAfter.kanbanPosition = calculateItemPosition(
				0,
				taskAfterAfter !== null ? taskAfterAfter.kanbanPosition : null,
			)

			await taskStore.update(newTaskAfter)
		}
	} finally {
		taskUpdating.value[task.id] = false
		oneTaskUpdating.value = false
	}
}

function toggleShowNewTaskInput(bucketId: IBucket['id']) {
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
	scrollTaskContainerToBottom(bucketId)
}

function scrollTaskContainerToBottom(bucketId: IBucket['id']) {
	const bucketEl = taskContainerRefs.value[bucketId]
	if (!bucketEl) {
		return
	}
	bucketEl.scrollTop = bucketEl.scrollHeight
}

async function createNewBucket() {
	if (newBucketTitle.value === '') {
		return
	}

	await kanbanStore.createBucket(new BucketModel({
		title: newBucketTitle.value,
		projectId: project.value.id,
	}))
	newBucketTitle.value = ''
	showNewBucketInput.value = false
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
	await kanbanStore.updateBucketTitle({
		id: bucketId,
		title: bucketTitle,
	})
	bucketTitleEditable.value = false
}

function updateBuckets(value: IBucket[]) {
	// (1) buckets get updated in store and tasks positions get invalidated
	kanbanStore.setBuckets(value)
}

// TODO: fix type
function updateBucketPosition(e: {newIndex: number}) {
	// (2) bucket positon is changed
	dragBucket.value = false

	const bucket = buckets.value[e.newIndex]
	const bucketBefore = buckets.value[e.newIndex - 1] ?? null
	const bucketAfter = buckets.value[e.newIndex + 1] ?? null

	kanbanStore.updateBucket({
		id: bucket.id,
		position: calculateItemPosition(
			bucketBefore !== null ? bucketBefore.position : null,
			bucketAfter !== null ? bucketAfter.position : null,
		),
	})
}

async function setBucketLimit(bucketId: IBucket['id'], limit: number) {
	if (limit < 0) {
		return
	}

	await kanbanStore.updateBucket({
		...kanbanStore.getBucketById(bucketId),
		limit,
	})
	success({message: t('project.kanban.bucketLimitSavedSuccess')})
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

async function toggleDoneBucket(bucket: IBucket) {
	await kanbanStore.updateBucket({
		...bucket,
		isDoneBucket: !bucket.isDoneBucket,
	})
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

<style lang="scss">
$ease-out: all .3s cubic-bezier(0.23, 1, 0.32, 1);
$bucket-width: 300px;
$bucket-header-height: 60px;
$bucket-right-margin: 1rem;

$crazy-height-calculation: '100vh - 4.5rem - 1.5rem - 1rem - 1.5rem - 11px';
$crazy-height-calculation-tasks: '#{$crazy-height-calculation} - 1rem - 2.5rem - 2rem - #{$button-height} - 1rem';
$filter-container-height: '1rem - #{$switch-view-height}';

// FIXME:
.app-content.project\.kanban, .app-content.task\.detail {
	padding-bottom: 0 !important;
}

.kanban {
	overflow-x: auto;
	overflow-y: hidden;
	height: calc(#{$crazy-height-calculation});
	margin: 0 -1.5rem;
	padding: 0 1.5rem;

	@media screen and (max-width: $tablet) {
		height: calc(#{$crazy-height-calculation} - #{$filter-container-height});
		scroll-snap-type: x mandatory;
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
			top: 0.25rem;
			right: 0.5rem;
			bottom: 0.25rem;
			left: 0.5rem;
			border: 3px dashed var(--grey-300);
			border-radius: $radius;
		}
	}

	.bucket {
		border-radius: $radius;
		position: relative;

		margin: 0 $bucket-right-margin 0 0;
		max-height: calc(100% - 1rem); // 1rem spacing to the bottom
		min-height: 20px;
		width: $bucket-width;
		display: flex;
		flex-direction: column;
		overflow: hidden; // Make sure the edges are always rounded

		@media screen and (max-width: $tablet) {
			scroll-snap-align: center;
		}

		.tasks {
			overflow: hidden auto;
			height: 100%;
		}

		.task-item {
			background-color: var(--grey-100);
			padding: .25rem .5rem;

			&:first-of-type {
				padding-top: .5rem;
			}
			&:last-of-type {
				padding-bottom: .5rem;
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
			min-width: calc(#{$bucket-width} + 1rem);
			background: transparent;
			padding-right: 1rem;

			.button {
				background: var(--grey-100);
				width: 100%;
			}
		}

		&.is-collapsed {
			align-self: flex-start;
			transform: rotate(90deg) translateY(-100%);
			transform-origin: top left;
			// Using negative margins instead of translateY here to make all other buckets fill the empty space
			margin-right: calc((#{$bucket-width} - #{$bucket-header-height} - #{$bucket-right-margin}) * -1);
			cursor: pointer;

			.tasks, .bucket-footer {
				display: none;
			}
		}
	}

	.bucket-header {
		background-color: var(--grey-100);
		height: min-content;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: .5rem;
		height: $bucket-header-height;

		.limit {
			padding: 0 .5rem;
			font-weight: bold;

			&.is-max {
				color: var(--danger);
			}
		}

		.title.input {
			height: auto;
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
		bottom: 0;
		height: min-content;
		padding: .5rem;
		background-color: var(--grey-100);
		border-bottom-left-radius: $radius;
		border-bottom-right-radius: $radius;
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