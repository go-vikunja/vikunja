<template>
	<Loading class="gantt-container" v-if="taskService.loading || taskCollectionService.loading"/>
	<div class="gantt-container" v-else>
		<g-gantt-chart
			:chart-start="`${dateFrom} 00:00`"
			:chart-end="`${dateTo} 23:59`"
			:precision="PRECISION"
			bar-start="startDate"
			bar-end="endDate"
			:grid="true"
			@dragend-bar="updateTask"
			@dblclick-bar="openTask"
			font="inherit"
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
			<g-gantt-row
				v-for="(bar, k) in ganttBars"
				:key="k"
				label=""
				:bars="bar"
			/>
		</g-gantt-chart>
	</div>
	<TaskForm v-if="canWrite" @create-task="createTask" />
</template>

<script setup lang="ts">
import {computed, ref, watchEffect, shallowReactive, type Ref, type PropType} from 'vue'
import TaskCollectionService from '@/services/taskCollection'
import TaskService from '@/services/task'
import {format, parse} from 'date-fns'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import {useStore} from '@/store'
import {RIGHTS} from '@/constants/rights'
import TaskModel from '@/models/task'
import {useRouter} from 'vue-router'
import Loading from '@/components/misc/loading.vue'
import type ListModel from '@/models/list'

// FIXME: these types should be exported from vue-ganttastic
// see: https://github.com/InfectoOne/vue-ganttastic/blob/master/src/models/models.ts

export interface GanttBarConfig {
	id: string,
	label?: string
	hasHandles?: boolean
	immobile?: boolean
	bundle?: string
	pushOnOverlap?: boolean
	dragLimitLeft?: number
	dragLimitRight?: number
	style?: CSSStyleSheet
}

export type GanttBarObject = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any,
  ganttBarConfig: GanttBarConfig
}

export type GGanttChartPropsRefs = {
  chartStart: Ref<string>
  chartEnd: Ref<string>
  precision: Ref<'hour' | 'day' | 'month'>
  barStart: Ref<string>
  barEnd: Ref<string>
  rowHeight: Ref<number>
  dateFormat: Ref<string>
  width: Ref<string>
  hideTimeaxis: Ref<boolean>
  colorScheme: Ref<string>
  grid: Ref<boolean>
  pushOnOverlap: Ref<boolean>
  noOverlap: Ref<boolean>
  gGanttChart: Ref<HTMLElement | null>
  font: Ref<string>
}

const PRECISION = 'day'

const DATE_FORMAT = 'yyyy-LL-dd HH:mm'

const store = useStore()
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
		type: Boolean as PropType<boolean>,
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

const canWrite = computed(() => store.state.currentList.maxRight > RIGHTS.READ)

const tasks = ref<Map<TaskModel['id'], TaskModel>>(new Map())
const ganttBars = ref<GanttBarObject[][]>([])

const defaultStartDate = format(new Date(), DATE_FORMAT)
const defaultEndDate = format(new Date((new Date()).setDate((new Date()).getDate() + 7)), DATE_FORMAT)

function transformTaskToGanttBar(t: TaskModel) {
	const black = 'var(--grey-800)'
	return [{
		startDate: t.startDate ? format(t.startDate, DATE_FORMAT) : defaultStartDate,
		endDate: t.endDate ? format(t.endDate, DATE_FORMAT) : defaultEndDate,
		ganttBarConfig: {
			id: t.id,
			label: t.title,
			hasHandles: true,
			style: {
				color: t.startDate ? (colorIsDark(t.getHexColor(t.hexColor)) ? black : 'white') : black,
				backgroundColor: t.startDate ? t.getHexColor(t.hexColor) : 'var(--grey-100)',
				border: t.startDate ? '' : '2px dashed var(--grey-300)',
				'text-decoration': t.done ? 'line-through' : null,
			},
		},
	} as GanttBarObject]
}

// We need a "real" ref object for the gantt bars to instantly update the tasks when they are dragged on the chart.
// A computed won't work directly.
function mapGanttBars() {
	ganttBars.value = []

	tasks.value.forEach(t => ganttBars.value.push(transformTaskToGanttBar(t)))
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

	mapGanttBars()
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
	mapGanttBars()

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

<style lang="scss">
// Not scoped because we need to style the elements inside the gantt chart component
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

<style scoped lang="scss">
.gantt-container {
	overflow-x: auto;
}

#g-gantt-chart {
	width: 2000px;
}
</style>
