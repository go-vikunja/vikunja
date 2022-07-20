<template>
	<div class="gantt-container">
		<g-gantt-chart
			:chart-start="`${dateFrom} 00:00`"
			:chart-end="`${dateTo} 23:59`"
			:precision="precision"
			bar-start="startDate"
			bar-end="endDate"
			:grid="true"
			@dragend-bar="updateTask"
			@dblclick-bar="openTask"
			font="'Open Sans', sans-serif"
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
	<form
		@submit.prevent="createTask()"
		class="add-new-task"
		v-if="canWrite"
	>
		<transition name="width">
			<input
				@blur="hideCreateNewTask"
				@keyup.esc="newTaskFieldActive = false"
				class="input"
				ref="newTaskTitleField"
				type="text"
				v-if="newTaskFieldActive"
				v-model="newTaskTitle"
			/>
		</transition>
		<x-button @click="showCreateNewTask" :shadow="false" icon="plus">
			{{ $t('task.new') }}
		</x-button>
	</form>
</template>

<script setup lang="ts">
import {computed, nextTick, ref} from 'vue'
import TaskCollectionService from '@/services/taskCollection'
import {format, parse} from 'date-fns'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import TaskService from '@/services/task'
import {useStore} from 'vuex'
import Rights from '../../models/constants/rights.json'
import TaskModel from '@/models/task'
import {useRouter} from 'vue-router'

const dateFormat = 'yyyy-LL-dd HH:mm'

const store = useStore()
const router = useRouter()

const props = defineProps({
	listId: {
		type: Number,
		required: true,
	},
	precision: {
		type: String,
		default: 'day',
	},
	dateFrom: {
		type: String,
		required: true,
	},
	dateTo: {
		type: String,
		required: true,
	},
})

const dateFromDate = computed(() => parse(props.dateFrom, 'yyyy-LL-dd', new Date()))
const dateToDate = computed(() => parse(props.dateTo, 'yyyy-LL-dd', new Date()))

const DAY_WIDTH_PIXELS = 30
const ganttChartWidth = computed(() => {
	const dateDiff = Math.floor((dateToDate.value - dateFromDate.value) / (1000 * 60 * 60 * 24))

	return dateDiff * DAY_WIDTH_PIXELS
})

const canWrite = computed(() => store.state.currentList.maxRight > Rights.READ)

const tasks = ref([])
const ganttBars = ref([])

const defaultStartDate = format(new Date(), dateFormat)
const defaultEndDate = format(new Date((new Date()).setDate((new Date()).getDate() + 7)), dateFormat)

function transformTaskToGanttBar(t: TaskModel) {
	const black = 'var(--grey-800)'
	return [{
		startDate: t.startDate ? format(t.startDate, dateFormat) : defaultStartDate,
		endDate: t.endDate ? format(t.endDate, dateFormat) : defaultEndDate,
		ganttBarConfig: {
			id: t.id,
			label: t.title,
			hasHandles: true,
			style: {
				color: t.startDate ? (colorIsDark(t.getHexColor()) ? black : 'white') : black,
				backgroundColor: t.startDate ? t.getHexColor() : 'var(--grey-100)',
				border: t.startDate ? '' : '2px dashed var(--grey-300)',
				'text-decoration': t.done ? 'line-through' : null,
			},
		},
	}]
}

// We need a "real" ref object for the gantt bars to instantly update the tasks when they are dragged on the chart.
// A computed won't work directly.
function mapGanttBars() {
	ganttBars.value = []

	tasks.value.forEach(t => ganttBars.value.push(transformTaskToGanttBar(t)))
}

async function loadTasks() {
	tasks.value = new Map()

	const params = {
		sort_by: ['start_date', 'done', 'id'],
		order_by: ['asc', 'asc', 'desc'],
		filter_by: ['start_date', 'start_date'],
		filter_comparator: ['greater_equals', 'less_equals'],
		filter_value: [props.dateFrom, props.dateTo],
		filter_concat: 'and',
		filter_include_nulls: true,
	}

	const taskCollectionService = new TaskCollectionService()

	const getAllTasks = async (page = 1) => {
		const tasks = await taskCollectionService.getAll({listId: props.listId}, params, page)
		if (page < taskCollectionService.totalPages) {
			const nextTasks = await getAllTasks(page + 1)
			return tasks.concat(nextTasks)
		}
		return tasks
	}

	const loadedTasks = await getAllTasks()

	loadedTasks
		.forEach(t => {
			tasks.value.set(t.id, t)
		})

	mapGanttBars()
}

loadTasks()

async function updateTask(e) {
	const task = tasks.value.get(e.bar.ganttBarConfig.id)
	task.startDate = e.bar.startDate
	task.endDate = e.bar.endDate
	const taskService = new TaskService()
	const r = await taskService.update(task)
	// TODO: Loading animation
	for (const i in ganttBars.value) {
		if (ganttBars.value[i][0].ganttBarConfig.id === task.id) {
			ganttBars.value[i] = transformTaskToGanttBar(r)
		}
	}
}

const newTaskFieldActive = ref(false)
const newTaskTitleField = ref()
const newTaskTitle = ref('')

function showCreateNewTask() {
	if (!newTaskFieldActive.value) {
		// Timeout to not send the form if the field isn't even shown
		setTimeout(() => {
			newTaskFieldActive.value = true
			nextTick(() => newTaskTitleField.value.focus())
		}, 100)
	}
}

function hideCreateNewTask() {
	if (newTaskTitle.value === '') {
		nextTick(() => (newTaskFieldActive.value = false))
	}
}

async function createTask() {
	if (!newTaskFieldActive.value) {
		return
	}
	let task = new TaskModel({
		title: newTaskTitle.value,
		listId: props.listId,
		startDate: defaultStartDate,
		endDate: defaultEndDate,
	})
	const taskService = new TaskService()
	const r = await taskService.create(task)
	tasks.value.set(r.id, r)
	mapGanttBars()
	newTaskTitle.value = ''
	hideCreateNewTask()
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

.add-new-task {
	padding: 1rem .7rem .4rem .7rem;
	display: flex;
	max-width: 450px;

	.input {
		margin-right: .7rem;
		font-size: .8rem;
	}

	.button {
		font-size: .68rem;
	}
}
</style>
