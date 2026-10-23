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
					<FormField
						id="zoom"
						:label="$t('project.gantt.zoom')"
					>
						<template #default="{ id }">
							<div
								:id="id"
								class="gantt-zoom-controls"
								role="group"
								:aria-label="$t('project.gantt.zoom')"
							>
								<XButton
									variant="secondary"
									:shadow="false"
									icon="minus"
									:aria-label="$t('misc.zoomOut')"
									:aria-disabled="!ganttChart?.canZoomOut"
									@click="ganttChart?.zoomOut()"
								/>
								<XButton
									variant="secondary"
									:shadow="false"
									icon="plus"
									:aria-label="$t('misc.zoomIn')"
									:aria-disabled="!ganttChart?.canZoomIn"
									@click="ganttChart?.zoomIn()"
								/>
								<XButton
									variant="secondary"
									:shadow="false"
									@click="ganttChart?.zoomToFit()"
								>
									{{ $t('project.gantt.zoomToFit') }}
								</XButton>
								<XButton
									variant="secondary"
									:shadow="false"
									@click="goToToday"
								>
									{{ $t('input.datepicker.today') }}
								</XButton>
							</div>
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
				<div
					class="gantt-quick-ranges"
					role="group"
					:aria-label="$t('project.gantt.quickRange')"
				>
					<XButton
						v-for="quickRange in QUICK_RANGES"
						:key="quickRange.key"
						variant="secondary"
						:shadow="false"
						@click="applyQuickRange(quickRange)"
					>
						{{ $t(`project.gantt.${quickRange.key}`) }}
					</XButton>
				</div>
			</Card>

			<div class="gantt-chart-container">
				<Card
					:has-content="false"
					:padding="false"
					class="has-overflow"
				>
					<GanttChart
						ref="ganttChart"
						:filters="filters"
						:tasks="tasks"
						:is-loading="isLoading"
						:default-task-start-date="defaultTaskStartDate"
						:default-task-end-date="defaultTaskEndDate"
						@update:task="updateTask"
						@update:zoom="filters.zoom = $event"
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
import {computed, nextTick, ref, toRefs} from 'vue'
import type {RouteLocationNormalized} from 'vue-router'

import {useCurrentProject} from '@/composables/useCurrentProject'

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
import {GANTT_ZOOM_PRESETS} from '@/helpers/time/ganttTimeline'

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

const ganttChart = ref<InstanceType<typeof GanttChart> | null>(null)

interface QuickRange {
	key: 'thisMonth' | 'thisQuarter' | 'nextThreeMonths' | 'thisYear' | 'nextYear'
	zoom: number
	range: (now: Date) => {start: Date, end: Date}
}

const QUICK_RANGES: QuickRange[] = [
	{
		key: 'thisMonth',
		zoom: GANTT_ZOOM_PRESETS.day,
		range: now => ({
			start: new Date(now.getFullYear(), now.getMonth(), 1),
			end: new Date(now.getFullYear(), now.getMonth() + 1, 0),
		}),
	},
	{
		key: 'thisQuarter',
		zoom: GANTT_ZOOM_PRESETS.week,
		range: now => {
			const quarterStartMonth = now.getMonth() - (now.getMonth() % 3)
			return {
				start: new Date(now.getFullYear(), quarterStartMonth, 1),
				end: new Date(now.getFullYear(), quarterStartMonth + 3, 0),
			}
		},
	},
	{
		key: 'nextThreeMonths',
		zoom: GANTT_ZOOM_PRESETS.week,
		range: now => ({
			start: new Date(now.getFullYear(), now.getMonth(), now.getDate()),
			end: new Date(now.getFullYear(), now.getMonth() + 3, now.getDate()),
		}),
	},
	{
		key: 'thisYear',
		zoom: GANTT_ZOOM_PRESETS.month,
		range: now => ({
			start: new Date(now.getFullYear(), 0, 1),
			end: new Date(now.getFullYear(), 11, 31),
		}),
	},
	{
		key: 'nextYear',
		zoom: GANTT_ZOOM_PRESETS.month,
		range: now => ({
			start: new Date(now.getFullYear() + 1, 0, 1),
			end: new Date(now.getFullYear() + 1, 11, 31),
		}),
	},
]

function applyQuickRange(quickRange: QuickRange) {
	const {start, end} = quickRange.range(new Date())
	Object.assign(filters.value, {
		dateFrom: start.toISOString(),
		dateTo: end.toISOString(),
		zoom: quickRange.zoom,
	})
}

async function goToToday() {
	const today = new Date()
	const from = new Date(filters.value.dateFrom)
	const to = new Date(filters.value.dateTo)

	if (today < from || today > to) {
		const half = (to.valueOf() - from.valueOf()) / 2
		Object.assign(filters.value, {
			dateFrom: new Date(today.valueOf() - half).toISOString(),
			dateTo: new Date(today.valueOf() + half).toISOString(),
		})
		await nextTick()
	}

	ganttChart.value?.scrollToDate(today)
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
.gantt-chart-container {
	padding-block-end: 1rem;
	position: relative;
	z-index: 0;
}

.gantt-options {
	display: flex;
	justify-content: space-between;
	align-items: center;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.gantt-zoom-controls {
	display: flex;
	gap: .25rem;
}

.gantt-quick-ranges {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem;
	padding: 0 1rem 1rem;
	margin-block-end: 1rem;
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
