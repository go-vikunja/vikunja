<script setup lang="ts">
import {ref, computed, onMounted, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'

import TaskCollectionService from '@/services/taskCollection'
import TaskModel from '@/models/task'
import {useProjectStore} from '@/stores/projects'
import {formatDate} from '@/helpers/time/formatDate'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'

const PRIORITY_LABELS: Record<number, string> = {
	0: 'Low',
	1: 'Low',
	2: 'Medium',
	3: 'High',
	4: 'Critical',
	5: 'Urgent',
}

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()

// Raw tasks from API
const allTasks = ref<ITask[]>([])
const isLoading = ref(false)

// Filter state
const selectedProjectId = ref<number | null>(null)
const selectedAssigneeId = ref<number | null>(null)
const dueDateFrom = ref<string>('')
const dueDateTo = ref<string>('')

// Current status tab from query param
const currentFilter = computed<'today' | 'completed'>(() => {
	const q = route.query.filter as string
	if (q === 'completed') return 'completed'
	return 'today'
})

const pageTitle = computed(() =>
	currentFilter.value === 'completed' ? 'Completed Tasks' : "Today's Tasks",
)

const pageDescription = computed(() =>
	currentFilter.value === 'completed'
		? 'All tasks you have completed'
		: 'All your current pending tasks',
)

// Unique projects available in the loaded task list
const availableProjects = computed(() => {
	const seen = new Map<number, string>()
	for (const t of allTasks.value) {
		if (!seen.has(t.projectId)) {
			seen.set(t.projectId, projectStore.projects[t.projectId]?.title ?? `Project ${t.projectId}`)
		}
	}
	return Array.from(seen.entries()).map(([id, title]) => ({id, title}))
})

// Unique assignees available in the loaded task list
const availableAssignees = computed(() => {
	const seen = new Map<number, IUser>()
	for (const t of allTasks.value) {
		for (const a of (t.assignees ?? [])) {
			if (!seen.has(a.id)) {
				seen.set(a.id, a)
			}
		}
	}
	return Array.from(seen.values())
})

// Tasks after client-side filters applied
const tasks = computed(() => {
	const from = dueDateFrom.value ? new Date(dueDateFrom.value).getTime() : null
	const to = dueDateTo.value ? new Date(dueDateTo.value).getTime() + 86_400_000 - 1 : null
	return allTasks.value.filter(t => {
		if (selectedProjectId.value !== null && t.projectId !== selectedProjectId.value) {
			return false
		}
		if (selectedAssigneeId.value !== null) {
			const hasAssignee = (t.assignees ?? []).some(a => a.id === selectedAssigneeId.value)
			if (!hasAssignee) return false
		}
		if (from !== null || to !== null) {
			if (!t.dueDate) return false
			const due = new Date(t.dueDate).getTime()
			if (from !== null && due < from) return false
			if (to !== null && due > to) return false
		}
		return true
	})
})

const activeFilterCount = computed(() =>
	(selectedProjectId.value !== null ? 1 : 0) +
	(selectedAssigneeId.value !== null ? 1 : 0) +
	(dueDateFrom.value || dueDateTo.value ? 1 : 0),
)

function clearFilters() {
	selectedProjectId.value = null
	selectedAssigneeId.value = null
	dueDateFrom.value = ''
	dueDateTo.value = ''
}

async function loadTasks() {
	isLoading.value = true
	try {
		const svc = new TaskCollectionService()
		const filterStr = currentFilter.value === 'completed'
			? 'done = true'
			: 'done = false'
		const result = await svc.getAll(new TaskModel({}), {
			sort_by: ['due_date', 'id'],
			order_by: ['asc', 'desc'],
			filter: filterStr,
			filter_include_nulls: true,
			s: '',
			per_page: 100,
		})
		allTasks.value = result
	} catch (e) {
		console.error('Failed to load tasks', e)
	} finally {
		isLoading.value = false
	}
}

function assigneeName(user: IUser): string {
	return user.name?.trim() || user.username
}

function projectName(projectId: number): string {
	return projectStore.projects[projectId]?.title ?? '—'
}

function formatDueDate(date: Date | null): string {
	if (!date) return '—'
	return formatDate(date, 'MMM D, YYYY')
}

function priorityLabel(priority: number): string {
	return PRIORITY_LABELS[priority] ?? 'Low'
}

function priorityClass(priority: number): string {
	if (priority >= 4) return 'tbl-priority tbl-priority--critical'
	if (priority === 3) return 'tbl-priority tbl-priority--high'
	if (priority === 2) return 'tbl-priority tbl-priority--med'
	return 'tbl-priority tbl-priority--low'
}

function goToTask(task: ITask) {
	router.push({name: 'task.detail', params: {id: task.id}})
}

function setFilter(f: 'today' | 'completed') {
	// Reset client-side filters when switching tab
	selectedProjectId.value = null
	selectedAssigneeId.value = null
	dueDateFrom.value = ''
	dueDateTo.value = ''
	router.push({query: {filter: f}})
}

onMounted(loadTasks)
watch(currentFilter, loadTasks)
</script>

<template>
	<div class="tbl-page">
		<!-- Header -->
		<div class="tbl-header">
			<div>
				<h1 class="tbl-title">
					{{ pageTitle }}
				</h1>
				<p class="tbl-subtitle">
					{{ pageDescription }}
				</p>
			</div>
			<!-- Status tabs -->
			<div class="tbl-tabs">
				<button
					class="tbl-tab"
					:class="{'tbl-tab--active': currentFilter === 'today'}"
					@click="setFilter('today')"
				>
					Today's Tasks
				</button>
				<button
					class="tbl-tab"
					:class="{'tbl-tab--active': currentFilter === 'completed'}"
					@click="setFilter('completed')"
				>
					Completed Tasks
				</button>
			</div>
		</div>

		<!-- Filter bar -->
		<div class="tbl-filterbar">
			<!-- Project filter -->
			<div class="tbl-filter-group">
				<label
					class="tbl-filter-label"
					for="filter-project"
				>Project</label>
				<select
					id="filter-project"
					v-model="selectedProjectId"
					class="tbl-select"
				>
					<option :value="null">
						All projects
					</option>
					<option
						v-for="p in availableProjects"
						:key="p.id"
						:value="p.id"
					>
						{{ p.title }}
					</option>
				</select>
			</div>

			<!-- Assignee filter -->
			<div class="tbl-filter-group">
				<label
					class="tbl-filter-label"
					for="filter-assignee"
				>Assignee</label>
				<select
					id="filter-assignee"
					v-model="selectedAssigneeId"
					class="tbl-select"
				>
					<option :value="null">
						All assignees
					</option>
					<option
						v-for="u in availableAssignees"
						:key="u.id"
						:value="u.id"
					>
						{{ assigneeName(u) }}
					</option>
				</select>
			</div>

			<!-- Due date range filter -->
			<div class="tbl-filter-group">
				<label class="tbl-filter-label">Due Date</label>
				<div class="tbl-date-range">
					<input
						v-model="dueDateFrom"
						type="date"
						class="tbl-date-input"
						title="From"
					>
					<span class="tbl-date-sep">—</span>
					<input
						v-model="dueDateTo"
						type="date"
						class="tbl-date-input"
						title="To"
					>
				</div>
			</div>

			<!-- Clear filters -->
			<button
				v-if="activeFilterCount > 0"
				class="tbl-clear-btn"
				@click="clearFilters"
			>
				Clear filters
				<span class="tbl-filter-count">{{ activeFilterCount }}</span>
			</button>

			<!-- Result count -->
			<span class="tbl-result-count">
				{{ tasks.length }} task{{ tasks.length !== 1 ? 's' : '' }}
			</span>
		</div>

		<!-- Loading state -->
		<div
			v-if="isLoading"
			class="tbl-loading"
		>
			Loading tasks...
		</div>

		<!-- Empty state -->
		<div
			v-else-if="tasks.length === 0"
			class="tbl-empty"
		>
			<p class="tbl-empty-title">
				No tasks found
			</p>
			<p class="tbl-empty-sub">
				{{ activeFilterCount > 0 ? 'Try adjusting your filters.' : currentFilter === 'completed' ? 'You have not completed any tasks yet.' : 'Nothing to do — enjoy your day!' }}
			</p>
		</div>

		<!-- Table -->
		<div
			v-else
			class="tbl-wrap"
		>
			<table class="tbl">
				<thead>
					<tr>
						<th class="tbl-th tbl-th--status">
							Status
						</th>
						<th class="tbl-th tbl-th--title">
							Title
						</th>
						<th class="tbl-th">
							Project
						</th>
						<th class="tbl-th">
							Assignees
						</th>
						<th class="tbl-th">
							Priority
						</th>
						<th class="tbl-th">
							Due Date
						</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="task in tasks"
						:key="task.id"
						class="tbl-row"
						@click="goToTask(task)"
					>
						<td class="tbl-td">
							<span
								class="tbl-status"
								:class="task.done ? 'tbl-status--done' : 'tbl-status--pending'"
							>
								{{ task.done ? 'Done' : 'Pending' }}
							</span>
						</td>
						<td class="tbl-td tbl-td--title">
							{{ task.title }}
						</td>
						<td class="tbl-td tbl-td--muted">
							{{ projectName(task.projectId) }}
						</td>
						<td class="tbl-td">
							<span
								v-if="task.assignees && task.assignees.length"
								class="tbl-assignees"
							>
								<span
									v-for="user in task.assignees"
									:key="user.id"
									class="tbl-assignee-chip"
								>{{ assigneeName(user) }}</span>
							</span>
							<span
								v-else
								class="tbl-td--muted"
							>—</span>
						</td>
						<td class="tbl-td">
							<span :class="priorityClass(task.priority)">
								{{ priorityLabel(task.priority) }}
							</span>
						</td>
						<td class="tbl-td tbl-td--muted">
							{{ formatDueDate(task.dueDate) }}
						</td>
					</tr>
				</tbody>
			</table>
		</div>
	</div>
</template>

<style scoped>
.tbl-page {
	padding: 32px;
	padding-top: calc(4rem + 32px);
	min-height: 100%;
	background: #0e0f13;
	color: #f0ede8;
	font-family: 'DM Sans', sans-serif;
}

.tbl-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	flex-wrap: wrap;
	gap: 16px;
	margin-bottom: 20px;
}

