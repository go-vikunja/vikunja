<template>
	<Loading class="gantt-container" v-if="taskService.loading || taskCollectionService.loading"/>
	<div class="gantt-container" v-else>
		<GGanttChart
			:chart-start="`${dateFrom} 00:00`"
			:chart-end="`${dateTo} 23:59`"
			:precision="PRECISION"
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
					:class="{'today': dayIsToday(label)}">
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
</template>

<script setup lang="ts">
import {computed, ref, watch, watchEffect, shallowReactive, type PropType} from 'vue'
import {useRouter} from 'vue-router'
import {format, parse} from 'date-fns'

import TaskCollectionService from '@/services/taskCollection'
import TaskService from '@/services/task'
import TaskModel, { getHexColor } from '@/models/task'

import type ListModel from '@/models/list'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import {RIGHTS} from '@/constants/rights'

import {
	extendDayjs,
	GGanttChart,
	GGanttRow,
	type GanttBarObject,
} from '@infectoone/vue-ganttastic'

import Loading from '@/components/misc/loading.vue'
import TaskForm from '@/components/tasks/TaskForm.vue'

import {useBaseStore} from '@/stores/base'

extendDayjs()

const PRECISION = 'day' as const
const DATE_FORMAT = 'yyyy-LL-dd HH:mm'

const baseStore = useBaseStore()
const router = useRouter()

const props = defineProps({
	listId: {
		type: Number as PropType<ListModel['id']>,
		required: true,
	},
	dateFrom: {
		type: String as PropType<any>,
		required: true,
	},
	dateTo: {
		type: String as PropType<any>,
		required: true,
	},
	showTasksWithoutDates: {
		type: Boolean,
		default: false,
	},
})

const taskCollectionService = shallowReactive(new TaskCollectionService())
const taskService = shallowReactive(new TaskService())

const dateFromDate = computed(() => parse(props.dateFrom, 'yyyy-LL-dd', new Date()))
const dateToDate = computed(() => parse(props.dateTo, 'yyyy-LL-dd', new Date()))

const DAY_WIDTH_PIXELS = 30
const ganttChartWidth = computed(() => {
	const dateDiff = Math.floor((dateToDate.value - dateFromDate.value) / (1000 * 60 * 60 * 24))

	return dateDiff * DAY_WIDTH_PIXELS
})

const canWrite = computed(() => baseStore.currentList.maxRight > RIGHTS.READ)

const tasks = ref<Map<TaskModel['id'], TaskModel>>(new Map())
const ganttBars = ref<GanttBarObject[][]>([])

watch(
	tasks,
	// We need a "real" ref object for the gantt bars to instantly update the tasks when they are dragged on the chart.
	// A computed won't work directly.
	// function mapGanttBars() {
	() => {
		ganttBars.value = []
		tasks.value.forEach(t => ganttBars.value.push(transformTaskToGanttBar(t)))
	},
	{deep: true}
)

const defaultStartDate = format(new Date(), DATE_FORMAT)
const defaultEndDate = format(new Date((new Date()).setDate((new Date()).getDate() + 7)), DATE_FORMAT)

function transformTaskToGanttBar(t: TaskModel) {
	const black = 'var(--grey-800)'
	return [{
		startDate: t.startDate ? format(t.startDate, DATE_FORMAT) : defaultStartDate,
		endDate: t.endDate ? format(t.endDate, DATE_FORMAT) : defaultEndDate,
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

async function getAllTasks(params: GetAllTasksParams, page = 1): Promise<TaskModel[]> {
	const tasks = await taskCollectionService.getAll({listId: props.listId}, params, page) as TaskModel[]
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

	const params = {
		sort_by: ['start_date', 'done', 'id'],
		order_by: ['asc', 'asc', 'desc'],
		filter_by: ['start_date', 'start_date'],
		filter_comparator: ['greater_equals', 'less_equals'],
		filter_value: [dateFrom, dateTo],
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

async function createTask(title: TaskModel['title']) {
	const newTask = await taskService.create(new TaskModel({
		title,
		listId: props.listId,
		startDate: defaultStartDate,
		endDate: defaultEndDate,
	}))
	tasks.value.set(newTask.id, newTask)

	return newTask
}

async function updateTask(e) {
	const task = tasks.value.get(e.bar.ganttBarConfig.id)

	if (!task) return

	task.startDate = e.bar.startDate
	task.endDate = e.bar.endDate
	const updatedTask = await taskService.update(task)
	ganttBars.value.map(gantBar => {
		return gantBar[0].ganttBarConfig.id === task.id
			? transformTaskToGanttBar(updatedTask)
			: gantBar
	})
}

function openTask(e) {
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
	display: none;
}

.g-upper-timeunit, .g-timeunit {
	background: var(--white);
	font-family: $vikunja-font;
}

.g-upper-timeunit {
	font-weight: bold;
	border-right: 1px solid var(--grey-200);
	padding: .5rem 0;
}

.g-timeunit .timeunit-wrapper {
	padding: 0.5rem 0;
	font-size: 1rem;
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
	height: auto;
	box-shadow: none;
}

.g-gantt-row > .g-gantt-row-bars-container {
	border-bottom: none;
	border-top: none;
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
		width: 6px;
		height: 75%;
		opacity: .75;
		border-radius: $radius;
		margin-top: 4px;
	}
}
</style>