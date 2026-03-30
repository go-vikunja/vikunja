<template>
	<div
		v-if="kanbanViews.length > 0"
		class="bucket-select"
	>
		<template v-if="kanbanViews.length > 1">
			<div
				v-for="view in kanbanViews"
				:key="view.id"
				class="bucket-view-group"
			>
				<span class="bucket-view-label">{{ view.title }}</span>
				<div class="select">
					<select
						:value="selectedBucketByView[view.id] ?? 0"
						:disabled="disabled || loading"
						data-cy="bucket-select"
						@change="changeBucket(view, ($event.target as HTMLSelectElement).value)"
					>
						<option
							:value="0"
							disabled
						>
							{{ $t('task.detail.bucket.placeholder') }}
						</option>
						<option
							v-for="bucket in bucketsByView[view.id]"
							:key="bucket.id"
							:value="bucket.id"
						>
							{{ bucket.title }}
						</option>
					</select>
				</div>
			</div>
		</template>
		<template v-else>
			<div class="select">
				<select
					:value="selectedBucketByView[kanbanViews[0].id] ?? 0"
					:disabled="disabled || loading"
					data-cy="bucket-select"
					@change="changeBucket(kanbanViews[0], ($event.target as HTMLSelectElement).value)"
				>
					<option
						:value="0"
						disabled
					>
						{{ $t('task.detail.bucket.placeholder') }}
					</option>
					<option
						v-for="bucket in bucketsByView[kanbanViews[0].id]"
						:key="bucket.id"
						:value="bucket.id"
					>
						{{ bucket.title }}
					</option>
				</select>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'

import type {ITask} from '@/modelTypes/ITask'
import type {IProjectView} from '@/modelTypes/IProjectView'
import type {IBucket} from '@/modelTypes/IBucket'
import {PROJECT_VIEW_KINDS} from '@/modelTypes/IProjectView'

import BucketModel from '@/models/bucket'
import BucketService from '@/services/bucket'
import TaskBucketService from '@/services/taskBucket'
import TaskBucketModel from '@/models/taskBucket'

import {useProjectStore} from '@/stores/projects'
import {useBaseStore} from '@/stores/base'
import {useKanbanStore} from '@/stores/kanban'

import {success} from '@/message'
import {useI18n} from 'vue-i18n'

const props = withDefaults(defineProps<{
	task: ITask,
	disabled?: boolean,
}>(), {
	disabled: false,
})

const emit = defineEmits<{
	'update:task': [task: ITask],
}>()

const {t} = useI18n({useScope: 'global'})

const projectStore = useProjectStore()
const baseStore = useBaseStore()
const kanbanStore = useKanbanStore()

const loading = ref(false)
const bucketsByView = ref<Record<number, IBucket[]>>({})
const selectedBucketByView = ref<Record<number, number>>({})

const kanbanViews = computed(() => {
	const project = projectStore.projects[props.task.projectId]
	if (!project?.views) return [] as IProjectView[]
	return project.views.filter(v => v.viewKind === PROJECT_VIEW_KINDS.KANBAN) as IProjectView[]
})

async function loadBucketsForView(view: IProjectView) {
	const bucketService = new BucketService()
	const buckets = await bucketService.getAll(new BucketModel({
		projectId: props.task.projectId,
		projectViewId: view.id,
	}))
	bucketsByView.value[view.id] = buckets
}

async function loadAllBuckets() {
	loading.value = true
	try {
		await Promise.all(kanbanViews.value.map(loadBucketsForView))
	} finally {
		loading.value = false
	}
}

async function loadTaskBucketPositions() {
	// For each kanban view, figure out which bucket the task is in
	// We check the loaded buckets to find which one contains this task
	for (const view of kanbanViews.value) {
		const viewBuckets = bucketsByView.value[view.id]
		if (!viewBuckets) continue

		// Check if buckets have tasks loaded (from kanban store)
		// If the current view is this kanban view, use the kanban store
		if (baseStore.currentProjectViewId === view.id && kanbanStore.buckets.length > 0) {
			const result = kanbanStore.getTaskById(props.task.id)
			if (result.bucketIndex !== null) {
				selectedBucketByView.value[view.id] = kanbanStore.buckets[result.bucketIndex].id
				continue
			}
		}

		// Use the task's bucketId if available (it corresponds to the current view)
		if (props.task.bucketId && props.task.bucketId !== 0) {
			// Check if this bucketId belongs to this view
			const matchingBucket = viewBuckets.find(b => b.id === props.task.bucketId)
			if (matchingBucket) {
				selectedBucketByView.value[view.id] = matchingBucket.id
				continue
			}
		}

		// Fallback: try to find via task in bucket tasks
		for (const bucket of viewBuckets) {
			if (bucket.tasks?.some(t => t.id === props.task.id)) {
				selectedBucketByView.value[view.id] = bucket.id
				break
			}
		}
	}
}

async function changeBucket(view: IProjectView, newBucketIdStr: string) {
	const newBucketId = Number(newBucketIdStr)
	if (!newBucketId || newBucketId === selectedBucketByView.value[view.id]) return

	loading.value = true
	try {
		const taskBucketService = new TaskBucketService()
		const updatedTaskBucket = await taskBucketService.update(new TaskBucketModel({
			taskId: props.task.id,
			bucketId: newBucketId,
			projectViewId: view.id,
			projectId: props.task.projectId,
		}))

		selectedBucketByView.value[view.id] = updatedTaskBucket.bucketId

		// Update the kanban store if this view is currently active
		if (baseStore.currentProjectViewId === view.id && kanbanStore.buckets.length > 0) {
			kanbanStore.moveTaskToBucket(props.task, updatedTaskBucket.bucketId)
			if (updatedTaskBucket.bucket) {
				kanbanStore.setBucketById(updatedTaskBucket.bucket, false)
			}
		}

		// Emit updated task
		if (updatedTaskBucket.task) {
			emit('update:task', updatedTaskBucket.task)
		}

		success({message: t('task.detail.bucket.updateSuccess')})
	} catch {
		// Revert on error - selectedBucketByView will stay as-is since we only update on success
	} finally {
		loading.value = false
	}
}

watch(() => props.task.id, async () => {
	if (props.task.id && kanbanViews.value.length > 0) {
		await loadAllBuckets()
		await loadTaskBucketPositions()
	}
}, {immediate: true})

watch(kanbanViews, async (views) => {
	if (views.length > 0 && props.task.id) {
		await loadAllBuckets()
		await loadTaskBucketPositions()
	}
})
</script>

<style lang="scss" scoped>
.bucket-select {
	.bucket-view-group {
		margin-block-end: 0.5rem;

		&:last-child {
			margin-block-end: 0;
		}
	}

	.bucket-view-label {
		display: block;
		font-size: 0.85em;
		color: var(--grey-500);
		margin-block-end: 0.25rem;
	}

	.select {
		width: 100%;

		select {
			width: 100%;
		}
	}
}
</style>
