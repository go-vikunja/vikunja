<template>
	<Loading
		v-if="isLoading && !ganttBars.length || dayjsLanguageLoading"
		class="gantt-container"
	/>
	<div
		v-else
		ref="ganttContainer"
		class="gantt-container"
	>
		<div class="gantt-chart-wrapper">
			<!-- Timeline Header -->
			<div class="gantt-timeline">
				<!-- Upper timeunit for months -->
				<div class="gantt-timeline-upper">
					<div
						v-for="monthGroup in monthGroups"
						:key="monthGroup.key"
						class="upper-timeunit"
						:style="{ width: `${monthGroup.width}px` }"
					>
						{{ monthGroup.label }}
					</div>
				</div>
				
				<!-- Lower timeunit for days -->
				<div class="gantt-timeline-lower">
					<div
						v-for="date in timelineData"
						:key="date.toISOString()"
						class="timeunit"
						:style="{ width: `${DAY_WIDTH_PIXELS}px` }"
					>
						<div
							class="timeunit-wrapper"
							:class="{'today': dateIsToday(date)}"
						>
							<span>{{ date.getDate() }}</span>
							<span class="weekday">
								{{ weekDayFromDate(date) }}
							</span>
						</div>
					</div>
				</div>
			</div>

			<GanttVerticalGridLines
				:timeline-data="timelineData"
				:total-width="totalWidth"
				:height="ganttRows.length * 40"
				:day-width-pixels="DAY_WIDTH_PIXELS"
			/>

			<!-- Gantt Chart Body -->
			<VikunjaStyledGanttChart
				:rows="ganttRows"
				:cells-by-row="cellsByRow"
				@update:focused="onFocusChanged"
			>
				<div class="gantt-rows">
					<VikunjaStyledGanttRow
						v-for="(rowId, index) in ganttRows"
						:id="rowId"
						:key="rowId"
						:index="index"
						:selected="false"
						@select="onRowSelect"
						@focus="onRowFocus"
					>
						<!-- Row content with relative positioning -->
						<div class="gantt-row-content">
							<GanttRowBars
								:bars="ganttBars[index]"
								:total-width="totalWidth"
								:date-from-date="dateFromDate"
								:date-to-date="dateToDate"
								:day-width-pixels="DAY_WIDTH_PIXELS"
								@updateTask="updateGanttTask"
								@openTask="openTask"
								@startResize="startResize"
							/>
						</div>
					</VikunjaStyledGanttRow>
				</div>
			</VikunjaStyledGanttChart>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, toRefs, onUnmounted} from 'vue'
import {useRouter} from 'vue-router'

import { useGlobalNow } from '@/composables/useGlobalNow'
import {getHexColor} from '@/models/task'

import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'
import type {DateISO} from '@/types/DateISO'
import type {GanttFilters} from '@/views/project/helpers/useGanttFilters'
import type {GanttBarModel} from '@/composables/useGanttBar'

import VikunjaStyledGanttChart from '@/components/gantt/styled/VikunjaStyledGanttChart.vue'
import VikunjaStyledGanttRow from '@/components/gantt/styled/VikunjaStyledGanttRow.vue'
import Loading from '@/components/misc/Loading.vue'
import GanttVerticalGridLines from '@/components/gantt/GanttVerticalGridLines.vue'
import GanttRowBars from '@/components/gantt/GanttRowBars.vue'

import {MILLISECONDS_A_DAY} from '@/constants/date'
import {useWeekDayFromDate} from '@/helpers/time/formatDate'
import dayjs from 'dayjs'
import {useDayjsLanguageSync} from '@/i18n/useDayjsLanguageSync'

export interface GanttChartProps {
	isLoading: boolean,
	filters: GanttFilters,
	tasks: Map<ITask['id'], ITask>,
	defaultTaskStartDate: DateISO
	defaultTaskEndDate: DateISO
}

const props = defineProps<GanttChartProps>()

const emit = defineEmits<{
  (e: 'update:task', task: ITaskPartialWithId): void
}>()

const {tasks, filters} = toRefs(props)

const dayjsLanguageLoading = useDayjsLanguageSync(dayjs)
const ganttContainer = ref(null)
const router = useRouter()

// Event listener cleanup functions for resize operations
let dragMoveHandler: ((e: PointerEvent) => void) | null = null
let dragStopHandler: (() => void) | null = null

const dateFromDate = computed(() => dayjs(filters.value.dateFrom).startOf('day').toDate())
const dateToDate = computed(() => dayjs(filters.value.dateTo).endOf('day').toDate())

const DAY_WIDTH_PIXELS = 30
const totalWidth = computed(() => {
	const dateDiff = Math.ceil((dateToDate.value.valueOf() - dateFromDate.value.valueOf()) / MILLISECONDS_A_DAY)
	return dateDiff * DAY_WIDTH_PIXELS
})

// Generate timeline data (array of dates)
const timelineData = computed(() => {
	const dates: Date[] = []
	const currentDate = new Date(dateFromDate.value)
	
	while (currentDate <= dateToDate.value) {
		dates.push(new Date(currentDate))
		currentDate.setDate(currentDate.getDate() + 1)
	}
	
	return dates
})