.tbl-title {
	font-size: 24px;
	font-weight: 600;
	color: #f0ede8;
	margin: 0 0 4px;
	border: none;
	padding: 0;
}

.tbl-subtitle {
	font-size: 14px;
	color: #8a8a9a;
	margin: 0;
}

/* Status tabs */
.tbl-tabs {
	display: flex;
	gap: 8px;
	background: #13141a;
	padding: 4px;
	border-radius: 8px;
	border: 1px solid #1e1f28;
}

.tbl-tab {
	padding: 6px 16px;
	border-radius: 6px;
	font-size: 13px;
	font-weight: 500;
	border: none;
	cursor: pointer;
	color: #8a8a9a;
	background: transparent;
	transition: background 0.15s, color 0.15s;
}

.tbl-tab:hover {
	color: #f0ede8;
	background: #1e1f28;
}

.tbl-tab--active {
	background: #6c63f5 !important;
	color: #fff !important;
}

/* Filter bar */
.tbl-filterbar {
	display: flex;
	align-items: flex-end;
	flex-wrap: wrap;
	gap: 12px;
	margin-bottom: 20px;
	padding: 14px 16px;
	background: #13141a;
	border: 1px solid #1e1f28;
	border-radius: 10px;
}

.tbl-filter-group {
	display: flex;
	flex-direction: column;
	gap: 4px;
}

