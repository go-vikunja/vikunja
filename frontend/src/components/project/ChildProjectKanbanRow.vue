<template>
	<div class="child-project-kanban-row">
		<div class="child-project-divider">
			<span class="child-project-divider__line" />
			<RouterLink
				class="child-project-divider__title"
				:to="{ name: 'project.index', params: { projectId: project.id } }"
			>
				{{ project.title }}
			</RouterLink>
			<span class="child-project-divider__line" />
		</div>

		<div
			v-if="isLoading"
			class="child-kanban-buckets loader-container is-loading"
		/>

		<div
			v-else-if="buckets.length === 0"
			class="child-kanban-empty"
		>
			{{ $t('project.kanban.noBuckets') }}
		</div>

		<ul
			v-else
			class="child-kanban-buckets"
		>
			<li
				v-for="bucket in buckets"
				:key="bucket.id"
				class="bucket"
			>
				<div class="bucket-header">
					<span
						v-if="kanbanView && bucket.id === kanbanView.doneBucketId"
						v-tooltip="$t('project.kanban.doneBucketHint')"
						class="icon is-small has-text-success mie-2"
					>
						<Icon icon="check-double" />
					</span>
					<h2 class="title">
						{{ bucket.title }}
					</h2>
					<span
						v-if="bucket.limit > 0 || bucket.count > 0"
						class="limit"
						:class="{'is-max': bucket.limit > 0 && bucket.count >= bucket.limit}"
					>
						{{ bucket.limit > 0 ? `${bucket.count}/${bucket.limit}` : bucket.count }}
					</span>
				</div>

				<draggable
					:model-value="bucket.tasks"
					:group="`child-project-${project.id}`"
					:animation="150"
					ghost-class="ghost"
					drag-class="task-dragging"
					:delay-on-touch-only="true"
					:delay="1000"
					item-key="id"
					tag="ul"
					class="tasks"
					:data-bucket-id="bucket.id"
					@update:modelValue="(tasks) => updateBucketTasks(bucket.id, tasks)"
					@start="(e) => handleDragStart(e, bucket.id)"
					@end="updateTaskPosition"
				>
					<template #item="{element: task}">
						<div
							class="task-item"
							:data-task-id="task.id"
						>
							<KanbanCard
								class="kanban-card"
								:task="task"
								:loading="taskUpdating[task.id] ?? false"
								:project-id="project.id"
							/>
						</div>
					</template>
				</draggable>
			</li>
		</ul>
	</div>
</template>

<script setup lang="ts">
import {ref, watch, computed} from 'vue'
import draggable from 'zhyswan-vuedraggable'
import {klona} from 'klona/lite'

import KanbanCard from '@/components/tasks/partials/KanbanCard.vue'
import TaskCollectionService, {type TaskFilterParams} from '@/services/taskCollection'
import TaskPositionService from '@/services/taskPosition'
import TaskPositionModel from '@/models/taskPosition'
import TaskBucketService from '@/services/taskBucket'
import TaskBucketModel from '@/models/taskBucket'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'

import type {IProject} from '@/modelTypes/IProject'
import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'
import type {IProjectView} from '@/modelTypes/IProjectView'
import {PROJECT_VIEW_KINDS} from '@/modelTypes/IProjectView'

const props = defineProps<{
	project: IProject
	filterParams?: Partial<TaskFilterParams>
}>()

const isLoading = ref(false)
const buckets = ref<IBucket[]>([])
const taskUpdating = ref<{ [id: ITask['id']]: boolean }>({})
const sourceBucketId = ref<IBucket['id'] | null>(null)

const kanbanView = computed<IProjectView | undefined>(() =>
	props.project.views?.find(v => v.viewKind === PROJECT_VIEW_KINDS.KANBAN),
)

async function loadBuckets() {
	if (!kanbanView.value) return

	isLoading.value = true
	try {
		const service = new TaskCollectionService()
		const result = await service.getAll(
			{projectId: props.project.id, viewId: kanbanView.value.id},
			{
				...(props.filterParams ?? {}),
				expand: ['comment_count', 'is_unread'],
				per_page: 25,
				include_child_tasks: false,
			},
		)
		buckets.value = result as IBucket[]
	} finally {
		isLoading.value = false
	}
}

watch(
	() => [props.project.id, kanbanView.value?.id, props.filterParams],
	() => loadBuckets(),
	{immediate: true, deep: true},
)