// Generate month groups for the upper timeline
const monthGroups = computed(() => {
	const groups: Array<{key: string; label: string; width: number}> = []
	let currentMonth = -1
	let currentYear = -1
	let dayCount = 0
	
	timelineData.value.forEach((date, index) => {
		const month = date.getMonth()
		const year = date.getFullYear()
		
		if (month !== currentMonth || year !== currentYear) {
			// Finish previous group
			if (currentMonth !== -1) {
				groups[groups.length - 1].width = dayCount * DAY_WIDTH_PIXELS
			}
			
			// Start new group
			currentMonth = month
			currentYear = year
			dayCount = 1
			
			const monthName = dayjs(date).format('MMMM YYYY')
			groups.push({
				key: `${year}-${month}`,
				label: monthName,
				width: 0, // Will be set when we finish the group
			})
		} else {
			dayCount++
		}
		
		// Handle last group
		if (index === timelineData.value.length - 1) {
			groups[groups.length - 1].width = dayCount * DAY_WIDTH_PIXELS
		}
	})
	
	return groups
})

// Transform tasks to gantt bars
const ganttBars = ref<GanttBarModel[][]>([])
const ganttRows = ref<string[]>([])
const cellsByRow = ref<Record<string, string[]>>({})

function transformTaskToGanttBar(t: ITask): GanttBarModel {
	const startDate = t.startDate ? new Date(t.startDate) : new Date(props.defaultTaskStartDate)
	const endDate = t.endDate ? new Date(t.endDate) : new Date(props.defaultTaskEndDate)
	
	const taskColor = getHexColor(t.hexColor)
	
	const bar = {
		id: String(t.id),
		start: startDate,
		end: endDate,
		meta: {
			label: t.title,
			task: t,
			color: taskColor,
			hasActualDates: Boolean(t.startDate && t.endDate),
			isDone: t.done,
		},
	}
	
	
	return bar
}


/**
 * Update ganttBars when tasks change
 */
watch(
	[tasks, filters],
	() => {
		const bars: GanttBarModel[] = []
		const rows: string[] = []
		const cells: Record<string, string[]> = {}
		
		// Filter tasks based on current filters
		const filteredTasks = Array.from(tasks.value.values()).filter(task => {
			// If showTasksWithoutDates is false, only show tasks with actual dates
			if (!filters.value.showTasksWithoutDates && (!task.startDate || !task.endDate)) {
				return false
			}
			
			// Check if task is within the date range
			const taskStart = task.startDate ? new Date(task.startDate) : new Date(props.defaultTaskStartDate)
			const taskEnd = task.endDate ? new Date(task.endDate) : new Date(props.defaultTaskEndDate)
			
			// Task is visible if it overlaps with the current date range
			return taskStart <= dateToDate.value && taskEnd >= dateFromDate.value
		})
		
		// For now, create one row per task (simple implementation)
		// In the future, this could group tasks by project, parent task, etc.
		filteredTasks.forEach((t, index) => {
			const bar = transformTaskToGanttBar(t)
			bars.push(bar)
			
			const rowId = `row-${index}`
			rows.push(rowId)
			
			// Create cells for each day in the timeline
			const rowCells: string[] = []
			timelineData.value.forEach((date, dayIndex) => {
				rowCells.push(`${rowId}-cell-${dayIndex}`)
			})
			cells[rowId] = rowCells
		})
		
		// Group bars by rows (one bar per row for now)
		ganttBars.value = bars.map(bar => [bar])
		ganttRows.value = rows
		cellsByRow.value = cells
		
	},
	{deep: true, immediate: true},
)

// Debounce task updates to only emit on drag end, but allow immediate visual feedback
const updateTimeouts = ref<Map<string, NodeJS.Timeout>>(new Map())

function updateGanttTask(id: string, newStart: Date, newEnd: Date) {
	// Clear existing timeout for this task
	const existingTimeout = updateTimeouts.value.get(id)
	if (existingTimeout) {
		clearTimeout(existingTimeout)
	}
	
	// Set a new timeout to emit the change after drag stops
	const timeout = setTimeout(() => {
		emit('update:task', {
			id: Number(id),
			startDate: dayjs(newStart).startOf('day').toDate(),
			endDate: dayjs(newEnd).endOf('day').toDate(),
		})
		updateTimeouts.value.delete(id)
	}, 150) // 150ms delay to detect end of drag
	
	updateTimeouts.value.set(id, timeout)
}

