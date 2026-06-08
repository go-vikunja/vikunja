<template>
	<p
		v-if="rows.length === 0"
		class="has-text-centered has-text-grey is-italic"
	>
		{{ emptyText }}
	</p>
	<component
		:is="card ? Card : 'div'"
		v-else
		v-bind="card ? {padding: false, hasContent: false} : {}"
	>
		<div class="has-horizontal-overflow">
			<table class="table has-actions is-hoverable is-fullwidth mbe-0">
				<thead>
					<tr>
						<th v-if="!hideLabelColumn">
							{{ $t('task.attributes.project') }}
						</th>
						<th v-if="!hideLabelColumn">
							{{ $t('timeTracking.form.task') }}
						</th>
						<th>{{ $t('task.comment.comment') }}</th>
						<th class="nowrap">
							{{ $t('timeTracking.list.time') }}
						</th>
						<th class="nowrap has-text-right">
							{{ $t('timeTracking.list.duration') }}
						</th>
						<th />
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="row in rows"
						:key="row.entry.id"
						v-cy="'timeEntry'"
					>
						<td v-if="!hideLabelColumn">
							<template
								v-for="(project, i) in row.projectChain"
								:key="project.id"
							>
								<RouterLink :to="{ name: 'project.index', params: { projectId: project.id } }">
									{{ project.title }}
								</RouterLink>
								<span
									v-if="i < row.projectChain.length - 1"
									class="has-text-grey"
								> &gt; </span>
							</template>
						</td>
						<td v-if="!hideLabelColumn">
							<RouterLink
								v-if="row.entry.taskId > 0"
								:to="{ name: 'task.detail', params: { id: row.entry.taskId } }"
							>
								{{ row.taskIdentifier }}{{ row.taskTitle ? ` - ${row.taskTitle}` : '' }}
							</RouterLink>
						</td>
						<td class="has-text-grey">
							{{ row.entry.comment }}
						</td>
						<td class="nowrap has-text-grey">
							{{ timeRange(row.entry) }}
						</td>
						<td class="nowrap has-text-right has-text-weight-semibold">
							{{ row.seconds === null ? '' : formatDuration(row.seconds) }}
						</td>
						<td class="nowrap has-text-right">
							<template v-if="row.entry.userId === currentUserId">
								<BaseButton
									v-tooltip="$t('menu.edit')"
									v-cy="'editTimeEntry'"
									class="entry-action"
									:aria-label="$t('menu.edit')"
									@click="emit('edit', row.entry)"
								>
									<Icon icon="pen" />
								</BaseButton>
								<BaseButton
									v-tooltip="$t('misc.delete')"
									v-cy="'deleteTimeEntry'"
									class="entry-action entry-delete"
									:aria-label="$t('misc.delete')"
									@click="emit('delete', row.entry.id)"
								>
									<Icon icon="trash-alt" />
								</BaseButton>
							</template>
						</td>
					</tr>
				</tbody>
				<tfoot>
					<tr>
						<td
							:colspan="hideLabelColumn ? 2 : 4"
							class="has-text-weight-bold"
						>
							{{ $t('timeTracking.list.total') }}
						</td>
						<td class="nowrap has-text-right has-text-weight-bold">
							{{ formatDuration(totalSeconds) }}
						</td>
						<td />
					</tr>
				</tfoot>
			</table>
		</div>
	</component>
</template>

<script setup lang="ts">
import {ref, computed, watch} from 'vue'

import Card from '@/components/misc/Card.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {formatDate} from '@/helpers/time/formatDate'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'

import type {ITimeEntry} from '@/modelTypes/ITimeEntry'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'

const props = withDefaults(defineProps<{
	entries: ITimeEntry[]
	// Drop the project + task columns when every entry belongs to the same task
	// (e.g. the task-detail page).
	hideLabelColumn?: boolean
	// Wrap the table in a Card box; set false to render it inline (no card background).
	card?: boolean
	// Override the empty-state message (defaults to the per-day wording).
	emptyText?: string
}>(), {
	hideLabelColumn: false,
	card: true,
	emptyText: '',
})

const emit = defineEmits<{
	delete: [id: number]
	edit: [entry: ITimeEntry]
}>()

const projectStore = useProjectStore()
const {store: timeFormat} = useTimeFormat()

// Only the author can update/delete (enforced server-side); shared lists include
// others' entries, so hide the controls on rows the current user doesn't own.
const authStore = useAuthStore()
const currentUserId = computed(() => authStore.info?.id)

// Task entries carry only a task id; resolve the full task lazily (for its
// title, identifier, and parent project) and cache it.
const taskService = new TaskService()
const tasks = ref<Record<number, ITask>>({})
const inFlight = new Set<number>()
async function ensureTask(taskId: number) {
	if (taskId === 0 || tasks.value[taskId] !== undefined || inFlight.has(taskId)) {
		return
	}
	inFlight.add(taskId)
	try {
		tasks.value[taskId] = await taskService.get(new TaskModel({id: taskId}))
	} catch {
		// Leave unresolved — the row falls back to #<id>.
	} finally {
		inFlight.delete(taskId)
	}
}

watch(() => props.entries, entries => {
	entries.forEach(entry => ensureTask(entry.taskId))
}, {immediate: true})

function entrySeconds(entry: ITimeEntry): number {
	const end = entry.endTime ?? new Date()
	return Math.floor((end.getTime() - entry.startTime.getTime()) / 1000)
}

const rows = computed(() => props.entries.map(entry => {
	const task = entry.taskId > 0 ? tasks.value[entry.taskId] : undefined
	const projectId = task?.projectId ?? (entry.projectId > 0 ? entry.projectId : 0)
	const project = projectId > 0 ? projectStore.projects[projectId] as IProject | undefined : undefined
	const ancestors = project ? projectStore.getAncestors(project) : []

	return {
		entry,
		// Full ancestor chain (root → leaf), each link-able.
		projectChain: ancestors.map(p => ({id: p.id, title: getProjectTitle(p)})),
		taskIdentifier: task ? (task.identifier || `#${task.index}`) : (entry.taskId > 0 ? `#${entry.taskId}` : ''),
		taskTitle: task?.title ?? '',
		// A running entry (no end) has no settled duration — leave it blank.
		seconds: entry.endTime !== null ? entrySeconds(entry) : null,
	}
}))

const totalSeconds = computed(() => rows.value.reduce((sum, row) => sum + (row.seconds ?? 0), 0))

function formatDuration(seconds: number): string {
	const hours = Math.floor(seconds / 3600)
	const minutes = Math.floor((seconds % 3600) / 60)
	return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`
}

function formatTime(date: Date): string {
	return formatDate(date, timeFormat.value === TIME_FORMAT.HOURS_24 ? 'HH:mm' : 'hh:mm A')
}

function timeRange(entry: ITimeEntry): string {
	const start = formatTime(entry.startTime)
	if (entry.endTime === null) {
		return `${start} – …`
	}
	return `${start} – ${formatTime(entry.endTime)}`
}
</script>

<style lang="scss" scoped>
.nowrap {
	white-space: nowrap;
}

.entry-action {
	color: var(--grey-400);
	transition: color $transition;

	& + & {
		margin-inline-start: .5rem;
	}

	&:hover {
		color: var(--primary);
	}
}

.entry-delete:hover {
	color: var(--danger);
}
</style>
