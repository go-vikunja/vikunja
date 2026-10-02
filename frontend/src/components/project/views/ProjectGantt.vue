<template>
	<ProjectWrapper
		class="project-gantt"
		:is-loading-project="isLoadingProject"
		:project-id
		:view-id
	>
		<template #default>
			<Card :has-content="false">
				<div class="gantt-options">
					<FormField
						id="range"
						:label="$t('misc.dateRange')"
					>
						<template #default="{ id }">
							<DateRangeInput
								:id="id"
								v-model="dateRange"
								:placeholder="$t('misc.dateRange')"
							/>
						</template>
					</FormField>
					<div
						v-if="!hasDefaultFilters"
						class="field"
					>
						<span
							class="label"
							aria-hidden="true"
						>Reset</span>
						<div class="control">
							<XButton @click="setDefaultFilters">
								Reset
							</XButton>
						</div>
					</div>
					<FancyCheckbox
						v-model="filters.showTasksWithoutDates"
						is-block
					>
						{{ $t('task.show.noDates') }}
					</FancyCheckbox>
				</div>
				<div class="gantt-toolbar">
					<div
						class="gantt-zoom"
						role="group"
						:aria-label="$t('project.gantt.zoom.label')"
					>
						<XButton
							v-for="level in GANTT_ZOOMS"
							:key="level"
							size="small"
							:variant="zoom === level ? 'primary' : 'secondary'"
							:aria-pressed="zoom === level"
							@click="zoom = level"
						>
							{{ $t(`project.gantt.zoom.${level}`) }}
						</XButton>
					</div>
					<FancyCheckbox v-model="showPane">
						{{ $t('project.gantt.showTable') }}
					</FancyCheckbox>
					<FancyCheckbox v-model="showCritical">
						{{ $t('project.gantt.criticalPath') }}
					</FancyCheckbox>
					<FancyCheckbox
						v-model="showBaseline"
						:disabled="baselineMap.size === 0"
					>
						{{ $t('project.gantt.showBaseline') }}
					</FancyCheckbox>
					<XButton
						v-if="canWrite"
						size="small"
						variant="secondary"
						:loading="saveBaseline.isPending.value"
						@click="onSaveBaseline"
					>
						{{ $t('project.gantt.setBaseline') }}
					</XButton>
					<XButton
						v-if="canWrite && baselineMap.size > 0"
						size="small"
						variant="tertiary"
						:loading="clearBaseline.isPending.value"
						@click="onClearBaseline"
					>
						{{ $t('project.gantt.clearBaseline') }}
					</XButton>
					<ViewExportMenu
						kind="gantt"
						:project-id="projectId"
						:view-id="viewId"
						:params="ganttExportParams"
						:build-sheets="buildGanttExportSheets"
						:print-query="ganttPrintQuery"
					/>
				</div>
				<p
					v-if="showCritical"
					class="help"
				>
					{{ $t('project.gantt.criticalPathHelp') }}
				</p>
				<p
					v-if="baselineSavedAt"
					class="help"
				>
					{{ $t('project.gantt.baselineSavedAt', {date: formatDateSince(baselineSavedAt)}) }}
				</p>
			</Card>

			<div class="gantt-chart-container">
				<Card
					:has-content="false"
					:padding="false"
					class="has-overflow"
				>
					<GanttChart
						:filters="filters"
						:tasks="tasks"
						:is-loading="isLoading"
						:default-task-start-date="defaultTaskStartDate"
						:default-task-end-date="defaultTaskEndDate"
						:zoom="zoom"
						:editable="canWrite"
						:show-pane="showPane"
						:pane-width="paneWidth"
						:show-critical="showCritical"
						:show-baseline="showBaseline"
						:baseline="baselineMap"
						@update:task="updateTask"
						@update:paneWidth="paneWidth = $event"
						@changeDates="onChangeDates"
						@createDependency="onCreateDependency"
						@deleteDependency="onDeleteDependency"
					/>
					<TaskForm
						v-if="canWrite"
						@createTask="addGanttTask"
					/>
				</Card>
			</div>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, toRefs} from 'vue'
import type {RouteLocationNormalized} from 'vue-router'
import {useStorage} from '@vueuse/core'
import {useMutation, useQuery} from '@tanstack/vue-query'
import {useI18n} from 'vue-i18n'

import {useCurrentProject} from '@/composables/useCurrentProject'
import {GANTT_ZOOMS, isGanttZoom, type GanttZoom} from '@/helpers/ganttZoom'
import {
	clearBaselineMutation,
	createDependencyMutation,
	deleteDependencyMutation,
	projectBaselineQuery,
	rescheduleTaskMutation,
	saveBaselineMutation,
} from '@/client/queries/gantt'
import {formatDateSince} from '@/helpers/time/formatDate'
import {error, success} from '@/message'
import ViewExportMenu from '@/components/project/partials/ViewExportMenu.vue'
import {buildGanttSheets} from '@/helpers/export/ganttExport'
import dayjs from 'dayjs'

