<template>
	<div>
		<Loading
			v-if="taskCollectionService.loading || dayjsLanguageLoading"
			class="gantt-container"
		/>
		<div class="gantt-container" v-else>
			<GGanttChart
				:date-format="DAYJS_ISO_DATE_FORMAT"
				:chart-start="isoToKebabDate(props.dateFrom)"
				:chart-end="isoToKebabDate(props.dateTo)"
				precision="day"
				bar-start="startDate"
				bar-end="endDate"
				:grid="true"
				@dragend-bar="updateTask"
				@dblclick-bar="openTask"
				:width="ganttChartWidth + 'px'"
			>
				<template #timeunit="{label, value}">
					<div
						class="timeunit-wrapper"
						:class="{'today': dayIsToday(label)}"
					>
						<span>{{ value }}</span>
						<span class="weekday">
							{{ weekdayFromTimeLabel(label) }}
						</span>
					</div>
				</template>
			<GGanttRow
				v-for="(bar, k) in ganttBars"
				:key="k"
				label=""
				:bars="bar"
			/>
		</GGanttChart>
		</div>
		<TaskForm v-if="canWrite" @create-task="createTask" />
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, watchEffect, shallowReactive} from 'vue'
import {useRouter} from 'vue-router'
import {format, parse} from 'date-fns'
import dayjs from 'dayjs'
import isToday from 'dayjs/plugin/isToday'
import cloneDeep from 'lodash.clonedeep'

import {useDayjsLanguageSync} from '@/i18n'
import TaskCollectionService from '@/services/taskCollection'
import TaskService from '@/services/task'
import TaskModel, {getHexColor} from '@/models/task'

import {colorIsDark} from '@/helpers/color/colorIsDark'
import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {parseKebabDate} from '@/helpers/time/parseKebabDate'
import {RIGHTS} from '@/constants/rights'

import type {ITask} from '@/modelTypes/ITask'
import type {IList} from '@/modelTypes/IList'

import {
	extendDayjs,
	GGanttChart,
	GGanttRow,
	type GanttBarObject,
} from '@infectoone/vue-ganttastic'

import Loading from '@/components/misc/loading.vue'
import TaskForm from '@/components/tasks/TaskForm.vue'

import {useBaseStore} from '@/stores/base'
import {error, success} from '@/message'

export interface GanttChartProps {
	listId: IList['id']
	showTasksWithoutDates: boolean
	dateFrom: string,
	dateTo: string,
}

const DAYJS_ISO_DATE_FORMAT = 'YYYY-MM-DD'

const props = withDefaults(defineProps<GanttChartProps>(), {
	showTasksWithoutDates: false,
})

// setup dayjs for vue-ganttastic
const dayjsLanguageLoading = useDayjsLanguageSync(dayjs)
dayjs.extend(isToday)
extendDayjs()

const baseStore = useBaseStore()
const router = useRouter()

const taskCollectionService = shallowReactive(new TaskCollectionService())
const taskService = shallowReactive(new TaskService())

const dateFromDate = computed(() => new Date(new Date(props.dateFrom).setHours(0,0,0,0)))
const dateToDate = computed(() => new Date(new Date(props.dateTo).setHours(23,59,0,0)))

const DAY_WIDTH_PIXELS = 30
const ganttChartWidth = computed(() => {
	const dateDiff = Math.floor((dateToDate.value.valueOf() - dateFromDate.value.valueOf()) / (1000 * 60 * 60 * 24))

	return dateDiff * DAY_WIDTH_PIXELS
})

const canWrite = computed(() => baseStore.currentList.maxRight > RIGHTS.READ)

const tasks = ref<Map<ITask['id'], ITask>>(new Map())
const ganttBars = ref<GanttBarObject[][]>([])

watch(
	tasks,
	() => {
		ganttBars.value = []
		tasks.value.forEach(t => ganttBars.value.push(transformTaskToGanttBar(t)))
	},
	{deep: true},
)

const today = new Date(new Date(props.dateFrom).setHours(0,0,0,0))
const defaultTaskStartDate = new Date(today)
const defaultTaskEndDate = new Date(today.getFullYear(), today.getMonth(), today.getDate() + 7, 23,59,0,0)

function transformTaskToGanttBar(t: ITask) {
	const black = 'var(--grey-800)'
	return [{
		startDate: isoToKebabDate(t.startDate ? t.startDate.toISOString() : defaultTaskStartDate.toISOString()),
		endDate: isoToKebabDate(t.endDate ? t.endDate.toISOString() : defaultTaskEndDate.toISOString()),
		ganttBarConfig: {
			id: String(t.id),
			label: t.title,
			hasHandles: true,
			style: {
				color: t.startDate ? (colorIsDark(getHexColor(t.hexColor)) ? black : 'white') : black,
				backgroundColor: t.startDate ? getHexColor(t.hexColor) : 'var(--grey-100)',
				border: t.startDate ? '' : '2px dashed var(--grey-300)',
				'text-decoration': t.done ? 'line-through' : null,
			},
		},
	} as GanttBarObject]
}

