<template>
	<template v-if="kanbanView">
		<span class="has-text-grey-light"> &gt; </span>
		<template v-if="canWrite">
			<Dropdown>
				<template #trigger="{toggleOpen, open}">
					<BaseButton
						class="bucket-name"
						:aria-expanded="open"
						:aria-label="$t('task.detail.bucketSelectLabel', {bucket: currentBucketTitle})"
						@click="toggleOpen"
					>
						{{ currentBucketTitle }}
						<Icon
							icon="pencil-alt"
							class="change-indicator d-print-none"
						/>
					</BaseButton>
				</template>
				<template #default="{close}">
					<div @click="close">
						<DropdownItem
							v-for="bucket in buckets"
							:key="bucket.id"
							:class="{'is-active': currentBucket?.id === bucket.id}"
							@click="changeBucket(bucket)"
						>
							{{ bucket.title }}
						</DropdownItem>
					</div>
				</template>
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
import {bucketsQuery} from '@/client/queries/kanban'
import {useMoveTaskMutation} from '@/client/queries/taskMutations'
import {useQuery} from '@tanstack/vue-query'
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'

import type {Task as ITask} from '@/client/generated'
import type {Bucket as IBucket} from '@/client/generated'

import {PROJECT_VIEW_KINDS} from '@/constants/projectView'

import {useProjects} from '@/composables/useProjects'
import {useBaseStore} from '@/stores/base'

import BaseButton from '@/components/base/BaseButton.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'


import {success} from '@/message'

const props = defineProps<{
	task: ITask
	canWrite: boolean
}>()


const {t} = useI18n({useScope: 'global'})

const projectList = useProjects()
const moveMutation = useMoveTaskMutation()
const baseStore = useBaseStore()

const project = computed(() => projectList.projects[props.task.project_id ?? 0])

// If the project has exactly one manual kanban view, always use it.
// If there are multiple, only show the selector when the active view is one of them.
const kanbanView = computed(() => {
	if (!project.value?.views) {
		return null
	}

	const manualKanbanViews = project.value.views.filter(
		view => view.view_kind === PROJECT_VIEW_KINDS.KANBAN
			&& view.bucket_configuration_mode === 'manual',
	)

	if (manualKanbanViews.length === 1) {
		return manualKanbanViews[0]
	}

	if (manualKanbanViews.length > 1) {
		const activeViewId = baseStore.currentProjectViewId
		return manualKanbanViews.find(v => v.id === activeViewId) || null
	}

	return null
})

const bucketsQueryResult = useQuery(computed(() => bucketsQuery(props.task.project_id ?? 0, kanbanView.value?.id ?? 0)))
const buckets = computed(() => bucketsQueryResult.data.value ?? [])

const currentBucket = computed(() => {
	if (!kanbanView.value) {
		return undefined
	}
	return props.task.buckets?.find(b => b.project_view_id === kanbanView.value.id)
})

const currentBucketTitle = computed(() => {
	return currentBucket.value?.title || t('task.detail.noBucket')
})

async function changeBucket(bucket: IBucket) {
	if (!kanbanView.value || currentBucket.value?.id === bucket.id) {
		return
	}

	await moveMutation.mutateAsync({
		project: props.task.project_id!,
		view: kanbanView.value.id!,
		bucket: bucket.id!,
		task: props.task,
	})

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