import DateRangeInput from '@/components/input/DateRangeInput.vue'
import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import TaskForm from '@/components/tasks/TaskForm.vue'
import FormField from '@/components/input/FormField.vue'

import GanttChart from '@/components/gantt/GanttChart.vue'
import {useGanttFilters} from '../../../views/project/helpers/useGanttFilters'
import {PERMISSIONS} from '@/constants/permissions'

import type {DateISO} from '@/types/DateISO'
import type {Task as ITask} from '@/client/generated'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	route: RouteLocationNormalized
	viewId: number
}>()


const {currentProject} = useCurrentProject()
const canWrite = computed(() =>
	typeof currentProject.value?.max_permission === 'number' &&
	currentProject.value.max_permission > PERMISSIONS.READ,
)

const {route, projectId, viewId} = toRefs(props)
const {
	filters,
	hasDefaultFilters,
	setDefaultFilters,
	tasks,
	isLoading,
	addTask,
	updateTask,
} = useGanttFilters(route, projectId, viewId)

const {t} = useI18n({useScope: 'global'})

// ---- View options, remembered per browser ----
const storedZoom = useStorage<string>('ganttZoom', 'fit')
const zoom = computed<GanttZoom>({
	get: () => isGanttZoom(storedZoom.value) ? storedZoom.value : 'fit',
	set: value => { storedZoom.value = value },
})
const showPane = useStorage('ganttShowPane', true)
const paneWidth = useStorage('ganttPaneWidth', 520)
const showCritical = useStorage('ganttShowCritical', false)
const showBaseline = useStorage('ganttShowBaseline', false)

// ---- Baseline ----
const baselineQuery = useQuery(computed(() => projectBaselineQuery(() => projectId.value)))
const baselineMap = computed(() => new Map(
	(baselineQuery.data.value?.items ?? []).map(item => [
		item.task_id,
		{
			start: item.start_date ? new Date(item.start_date) : null,
			end: item.end_date ? new Date(item.end_date) : null,
		},
	]),
))
const baselineSavedAt = computed(() => baselineQuery.data.value?.saved_at ?? null)
const saveBaseline = useMutation(saveBaselineMutation())
const clearBaseline = useMutation(clearBaselineMutation())

async function onSaveBaseline() {
	try {
		const {tasks: count} = await saveBaseline.mutateAsync(projectId.value)
		showBaseline.value = true
		success({message: t('project.gantt.baselineSaved', {count})})
	} catch (e) {
		error(e)
	}
}

async function onClearBaseline() {
	try {
		await clearBaseline.mutateAsync(projectId.value)
		showBaseline.value = false
	} catch (e) {
		error(e)
	}
}

// ---- Export ----
// The sheets show what the chart shows: its tasks, zoom and date range.
const ganttExportParams = computed(() => ({filter: '', q: ''}))
const ganttPrintQuery = computed(() => ({
	zoom: zoom.value,
	dateFrom: filters.value.dateFrom,
	dateTo: filters.value.dateTo,
	critical: String(showCritical.value),
	withoutDates: String(filters.value.showTasksWithoutDates),
}))

function formatTimelineUnit(date: Date, level: GanttZoom): string {
	switch (level) {
		case 'fit':
		case 'day':
			return dayjs(date).format('D')
		case 'week':
			return dayjs(date).format('D MMM')
		case 'month':
			return dayjs(date).format('MMM YY')
		case 'quarter':
			return `Q${Math.floor(date.getMonth() / 3) + 1} ${date.getFullYear()}`
	}
}

async function buildGanttExportSheets() {
	const result = buildGanttSheets({
		tasks: tasks.value,
		zoom: zoom.value,
		from: new Date(filters.value.dateFrom),
		to: new Date(filters.value.dateTo),
		showCritical: showCritical.value,
		formatUnit: formatTimelineUnit,
		labels: {
			tasksSheet: t('export.sheetTasks'),
			timelineSheet: t('export.sheetTimeline'),
			wbs: t('project.gantt.pane.wbs'),
			name: t('project.gantt.pane.name'),
			duration: t('project.gantt.pane.duration'),
			start: t('project.gantt.pane.start'),
			finish: t('project.gantt.pane.finish'),
			percent: t('project.gantt.pane.percent'),
			predecessors: t('project.gantt.pane.predecessors'),
			assignees: t('project.gantt.pane.assignees'),
			critical: t('project.gantt.criticalPath'),
			truncated: t('export.timelineTruncated'),
		},
	})
	return result
}