// FIXME: unite with other filter params types
interface GetAllTasksParams {
	sort_by: ('start_date' | 'done' | 'id')[],
	order_by: ('asc' | 'asc' | 'desc')[],
	filter_by: 'start_date'[],
	filter_comparator: ('greater_equals' | 'less_equals')[],
	filter_value: [string, string] // [dateFrom, dateTo],
	filter_concat: 'and',
	filter_include_nulls: boolean,
}

async function getAllTasks(params: GetAllTasksParams, page = 1): Promise<ITask[]> {
	const tasks = await taskCollectionService.getAll({listId: props.listId}, params, page) as ITask[]
	if (page < taskCollectionService.totalPages) {
		const nextTasks = await getAllTasks(params, page + 1)
		return tasks.concat(nextTasks)
	}
	return tasks
}

async function loadTasks({
	dateTo,
	dateFrom,
	showTasksWithoutDates,
}: {
	dateTo: string;
	dateFrom: string;
	showTasksWithoutDates: boolean;
}) {
	tasks.value = new Map()

	const params: GetAllTasksParams = {
		sort_by: ['start_date', 'done', 'id'],
		order_by: ['asc', 'asc', 'desc'],
		filter_by: ['start_date', 'start_date'],
		filter_comparator: ['greater_equals', 'less_equals'],
		filter_value: [isoToKebabDate(dateFrom), isoToKebabDate(dateTo)],
		filter_concat: 'and',
		filter_include_nulls: showTasksWithoutDates,
	}

	const loadedTasks = await getAllTasks(params)
	loadedTasks.forEach(t => tasks.value.set(t.id, t))
}

watchEffect(() => loadTasks({
	dateTo: props.dateTo,
	dateFrom: props.dateFrom,
	showTasksWithoutDates: props.showTasksWithoutDates,
}))

async function createTask(title: ITask['title']) {
	const newTask = await taskService.create(new TaskModel({
		title,
		listId: props.listId,
		startDate: defaultTaskStartDate.toISOString(),
		endDate: defaultTaskEndDate.toISOString(),
	}))
	tasks.value.set(newTask.id, newTask)

	return newTask
}

async function updateTask(e: {
    bar: GanttBarObject;
    e: MouseEvent;
    datetime?: string | undefined;
}) {
	const task = tasks.value.get(Number(e.bar.ganttBarConfig.id))

	if (!task) return

	const oldTask = cloneDeep(task)
	const newTask: ITask = {
		...task,
		startDate: new Date(parseKebabDate(e.bar.startDate).setHours(0,0,0,0)),
		endDate: new Date(parseKebabDate(e.bar.endDate).setHours(23,59,0,0)),
	}

	tasks.value.set(newTask.id, newTask)

	try {	
		const updatedTask = await taskService.update(newTask)
		tasks.value.set(updatedTask.id, updatedTask)
		success('Saved')
	} catch(e: any) {
		error('Something went wrong saving the task')
		tasks.value.set(task.id, oldTask)
	}
}

function openTask(e: {
    bar: GanttBarObject;
    e: MouseEvent;
    datetime?: string | undefined;
}) {
	router.push({
		name: 'task.detail',
		params: {id: e.bar.ganttBarConfig.id},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

function weekdayFromTimeLabel(label: string): string {
	const parsed = parse(label, 'dd.MMM', dateFromDate.value)
	return format(parsed, 'E')
}

function dayIsToday(label: string): boolean {
	const parsed = parse(label, 'dd.MMM', dateFromDate.value)
	const today = new Date()
	return parsed.getDate() === today.getDate() &&
		parsed.getMonth() === today.getMonth() &&
		parsed.getFullYear() === today.getFullYear()
}
</script>

<style scoped lang="scss">
.gantt-container {
	overflow-x: auto;
}
</style>
	

<style lang="scss">
// Not scoped because we need to style the elements inside the gantt chart component
.g-gantt-chart {
	width: 2000px;
}

.g-gantt-row-label {
	display: none !important;
}

.g-upper-timeunit, .g-timeunit {
	background: var(--white) !important;
	font-family: $vikunja-font;
}

.g-upper-timeunit {
	font-weight: bold;
	border-right: 1px solid var(--grey-200);
	padding: .5rem 0;
}

.g-timeunit .timeunit-wrapper {
	padding: 0.5rem 0;
	font-size: 1rem !important;
	display: flex;
	flex-direction: column;
	align-items: center;
	width: 100%;

	&.today {
		background: var(--primary);
		color: var(--white);
		border-radius: 5px 5px 0 0;
		font-weight: bold;
	}

	.weekday {
		font-size: 0.8rem;
	}
}

.g-timeaxis {
	height: auto !important;
	box-shadow: none !important;
}

.g-gantt-row > .g-gantt-row-bars-container {
	border-bottom: none !important;
	border-top: none !important;
}

.g-gantt-row:nth-child(odd) {
	background: hsla(var(--grey-100-hsl), .5);
}

.g-gantt-bar {
	border-radius: $radius * 1.5;
	overflow: visible;
	font-size: .85rem;

	&-handle-left,
	&-handle-right {
		width: 6px !important;
		height: 75% !important;
		opacity: .75 !important;
		border-radius: $radius !important;
		margin-top: 4px;
	}
}
</style>