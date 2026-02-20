<template>
	<div
		v-for="kanbanView in kanbanViews"
		:key="kanbanView.id"
		class="bucket-select"
	>
		<span class="has-text-grey-light"> &gt; </span>
		<template v-if="canWrite">
			<Dropdown>
				<template #trigger="{toggleOpen}">
					<BaseButton
						class="bucket-name"
						@click="toggleOpen"
					>
						<span
							v-if="kanbanViews.length > 1"
							class="view-title"
						>{{ kanbanView.title }}:</span>
						{{ currentBucketTitle(kanbanView) }}
						<Icon
							icon="pencil-alt"
							class="change-indicator"
						/>
					</BaseButton>
				</template>
				<DropdownItem
					v-for="bucket in viewBuckets[kanbanView.id] || []"
					:key="bucket.id"
					:class="{'is-active': isCurrentBucket(kanbanView, bucket)}"
					@click="changeBucket(kanbanView, bucket)"
				>
					{{ bucket.title }}
				</DropdownItem>
			</Dropdown>
		</template>
		<span
			v-else
			class="bucket-name is-readonly"
		>
			<span
				v-if="kanbanViews.length > 1"
				class="view-title"
			>{{ kanbanView.title }}:</span>
			{{ currentBucketTitle(kanbanView) }}
		</span>
	</div>
</template>

<script lang="ts" setup>
import {ref, computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import type {ITask} from '@/modelTypes/ITask'
import type {IBucket} from '@/modelTypes/IBucket'

import {PROJECT_VIEW_KINDS} from '@/modelTypes/IProjectView'

import {useProjectStore} from '@/stores/projects'
import {useKanbanStore} from '@/stores/kanban'

import BaseButton from '@/components/base/BaseButton.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'

import BucketService from '@/services/bucket'
import TaskBucketService from '@/services/taskBucket'
import TaskBucketModel from '@/models/taskBucket'

import {success} from '@/message'

const props = defineProps<{
	task: ITask
	canWrite: boolean
}>()

const emit = defineEmits<{
	'update:task': [task: ITask]
}>()

const {t} = useI18n({useScope: 'global'})

const projectStore = useProjectStore()
const kanbanStore = useKanbanStore()

const project = computed(() => projectStore.projects[props.task.projectId])

const kanbanViews = computed(() => {
	if (!project.value?.views) {
		return []
	}

	return project.value.views.filter(
		v => v.viewKind === PROJECT_VIEW_KINDS.KANBAN
			&& v.bucketConfigurationMode === 'manual',
	)
})

const viewBuckets = ref<Record<number, IBucket[]>>({})

watch(
	() => kanbanViews.value,
	async (views) => {
		const bucketService = new BucketService()
		for (const view of views) {
			if (viewBuckets.value[view.id]) {
				continue
			}
			try {
				const buckets = await bucketService.getAll({
					projectId: props.task.projectId,
					projectViewId: view.id,
				} as IBucket)
				viewBuckets.value[view.id] = buckets
			} catch {
				// silently ignore if we cannot load buckets
			}
		}
	},
	{immediate: true},
)

function currentBucketForView(view: {id: number}): IBucket | undefined {
	return props.task.buckets?.find(b => b.projectViewId === view.id)
}

function currentBucketTitle(view: {id: number}): string {
	const bucket = currentBucketForView(view)
	return bucket?.title || t('task.detail.noBucket')
}

function isCurrentBucket(view: {id: number}, bucket: IBucket): boolean {
	const current = currentBucketForView(view)
	return current?.id === bucket.id
}

async function changeBucket(view: {id: number}, bucket: IBucket) {
	const current = currentBucketForView(view)
	if (current?.id === bucket.id) {
		return
	}

	const taskBucketService = new TaskBucketService()
	try {
		await taskBucketService.update(new TaskBucketModel({
			taskId: props.task.id,
			bucketId: bucket.id,
			projectViewId: view.id,
			projectId: props.task.projectId,
		}))

		const updatedBuckets = (props.task.buckets || []).map(b => {
			if (b.projectViewId === view.id) {
				return {...bucket}
			}
			return b
		})

		// If the task was not yet in this view, add the bucket
		if (!updatedBuckets.find(b => b.projectViewId === view.id)) {
			updatedBuckets.push({...bucket})
		}

		// Update the kanban store if the board is loaded
		kanbanStore.moveTaskToBucket(props.task, bucket.id)

		emit('update:task', {
			...props.task,
			buckets: updatedBuckets,
			bucketId: bucket.id,
		})

		success({message: t('task.detail.bucketChangedSuccess')})
	} catch {
		// error is handled by the service layer
	}
}
</script>

<style lang="scss" scoped>
.bucket-select {
	display: inline;
}

.bucket-name {
	color: var(--grey-800);

	&:hover {
		color: var(--primary);
	}
}

.change-indicator {
	font-size: .75em;
	margin-inline-start: .25rem;
	color: var(--grey-400);
}

:deep(.dropdown) {
	display: inline;
}

:deep(.dropdown-trigger) {
	display: inline;
	padding: 0;
}
</style>