// ---- Dates and dependencies ----
const reschedule = useMutation(rescheduleTaskMutation())
const createDependency = useMutation(createDependencyMutation())
const deleteDependency = useMutation(deleteDependencyMutation())

function toIso(date: Date | null): string | null {
	return date ? date.toISOString() : null
}

// Moves a task. The server pushes the tasks that depend on it later and answers which ones, so
// the toast can say so and offer to put everything back.
async function onChangeDates(id: number, start: Date | null, end: Date | null) {
	const before = tasks.value.get(id)
	const previous = before
		? {start: before.start_date && !before.start_date.startsWith('0001') ? before.start_date : null,
			end: before.end_date && !before.end_date.startsWith('0001') ? before.end_date : null}
		: null

	try {
		const result = await reschedule.mutateAsync({id, start_date: toIso(start), end_date: toIso(end)})
		if (result.changed.length === 0 && result.skipped.length === 0) {
			return
		}
		const parts = [t('project.gantt.rescheduled', {count: result.changed.length})]
		if (result.skipped.length > 0) {
			parts.push(t('project.gantt.rescheduleSkipped', {count: result.skipped.length}))
		}
		success({message: parts.join(' ')}, previous && result.changed.length > 0 ? [{
			title: t('project.gantt.undo'),
			callback: () => {
				// Undo puts the moved task back. Its dependents were pushed by the move and stay where
				// they are: moving a task earlier never pulls them, see the API description.
				reschedule.mutate({id, start_date: previous.start, end_date: previous.end})
			},
		}] : [])
	} catch (e) {
		error(e)
	}
}

async function onCreateDependency(predecessorId: number, successorId: number) {
	try {
		await createDependency.mutateAsync({predecessorId, successorId})
		// The new dependency may already be violated: let the server settle it from the predecessor.
		const predecessor = tasks.value.get(predecessorId)
		const end = predecessor?.end_date && !predecessor.end_date.startsWith('0001') ? predecessor.end_date : null
		if (predecessor && end) {
			const start = predecessor.start_date && !predecessor.start_date.startsWith('0001') ? predecessor.start_date : null
			await onChangeDates(predecessorId, start ? new Date(start) : null, new Date(end))
		}
	} catch (e) {
		error(e)
	}
}

async function onDeleteDependency(predecessorId: number, successorId: number) {
	try {
		await deleteDependency.mutateAsync({predecessorId, successorId})
	} catch (e) {
		error(e)
	}
}

const DEFAULT_DATE_RANGE_DAYS = 7

const today = new Date()
const defaultTaskStartDate: DateISO = new Date(today.setHours(0, 0, 0, 0)).toISOString()
const defaultTaskEndDate: DateISO = new Date(new Date(
	today.getFullYear(),
	today.getMonth(),
	today.getDate() + DEFAULT_DATE_RANGE_DAYS,
).setHours(23, 59, 0, 0)).toISOString()

async function addGanttTask(title: ITask['title']) {
	return await addTask({
		title,
		project_id: filters.value.projectId,
		start_date: defaultTaskStartDate,
		end_date: defaultTaskEndDate,
	})
}

const dateRange = computed({
	get: () => ({
		start: new Date(filters.value.dateFrom),
		end: new Date(filters.value.dateTo),
	}),
	set({start, end}: {start: Date, end: Date}) {
		Object.assign(filters.value, {dateFrom: start.toISOString(), dateTo: end.toISOString()})
	},
})
</script>

<style lang="scss" scoped>
.gantt-toolbar {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.75rem 1rem;
	margin-block-end: 0.5rem;
}

.gantt-zoom {
	display: inline-flex;
	gap: 0.25rem;
}

.gantt-chart-container {
	padding-block-end: 1rem;
	position: relative;
	z-index: 0;
}

.gantt-options {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-block-end: 1rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

:global(.link-share-view:not(.has-background)) .gantt-options {
	border: none;
	box-shadow: none;

	.card-content {
		padding: .5rem;
	}
}

.field {
	margin-block-end: 0;
	inline-size: 33%;

	&:not(:last-child) {
		padding-inline-end: .5rem;
	}

	@media screen and (max-width: $tablet) {
		inline-size: 100%;
		max-inline-size: 100%;
		margin-block-start: .5rem;
		padding-inline-end: 0 !important;
	}

	&, .input {
		font-size: .8rem;
	}

	.select,
	.select select {
		block-size: auto;
		inline-size: 100%;
		font-size: .8rem;
	}

	.label {
		font-size: .9rem;
	}
}
</style>
