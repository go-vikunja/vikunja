<template>
	<template v-if="kanbanView">
		<span class="has-text-grey-light"> &gt; </span>
		<template v-if="canWrite">
			<Dropdown>
				<template #trigger="{toggleOpen}">
					<BaseButton
						class="bucket-name"
						@click="toggleOpen"
					>
						{{ currentBucketTitle }}
						<Icon
							icon="pencil-alt"
							class="change-indicator"
						/>
					</BaseButton>
				</template>
				<DropdownItem
					v-for="bucket in buckets"
					:key="bucket.id"
					:class="{'is-active': currentBucket?.id === bucket.id}"
					@click="changeBucket(bucket)"
				>
					{{ bucket.title }}
				</DropdownItem>
			</Dropdown>
		</template>
		<span
			v-else
			class="bucket-name"
		>
			{{ currentBucketTitle }}
		</span>
	</template>
</template>

<script lang="ts" setup>
import {ref, computed, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import type {ITask} from '@/modelTypes/ITask'
import type {IBucket} from '@/modelTypes/IBucket'

import {PROJECT_VIEW_KINDS} from '@/modelTypes/IProjectView'

import {useProjectStore} from '@/stores/projects'
import {useKanbanStore} from '@/stores/kanban'
import {useBaseStore} from '@/stores/base'

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
const baseStore = useBaseStore()

const project = computed(() => projectStore.projects[props.task.projectId])

// Only show the bucket selector when the active view is a manual kanban view
const kanbanView = computed(() => {
	if (!project.value?.views) {
		return null
	}

	const activeViewId = baseStore.currentProjectViewId
	if (!activeViewId) {
		return null
	}

	const activeView = project.value.views.find(v => v.id === activeViewId)
	if (activeView
		&& activeView.viewKind === PROJECT_VIEW_KINDS.KANBAN
		&& activeView.bucketConfigurationMode === 'manual') {
		return activeView
	}

	return null
})

const buckets = ref<IBucket[]>([])

watch(
	() => kanbanView.value,
	async (view) => {
		if (!view) {
			buckets.value = []
			return
		}

		const bucketService = new BucketService()
		try {
			buckets.value = await bucketService.getAll({
				projectId: props.task.projectId,
				projectViewId: view.id,
			} as IBucket)
		} catch (e) {
			console.error('Failed to load buckets:', e)
		}
	},
	{immediate: true},
)

const currentBucket = computed(() => {
	if (!kanbanView.value) {
		return undefined
	}
	return props.task.buckets?.find(b => b.projectViewId === kanbanView.value.id)
})

const currentBucketTitle = computed(() => {
	return currentBucket.value?.title || t('task.detail.noBucket')
})

async function changeBucket(bucket: IBucket) {
	if (!kanbanView.value || currentBucket.value?.id === bucket.id) {
		return
	}

	const taskBucketService = new TaskBucketService()
	const updatedTaskBucket = await taskBucketService.update(new TaskBucketModel({
		taskId: props.task.id,
		bucketId: bucket.id,
		projectViewId: kanbanView.value.id,
		projectId: props.task.projectId,
	}))

	const updatedBuckets = (props.task.buckets || []).map(b => {
		if (b.projectViewId === kanbanView.value.id) {
			return {...bucket}
		}
		return b
	})

	if (!updatedBuckets.find(b => b.projectViewId === kanbanView.value.id)) {
		updatedBuckets.push({...bucket})
	}

	kanbanStore.moveTaskToBucket(props.task, bucket.id)

	// Use the task from the API response to pick up done state changes
	// (moving to/from the done bucket toggles the done status)
	const updatedTask = {
		...props.task,
		...updatedTaskBucket.task,
		buckets: updatedBuckets,
		bucketId: bucket.id,
	}

	emit('update:task', updatedTask)

	success({message: t('task.detail.bucketChangedSuccess')})
}
</script>

<style lang="scss" scoped>
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
