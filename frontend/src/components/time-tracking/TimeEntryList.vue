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
								v-if="row.entry.task_id > 0"
								:to="{ name: 'task.detail', params: { id: row.entry.task_id } }"
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
							<template v-if="row.entry.user_id === currentUserId">
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
									@click="emit('delete', row.entry)"
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
							{{ $t(paged ? 'timeTracking.list.pageTotal' : 'timeTracking.list.total') }}
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
import {computed} from 'vue'

import Card from '@/components/misc/Card.vue'
import BaseButton from '@/components/base/BaseButton.vue'

import {useProjects} from '@/composables/useProjects'
import {useAuthStore} from '@/stores/auth'
import {taskQuery} from '@/client/queries/tasks'
import {useQueries} from '@tanstack/vue-query'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {formatDate} from '@/helpers/time/formatDate'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'

import type {TaskResponse} from '@/client/queries/tasks'
import type {TimeEntryResponse as ITimeEntry} from '@/client/queries/timeEntries'

const props = withDefaults(defineProps<{
	entries: ITimeEntry[]
	// Drop the project + task columns when every entry belongs to the same task
	// (e.g. the task-detail page).
	hideLabelColumn?: boolean
	// Wrap the table in a Card box; set false to render it inline (no card background).
	card?: boolean
	// Override the empty-state message (defaults to the per-day wording).
	emptyText?: string
	// The entries are one page of a longer list, so the footer sum is page-local.
	paged?: boolean
}>(), {
	hideLabelColumn: false,
	card: true,
	emptyText: '',
	paged: false,
})

const emit = defineEmits<{
	delete: [entry: ITimeEntry]
	edit: [entry: ITimeEntry]
}>()

const projectList = useProjects()
const {store: timeFormat} = useTimeFormat()

// Only the author can update/delete (enforced server-side); shared lists include
// others' entries, so hide the controls on rows the current user doesn't own.
const authStore = useAuthStore()
const currentUserId = computed(() => authStore.authUser ? authStore.info?.id : undefined)

const taskIds = computed(() => props.hideLabelColumn
	? []
	: [...new Set(props.entries.map(entry => entry.task_id).filter(id => id > 0))])

// Entries carry only a task id; the full task (title, identifier, parent project) is resolved lazily.
const taskQueries = useQueries({
	queries: computed(() => taskIds.value.map(id => taskQuery(id))),
})
const tasks = computed<Record<number, TaskResponse>>(() => Object.fromEntries(
	taskQueries.value.flatMap(result => result.data ? [[result.data.id, result.data] as const] : []),
))

// null when the entry has no settled duration: still running, or unusable timestamps.
function entrySeconds(entry: ITimeEntry): number | null {
	const start = parseDateOrNull(entry.start_time)
	const end = parseDateOrNull(entry.end_time)
	if (start === null || end === null) {
		return null
	}
	return Math.floor((end.getTime() - start.getTime()) / 1000)
}

const rows = computed(() => props.entries.map(entry => {
	const task = entry.task_id > 0 ? tasks.value[entry.task_id] : undefined
	const projectId = task?.project_id ?? (entry.project_id > 0 ? entry.project_id : 0)
	const project = projectId > 0 ? projectList.projects[projectId] : undefined
	const ancestors = project ? projectList.getAncestors(project) : []

	return {
		entry,
		// Full ancestor chain (root → leaf), each link-able.
		projectChain: ancestors.map(p => ({id: p.id, title: getProjectTitle(p)})),
		taskIdentifier: task ? (task.identifier || `#${task.index}`) : (entry.task_id > 0 ? `#${entry.task_id}` : ''),
		taskTitle: task?.title ?? '',
		seconds: entrySeconds(entry),
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
	const start = parseDateOrNull(entry.start_time)
	if (start === null) {
		return ''
	}
	const end = parseDateOrNull(entry.end_time)
	return end === null ? `${formatTime(start)} – …` : `${formatTime(start)} – ${formatTime(end)}`
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