.tbl-filter-label {
	font-size: 11px;
	font-weight: 600;
	text-transform: uppercase;
	letter-spacing: 0.05em;
	color: #8a8a9a;
}

.tbl-select {
	appearance: none;
	background: #1a1b24;
	border: 1px solid #2a2b35;
	border-radius: 6px;
	color: #f0ede8;
	font-size: 13px;
	padding: 6px 28px 6px 10px;
	cursor: pointer;
	min-width: 160px;
	background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%238a8a9a' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
	background-repeat: no-repeat;
	background-position: right 8px center;
	transition: border-color 0.15s;
}

.tbl-select:focus {
	outline: none;
	border-color: #6c63f5;
}

.tbl-select option {
	background: #1a1b24;
}

/* Due date range */
.tbl-date-range {
	display: flex;
	align-items: center;
	gap: 6px;
}

.tbl-date-input {
	appearance: none;
	background: #1a1b24;
	border: 1px solid #2a2b35;
	border-radius: 6px;
	color: #f0ede8;
	font-size: 13px;
	padding: 6px 10px;
	cursor: pointer;
	width: 136px;
	transition: border-color 0.15s;
	color-scheme: dark;
}

.tbl-date-input:focus {
	outline: none;
	border-color: #6c63f5;
}

.tbl-date-sep {
	color: #8a8a9a;
	font-size: 13px;
}

