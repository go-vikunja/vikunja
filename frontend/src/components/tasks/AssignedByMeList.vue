<template>
	<div
		v-cy="'assignedByMe'"
		class="is-max-width-desktop has-text-start"
	>
		<div class="assigned-by-me__header">
			<h2 class="title">
				{{ $t('task.assignedByMe.title') }}
			</h2>
			<FancyCheckbox v-model="showCompleted">
				{{ $t('task.assignedByMe.showCompleted') }}
			</FancyCheckbox>
		</div>
		<p class="help mbe-2">
			{{ $t('task.assignedByMe.help') }}
		</p>

		<p
			v-if="query.isError.value"
			class="has-text-danger"
		>
			{{ $t('task.assignedByMe.loadError') }}
		</p>
		<p
			v-else-if="!query.isPending.value && tasks.length === 0"
			class="assigned-by-me__empty"
		>
			{{ $t('task.assignedByMe.empty') }}
		</p>
		<Card
			v-else
			:padding="false"
			:has-content="false"
			:loading="query.isFetching.value"
		>
			<div class="has-horizontal-overflow">
				<table class="table has-actions is-hoverable is-fullwidth mbe-0">
					<thead>
						<tr>
							<th>{{ $t('task.attributes.title') }}</th>
							<th>{{ $t('task.attributes.project') }}</th>
							<th>{{ $t('task.assignedByMe.assignee') }}</th>
							<th>{{ $t('task.attributes.dueDate') }}</th>
							<th>{{ $t('task.assignedByMe.status') }}</th>
							<th>{{ $t('task.attributes.priority') }}</th>
						</tr>
					</thead>
					<tbody>
						<tr
							v-for="t in tasks"
							:key="t.id"
						>
							<td>
								<RouterLink :to="{name: 'task.detail', params: {id: t.id}}">
									{{ t.title }}
								</RouterLink>
							</td>
							<td>
								<RouterLink
									v-if="projectList.projects[t.project_id]"
									:to="{name: 'project.index', params: {projectId: t.project_id}}"
								>
									{{ projectList.projects[t.project_id].title }}
								</RouterLink>
							</td>
							<td>
								<AssigneeList
									:assignees="othersOf(t)"
									:avatar-size="28"
									:inline="true"
								/>
								<span class="assigned-by-me__names">{{ namesOf(t) }}</span>
							</td>
							<DateTableCell :date="t.due_date" />
							<td>
								<Done
									v-if="t.done"
									:is-done="true"
									variant="small"
								/>
								<span v-else>{{ Math.round(t.percent_done * 100) }}%</span>
							</td>
							<td>
								<PriorityLabel
									:priority="t.priority"
									:done="t.done"
									:show-all="true"
								/>
							</td>
						</tr>
					</tbody>
				</table>
			</div>
		</Card>
	</div>
</template>

<script setup lang="ts">
import {useStorage} from '@vueuse/core'

import Card from '@/components/misc/Card.vue'
import Done from '@/components/misc/Done.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import DateTableCell from '@/components/tasks/partials/DateTableCell.vue'

import {useTasks} from '@/composables/useTasks'
import {useProjects} from '@/composables/useProjects'
import {useAuthStore} from '@/stores/auth'
import {withAssignment} from '@/client/queries/tasks'
import type {TaskResponse} from '@/client/queries/tasks'

defineOptions({name: 'AssignedByMeList'})

const authStore = useAuthStore()
const projectList = useProjects()

// The list exists to follow what was handed out, and finished work is useful for a while, but
// off by default so the open work leads.
const showCompleted = useStorage('assignedByMeShowCompleted', false)

const query = useTasks(
	() => ({
		project: null,
		params: withAssignment({
			sort_by: ['due_date', 'id'],
			order_by: ['asc', 'desc'],
			filter: showCompleted.value ? '' : 'done = false',
			filter_include_nulls: showCompleted.value,
			q: '',
			filter_timezone: authStore.settings.timezone,
		}, 'assigned_by_me'),
	}),
	{enabled: () => authStore.authenticated},
)
const tasks = query.tasks

// The column is about the people it was given to, not the viewer who may also be an assignee.
function othersOf(task: TaskResponse) {
	return task.assignees.filter(a => a.id !== authStore.info?.id)
}

function namesOf(task: TaskResponse): string {
	return othersOf(task).map(a => a.name || a.username).join(', ')
}
</script>

<style lang="scss" scoped>
.assigned-by-me__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 1rem;
	flex-wrap: wrap;
}

.assigned-by-me__empty {
	color: var(--grey-500);
}

.assigned-by-me__names {
	margin-inline-start: 0.5rem;
}
</style>
