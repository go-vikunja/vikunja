<template>
	<span v-if="state.status === 'invalid'">{{ href }}</span>
	<a
		v-else-if="state.status === 'error'"
		:href="href"
		class="task-link-pill task-link-pill--fallback"
		target="_blank"
		rel="noopener noreferrer nofollow"
	>{{ href }}</a>
	<span
		v-else-if="state.status === 'loading'"
		class="task-link-pill task-link-pill--loading"
	>{{ href }}</span>
	<TaskGlanceTooltip
		v-else-if="state.status === 'loaded'"
		:task="state.task"
	>
		<a
			:href="href"
			class="task-link-pill task-link-pill--task"
			:class="{'task-link-pill--done': state.task.done}"
			@click="onClick"
		>
			<span
				v-if="projectPrefix"
				class="task-link-pill__project"
			>{{ projectPrefix }} <span aria-hidden="true">›</span></span>
			<span class="task-link-pill__identifier">{{ getTaskIdentifier(state.task) }}</span>
			<span class="task-link-pill__title">{{ state.task.title }}</span>
			<template v-if="state.task.done">
				<Icon
					icon="check-double"
					class="task-link-pill__done"
				/>
				<span class="is-sr-only">{{ $t('task.attributes.done') }}</span>
			</template>
		</a>
	</TaskGlanceTooltip>
</template>

<script lang="ts" setup>
import {computed, inject} from 'vue'

import type {TaskResponse} from '@/client/queries/tasks'
import {getTaskIdentifier} from '@/helpers/task'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {parseTaskIdFromUrl} from '@/helpers/parseTaskIdFromUrl'
import {useTask} from '@/composables/useTask'
import {useBaseStore} from '@/stores/base'
import {useProjects} from '@/composables/useProjects'
import TaskGlanceTooltip from '@/components/tasks/partials/TaskGlanceTooltip.vue'

import {taskLinkCurrentProjectIdKey} from './taskLinkContext'

type PillState =
	| {status: 'loading'}
	| {status: 'loaded', task: TaskResponse}
	| {status: 'error'}
	| {status: 'invalid'}

const props = defineProps<{
	href: string
}>()

const emit = defineEmits<{
	open: [task: TaskResponse]
}>()

const baseStore = useBaseStore()
const projectList = useProjects()
const providedProjectId = inject(taskLinkCurrentProjectIdKey, null)

const currentProjectId = computed(() => providedProjectId?.value ?? (baseStore.currentProjectId || undefined))

const taskId = computed(() => parseTaskIdFromUrl(props.href))
const query = useTask(() => taskId.value ?? 0)
const state = computed<PillState>(() => {
	if (taskId.value === null) return {status: 'invalid'}
	if (query.data.value) return {status: 'loaded', task: query.data.value}
	return {status: query.isError.value ? 'error' : 'loading'}
})

const projectPrefix = computed(() => {
	if (state.value.status !== 'loaded') {
		return ''
	}
	const projectId = state.value.task.project_id
	if (currentProjectId.value === projectId) {
		return ''
	}
	const project = projectList.projects[projectId]
	return project ? getProjectTitle(project) : ''
})

function onClick(e: MouseEvent) {
	// Leave modified clicks to the browser so they can open a new tab.
	if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) {
		return
	}
	e.preventDefault()
	if (state.value.status === 'loaded') {
		emit('open', state.value.task)
	}
}
</script>

<style lang="scss" scoped>
.task-link-pill--task {
	display: inline-flex;
	align-items: center;
	gap: .35em;
	max-inline-size: 100%;
	padding: .05em .5em;
	border: 1px solid var(--grey-400);
	border-radius: 9999px;
	background: var(--grey-200);
	color: var(--text);
	text-decoration: none;
	line-height: 1.5;
	vertical-align: baseline;
	transition: background-color $transition, border-color $transition;

	&:hover {
		background: var(--grey-300);
		border-color: var(--grey-500);
		color: var(--text);
	}

	.task-link-pill__project,
	.task-link-pill__identifier {
		color: var(--grey-600);
		font-size: .85em;
		white-space: nowrap;
	}

	.task-link-pill__title {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.task-link-pill__done {
		color: var(--grey-600);
		font-size: .85em;
	}

	&.task-link-pill--done .task-link-pill__title {
		text-decoration: line-through;
		color: var(--grey-600);
	}
}

.task-link-pill--loading {
	color: var(--link);
}
</style>