.tbl-clear-btn {
	display: flex;
	align-items: center;
	gap: 6px;
	padding: 6px 14px;
	border-radius: 6px;
	border: 1px solid #2a2b35;
	background: transparent;
	color: #8a8a9a;
	font-size: 13px;
	cursor: pointer;
	transition: border-color 0.15s, color 0.15s;
	align-self: flex-end;
}

.tbl-clear-btn:hover {
	border-color: #6c63f5;
	color: #f0ede8;
}

.tbl-filter-count {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	width: 18px;
	height: 18px;
	border-radius: 99px;
	background: #6c63f5;
	color: #fff;
	font-size: 11px;
	font-weight: 600;
}

.tbl-result-count {
	margin-left: auto;
	font-size: 13px;
	color: #8a8a9a;
	align-self: flex-end;
}

/* Loading / Empty */
.tbl-loading,
.tbl-empty {
	text-align: center;
	padding: 64px 16px;
}

.tbl-empty-title {
	font-size: 18px;
	font-weight: 500;
	color: #f0ede8;
	margin: 0 0 8px;
}

.tbl-empty-sub {
	font-size: 14px;
	color: #8a8a9a;
	margin: 0;
}

/* Table */
.tbl-wrap {
	overflow-x: auto;
	border-radius: 10px;
	border: 1px solid #1e1f28;
}

.tbl {
	width: 100%;
	border-collapse: collapse;
	background: #13141a;
	font-size: 14px;
}

.tbl-th {
	padding: 12px 16px;
	text-align: left;
	font-size: 12px;
	font-weight: 600;
	text-transform: uppercase;
	letter-spacing: 0.05em;
	color: #8a8a9a;
	background: #13141a;
	border-bottom: 1px solid #1e1f28;
	white-space: nowrap;
}

.tbl-th--title {
	width: 35%;
}

.tbl-row {
	border-bottom: 1px solid #1e1f28;
	cursor: pointer;
	transition: background 0.12s;
}

.tbl-row:last-child {
	border-bottom: none;
}

.tbl-row:hover {
	background: #181920;
}

.tbl-td {
	padding: 12px 16px;
	vertical-align: middle;
	color: #f0ede8;
}

.tbl-td--title {
	font-weight: 500;
}

.tbl-td--muted {
	color: #8a8a9a;
}

/* Assignees */
.tbl-assignees {
	display: flex;
	flex-wrap: wrap;
	gap: 4px;
}

.tbl-assignee-chip {
	display: inline-block;
	padding: 2px 8px;
	border-radius: 99px;
	background: #1e1f28;
	color: #c4b5fd;
	font-size: 12px;
	font-weight: 500;
	white-space: nowrap;
}

/* Status badge */
.tbl-status {
	display: inline-block;
	padding: 2px 10px;
	border-radius: 99px;
	font-size: 12px;
	font-weight: 500;
	white-space: nowrap;
}

.tbl-status--done {
	background: #0d3024;
	color: #10b981;
}

.tbl-status--pending {
	background: #1e1b3a;
	color: #a78bfa;
}

/* Priority badge */
.tbl-priority {
	display: inline-block;
	padding: 2px 10px;
	border-radius: 99px;
	font-size: 12px;
	font-weight: 500;
	white-space: nowrap;
}

.tbl-priority--low {
	background: #1a1a2e;
	color: #8a8a9a;
}

.tbl-priority--med {
	background: #1e2a10;
	color: #86efac;
}

.tbl-priority--high {
	background: #2a1a0e;
	color: #f59e0b;
}

.tbl-priority--critical {
	background: #2a0e0e;
	color: #ef4444;
}
</style>