function updateBucketTasks(bucketId: IBucket['id'], tasks: ITask[]) {
	const bucket = buckets.value.find(b => b.id === bucketId)
	if (bucket) {
		bucket.tasks = tasks
	}
}

function handleDragStart(e: {item: HTMLElement}, bucketId: IBucket['id']) {
	sourceBucketId.value = bucketId
}

const taskPositionService = ref(new TaskPositionService())
const taskBucketService = ref(new TaskBucketService())

async function updateTaskPosition(e: {newIndex: number, to: HTMLElement, from: HTMLElement}) {
	if (!kanbanView.value) return

	// Get destination bucket ID from the data attribute on the drop target
	const destBucketId = parseInt(e.to.dataset.bucketId ?? '0', 10)
	if (!destBucketId) return

	const destBucket = buckets.value.find(b => b.id === destBucketId)
	if (!destBucket) return

	const task = destBucket.tasks[e.newIndex]
	if (!task) return

	const taskBefore = destBucket.tasks[e.newIndex - 1] ?? null
	const taskAfter = destBucket.tasks[e.newIndex + 1] ?? null
	const position = calculateItemPosition(
		taskBefore?.position ?? null,
		taskAfter?.position ?? null,
	)

	taskUpdating.value[task.id] = true
	try {
		const newTask = klona(task)
		newTask.bucketId = destBucketId

		// Always update position
		await taskPositionService.value.update(new TaskPositionModel({
			position,
			projectViewId: kanbanView.value.id,
			taskId: task.id,
		}))
		newTask.position = position

		// Update bucket if it changed
		const bucketChanged = sourceBucketId.value !== null && sourceBucketId.value !== destBucketId
		if (bucketChanged) {
			const updatedTaskBucket = await taskBucketService.value.update(new TaskBucketModel({
				taskId: newTask.id,
				bucketId: destBucketId,
				projectViewId: kanbanView.value.id,
				projectId: props.project.id,
			}))
			// Merge updated task data (done status etc) from server response
			if (updatedTaskBucket?.task) {
				Object.assign(newTask, updatedTaskBucket.task)
			}
			newTask.bucketId = destBucketId

			// Update counts
			const srcBucket = buckets.value.find(b => b.id === sourceBucketId.value)
			if (srcBucket) srcBucket.count = Math.max(0, (srcBucket.count ?? 1) - 1)
			destBucket.count = (destBucket.count ?? 0) + 1
		}

		destBucket.tasks[e.newIndex] = newTask
	} finally {
		taskUpdating.value[task.id] = false
		sourceBucketId.value = null
	}
}
</script>

<style lang="scss" scoped>
$bucket-width: 300px;
$bucket-right-margin: 1rem;

.child-project-kanban-row {
	margin-block-start: 1.5rem;
}

.child-project-divider {
	display: flex;
	align-items: center;
	gap: 0.75rem;
	margin-block-end: 1rem;

	&__line {
		flex: 1;
		height: 1px;
		background: var(--grey-300);
	}

	&__title {
		font-weight: 600;
		font-size: 0.9rem;
		color: var(--text);
		white-space: nowrap;
		text-decoration: none;

		&:hover {
			color: var(--primary);
		}
	}
}

.child-kanban-buckets {
	display: flex;
	overflow-x: auto;
	gap: $bucket-right-margin;
	padding-block-end: 0.5rem;
	list-style: none;
	margin: 0;
	padding-inline-start: 0;
}

.child-kanban-empty {
	color: var(--grey-500);
	font-size: 0.9rem;
	padding: 0.5rem 0;
}

.bucket {
	min-inline-size: $bucket-width;
	max-inline-size: $bucket-width;
	background: var(--grey-100);
	border-radius: 0.5rem;
	padding: 0.75rem;
	display: flex;
	flex-direction: column;
	max-block-size: 400px;

	.bucket-header {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		margin-block-end: 0.5rem;

		.title {
			font-size: 1rem;
			font-weight: 600;
			margin: 0;
			flex: 1;
		}

		.limit {
			font-size: 0.8rem;
			color: var(--grey-500);

			&.is-max {
				color: var(--danger);
			}
		}
	}

	.tasks {
		overflow-y: auto;
		flex: 1;
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
}

.task-item {
	cursor: grab;

	&:active {
		cursor: grabbing;
	}
}

.ghost {
	position: relative;

	* {
		opacity: 0;
	}

	&::after {
		content: '';
		position: absolute;
		inset: 0.25rem;
		border: 3px dashed var(--grey-300);
		border-radius: 0.25rem;
	}
}
</style>