function openTask(bar: GanttBarModel) {
	router.push({
		name: 'task.detail',
		params: {id: bar.id},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

function onFocusChanged() {
	// Handle focus changes if needed
}

function onRowSelect() {
	// Handle row selection if needed
}

function onRowFocus() {
	// Handle row focus if needed
}



function startResize(bar: GanttBarModel, edge: 'start' | 'end', event: PointerEvent) {
	event.preventDefault()
	event.stopPropagation() // Prevent drag from triggering
	
	const startX = event.clientX
	const originalStart = new Date(bar.start)
	const originalEnd = new Date(bar.end)
	let finalDays = 0
	
	// Set col-resize cursor during resize
	document.body.style.setProperty('cursor', 'col-resize', 'important')
	
	// Find the bar element and set cursor directly for SVG elements
	const barGroup = (event.target as Element).closest('g')
	const barElement = barGroup?.querySelector('.gantt-bar')
	if (barElement) {
		(barElement as HTMLElement).style.setProperty('cursor', 'col-resize', 'important')
	}
	
	const handleMove = (e: PointerEvent) => {
		const diff = e.clientX - startX
		const days = Math.round(diff / DAY_WIDTH_PIXELS)
		
		// Validate resize bounds
		if (edge === 'start') {
			const newStart = new Date(originalStart)
			newStart.setDate(newStart.getDate() + days)
			if (newStart >= originalEnd) return
		} else {
			const newEnd = new Date(originalEnd)
			newEnd.setDate(newEnd.getDate() + days)
			if (newEnd <= originalStart) return
		}
		
		finalDays = days
	}
	
	const handleStop = () => {
		if (dragMoveHandler) {
			document.removeEventListener('pointermove', dragMoveHandler)
			dragMoveHandler = null
		}
		if (dragStopHandler) {
			document.removeEventListener('pointerup', dragStopHandler)
			dragStopHandler = null
		}
		
		// Reset cursor
		document.body.style.removeProperty('cursor')
		if (barElement) {
			(barElement as HTMLElement).style.removeProperty('cursor')
		}
		
		// Use the final days from the last move event
		if (finalDays !== 0) {
			if (edge === 'start') {
				const newStart = new Date(originalStart)
				newStart.setDate(newStart.getDate() + finalDays)
				
				// Ensure start doesn't go past end
				if (newStart < originalEnd) {
					updateGanttTask(bar.id, newStart, originalEnd)
				}
			} else {
				const newEnd = new Date(originalEnd)
				newEnd.setDate(newEnd.getDate() + finalDays)
				
				// Ensure end doesn't go before start
				if (newEnd > originalStart) {
					updateGanttTask(bar.id, originalStart, newEnd)
				}
			}
		}
	}
	
	// Store handlers for cleanup
	dragMoveHandler = handleMove
	dragStopHandler = handleStop
	
	document.addEventListener('pointermove', handleMove)
	document.addEventListener('pointerup', handleStop)
}

// Cleanup event listeners on component unmount
onUnmounted(() => {
	if (dragMoveHandler) {
		document.removeEventListener('pointermove', dragMoveHandler)
		dragMoveHandler = null
	}
	if (dragStopHandler) {
		document.removeEventListener('pointerup', dragStopHandler)
		dragStopHandler = null
	}
	// Clear any pending update timeouts
	updateTimeouts.value.forEach(timeout => clearTimeout(timeout))
	updateTimeouts.value.clear()
	// Reset cursor if component unmounts during resize
	document.body.style.removeProperty('cursor')
})

const weekDayFromDate = useWeekDayFromDate()

const {now: today} = useGlobalNow()
const dateIsToday = computed(() => (date: Date) => {
	return (
		date.getDate() === today.value.getDate() &&
		date.getMonth() === today.value.getMonth() &&
		date.getFullYear() === today.value.getFullYear()
	)
})
</script>

<style scoped lang="scss">
.gantt-container {
	overflow-x: auto;
	
	--bar-bg: var(--grey-100);
	--bar-bg-active: var(--primary);
	--bar-bg-drag: var(--primary-light);
	--bar-stroke-focus: var(--primary);
	--text-on-bar: var(--grey-800);
	--row-bg: var(--white);
	--row-alt-bg: hsla(var(--grey-100-hsl), .5);
	--row-selected-bg: var(--primary-light);
}

.gantt-chart-wrapper {
	width: max-content;
	min-width: 100%;
	position: relative;
}


.gantt-timeline {
	background: var(--white);
	border-bottom: 1px solid var(--grey-200);
	position: sticky;
	top: 0;
	z-index: 10;
}

.gantt-timeline-upper {
	display: flex;
	
	.upper-timeunit {
		background: var(--white);
		font-family: $vikunja-font;
		font-weight: bold;
		border-right: 1px solid var(--grey-200);
		padding: 0.5rem 0;
		text-align: center;
		font-size: 1rem;
		color: var(--grey-800);
	}
}

.gantt-timeline-lower {
	display: flex;
	
	.timeunit {	
		.timeunit-wrapper {
			padding: 0.5rem 0;
			font-size: 1rem;
			display: flex;
			flex-direction: column;
			align-items: center;
			width: 100%;
			font-family: $vikunja-font;
			
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
	}
}

.gantt-rows {
	position: relative;
	z-index: 2;
}

.gantt-row-content {
	position: relative;
	min-height: 40px;
	width: 100%;
}


// Ensure rows have minimum height and proper styling
:deep(.gantt-row) {
	min-height: 40px;
	position: relative;
	border-bottom: 1px solid var(--grey-200);
	z-index: 2;
	
	&:nth-child(odd) {
		background: var(--row-alt-bg);
	}
	
	&:nth-child(even) {
		background: var(--row-bg);
	}
}

</style>