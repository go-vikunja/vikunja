<template>
	<g-gantt-chart
		:chart-start="`${dateFrom} 00:00`"
		:chart-end="`${dateTo} 23:59`"
		:precision="precision"
		bar-start="startDate"
		bar-end="endDate"
		:grid="true"
		@dragend-bar="updateTask"
	>
		<g-gantt-row
			v-for="(bar, k) in ganttBars"
			:key="k"
			label=""
			:bars="bar"
		/>
	</g-gantt-chart>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import TaskCollectionService from '@/services/taskCollection'
import {format} from 'date-fns'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import TaskService from '@/services/task'

const dateFormat = 'yyyy-LL-dd kk:mm'

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

const tasks = ref([])
const ganttBars = ref([])

// We need a "real" ref object for the gantt bars to instantly update the tasks when they are dragged on the chart.
// A computed won't work directly.
function mapGanttBars() {
	const defaultStartDate = '2022-07-19 12:00'
	const defaultEndDate = '2022-07-25 12:00'

	tasks.value.forEach(t => ganttBars.value.push([{
		startDate: t.startDate ? format(t.startDate, dateFormat) : defaultStartDate,
		endDate: t.endDate ? format(t.endDate, dateFormat) : defaultEndDate,
		ganttBarConfig: {
			id: t.id,
			label: t.title,
			hasHandles: true,
			style: {
				color: colorIsDark(t.getHexColor()) ? 'black' : 'white',
				backgroundColor: t.getHexColor(),
			},
		},
	}]))
}

async function loadTasks() {
	tasks.value = new Map()

	const params = {
		sort_by: ['start_date', 'done', 'id'],
		order_by: ['asc', 'asc', 'desc'],
		filter_by: ['done', 'start_date', 'start_date'],
		filter_comparator: ['equals', 'greater_equals', 'less_equals'],
		filter_value: ['false', props.dateFrom, props.dateTo],
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
	await taskService.update(task)
	// TODO: Loading animation
}
</script>

<style>
.g-gantt-row-label {
	display: none !important;
}
</style>
