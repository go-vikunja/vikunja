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
							<!-- SVG container for bars in this row -->
							<svg
								class="gantt-row-bars"
								:width="totalWidth"
								height="32"
								xmlns="http://www.w3.org/2000/svg"
							>
								<g
									v-for="bar in ganttBars[index]"
									:key="bar.id"
									@dblclick="openTask(bar)"
									@pointerdown="startDrag(bar, $event)"
								>
									<rect
										:x="computeBarX(bar.start)"
										:y="2"
										:width="computeBarWidth(bar)"
										:height="28"
										:rx="4"
										:fill="getBarFill(bar)"
										:stroke="getBarStroke(bar)"
										:stroke-width="getBarStrokeWidth(bar)"
										:stroke-dasharray="!bar.meta?.hasActualDates ? '5,5' : 'none'"
										class="gantt-bar"
									/>
									<text
										:x="computeBarX(bar.start) + 8"
										:y="20"
										class="gantt-bar-text"
										:fill="getBarTextColor(bar)"
									>
										{{ bar.meta?.label || bar.id }}
									</text>
								</g>
							</svg>
						</div>
					</VikunjaStyledGanttRow>
				</div>
			</VikunjaStyledGanttChart>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, toRefs} from 'vue'
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

const dateFromDate = computed(() => new Date(new Date(filters.value.dateFrom).setHours(0,0,0,0)))
const dateToDate = computed(() => new Date(new Date(filters.value.dateTo).setHours(23,59,0,0)))

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
			
			const monthName = date.toLocaleDateString('en', { month: 'long', year: 'numeric' })
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

function updateGanttTask(id: string, newStart: Date, newEnd: Date) {
	emit('update:task', {
		id: Number(id),
		startDate: new Date(newStart.setHours(0,0,0,0)),
		endDate: new Date(newEnd.setHours(23,59,0,0)),
	})
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

// Direct SVG bar rendering functions
function computeBarX(startDate: Date) {
	const x = (startDate.getTime() - dateFromDate.value.getTime()) / (1000*60*60*24) * DAY_WIDTH_PIXELS
	return x
}

function computeBarWidth(bar: GanttBarModel) {
	const diff = (bar.end.getTime() - bar.start.getTime()) / (1000*60*60*24)
	const width = diff * DAY_WIDTH_PIXELS
	return width
}

function getBarFill(bar: GanttBarModel) {
	// Use task color if available and has actual dates
	if (bar.meta?.hasActualDates && bar.meta?.color) {
		return bar.meta.color
	}
	
	// Default colors
	if (bar.meta?.hasActualDates) {
		return '#1dd1a1' // Primary green
	}
	
	return '#d3d3d3' // Light gray for tasks without dates
}

function getBarStroke(bar: GanttBarModel) {
	if (!bar.meta?.hasActualDates) {
		return '#bdc3c7' // Gray for dashed border
	}
	return 'none'
}

function getBarStrokeWidth(bar: GanttBarModel) {
	if (!bar.meta?.hasActualDates) {
		return '2'
	}
	return '0'
}

function getBarTextColor(bar: GanttBarModel) {
	// For tasks without actual dates, use dark text
	if (!bar.meta?.hasActualDates) {
		return '#2c3e50'
	}
	
	// For tasks with color, determine text color based on background
	if (bar.meta?.color) {
		// Simple brightness check - you may want to import colorIsDark if needed
		return '#ffffff'
	}
	
	// Default for primary color background (white text on green)
	return '#ffffff'
}

function startDrag(bar: GanttBarModel, event: PointerEvent) {
	// Simple drag implementation
	const startX = event.clientX
	const originalStart = new Date(bar.start)
	
	const handleMove = (e: PointerEvent) => {
		const diff = e.clientX - startX
		const days = Math.round(diff / DAY_WIDTH_PIXELS)
		const newStart = new Date(originalStart)
		newStart.setDate(newStart.getDate() + days)
		const newEnd = new Date(bar.end)
		newEnd.setDate(newEnd.getDate() + days)
		
		updateGanttTask(bar.id, newStart, newEnd)
	}
	
	const handleStop = () => {
		document.removeEventListener('pointermove', handleMove)
		document.removeEventListener('pointerup', handleStop)
	}
	
	document.addEventListener('pointermove', handleMove)
	document.addEventListener('pointerup', handleStop)
}

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
}

.gantt-row-content {
	position: relative;
	min-height: 32px;
	width: 100%;
}

.gantt-row-bars {
	position: absolute;
	top: 0;
	left: 0;
	pointer-events: none;
	z-index: 5;
	
	:deep(rect) {
		pointer-events: all;
		cursor: grab;
		
		&:hover {
			opacity: 0.8;
		}
	}
	
	:deep(text) {
		pointer-events: none;
		user-select: none;
	}
}

// Ensure rows have minimum height and proper styling
:deep(.gantt-row) {
	min-height: 32px;
	position: relative;
	border-bottom: 1px solid var(--grey-200);
	
	&:nth-child(odd) {
		background: var(--row-alt-bg);
	}
	
	&:nth-child(even) {
		background: var(--row-bg);
	}
}

// SVG bar styling
.gantt-bar {
	cursor: grab;
	
	&:hover {
		opacity: 0.8;
	}
	
	&:active {
		cursor: grabbing;
	}
}

.gantt-bar-text {
	font-family: $vikunja-font;
	font-size: 14px;
	font-weight: 500;
	pointer-events: none;
	user-select: none;
}
</style>