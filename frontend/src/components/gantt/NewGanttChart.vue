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
							<!-- SVG container for bars in this row -->
							<svg
								class="gantt-row-bars"
								:width="totalWidth"
								height="40"
								xmlns="http://www.w3.org/2000/svg"
							>
								<g
									v-for="bar in ganttBars[index]"
									:key="bar.id"
								>
									<!-- Main bar -->
									<rect
										:x="getBarX(bar)"
										:y="4"
										:width="getBarWidth(bar)"
										:height="32"
										:rx="4"
										:fill="getBarFill(bar)"
										:stroke="getBarStroke(bar)"
										:stroke-width="getBarStrokeWidth(bar)"
										:stroke-dasharray="!bar.meta?.hasActualDates ? '5,5' : 'none'"
										class="gantt-bar"
										@pointerdown="handleBarPointerDown(bar, $event)"
									/>
									
									<!-- Left resize handle -->
									<rect
										:x="getBarX(bar) - 3"
										:y="4"
										:width="6"
										:height="32"
										:rx="3"
										fill="var(--white)"
										stroke="var(--primary)"
										stroke-width="1"
										class="gantt-resize-handle gantt-resize-left"
										@pointerdown="startResize(bar, 'start', $event)"
									/>
									
									<!-- Right resize handle -->
									<rect
										:x="getBarX(bar) + getBarWidth(bar) - 3"
										:y="4"
										:width="6"
										:height="32"
										:rx="3"
										fill="var(--white)"
										stroke="var(--primary)"
										stroke-width="1"
										class="gantt-resize-handle gantt-resize-right"
										@pointerdown="startResize(bar, 'end', $event)"
									/>
									
									<!-- Task label with clipping -->
									<defs>
										<clipPath :id="`clip-${bar.id}`">
											<rect
												:x="getBarX(bar) + 2"
												:y="4"
												:width="getBarWidth(bar) - 4"
												:height="32"
												:rx="4"
											/>
										</clipPath>
									</defs>
									<text
										:x="getBarTextX(bar)"
										:y="24"
										class="gantt-bar-text"
										:fill="getBarTextColor(bar)"
										:clip-path="`url(#clip-${bar.id})`"
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
import {computed, ref, watch, toRefs, onUnmounted} from 'vue'
import {useRouter} from 'vue-router'

import { useGlobalNow } from '@/composables/useGlobalNow'
import {getHexColor} from '@/models/task'
import {colorIsDark} from '@/helpers/color/colorIsDark'

import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'
import type {DateISO} from '@/types/DateISO'
import type {GanttFilters} from '@/views/project/helpers/useGanttFilters'
import type {GanttBarModel} from '@/composables/useGanttBar'

import VikunjaStyledGanttChart from '@/components/gantt/styled/VikunjaStyledGanttChart.vue'
import VikunjaStyledGanttRow from '@/components/gantt/styled/VikunjaStyledGanttRow.vue'
import Loading from '@/components/misc/Loading.vue'
import GanttVerticalGridLines from '@/components/gantt/GanttVerticalGridLines.vue'

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

// Reactive drag state
const isDragging = ref(false)
const isResizing = ref(false)
const dragState = ref<{
	barId: string
	startX: number
	originalStart: Date
	originalEnd: Date
	currentDays: number
	edge?: 'start' | 'end'
} | null>(null)

// Event listener cleanup functions
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

// Computed properties for dynamic bar positions during drag/resize
const getBarX = computed(() => (bar: GanttBarModel) => {
	if (isDragging.value && dragState.value?.barId === bar.id) {
		const originalX = computeBarX(dragState.value.originalStart)
		const offset = dragState.value.currentDays * DAY_WIDTH_PIXELS
		return originalX + offset
	}
	if (isResizing.value && dragState.value?.barId === bar.id && dragState.value.edge === 'start') {
		const newStart = new Date(dragState.value.originalStart)
		newStart.setDate(newStart.getDate() + dragState.value.currentDays)
		return computeBarX(newStart)
	}
	return computeBarX(bar.start)
})

const getBarWidth = computed(() => (bar: GanttBarModel) => {
	if (isResizing.value && dragState.value?.barId === bar.id) {
		if (dragState.value.edge === 'start') {
			const newStart = new Date(dragState.value.originalStart)
			newStart.setDate(newStart.getDate() + dragState.value.currentDays)
			const originalEndX = computeBarX(dragState.value.originalEnd)
			const newStartX = computeBarX(newStart)
			return Math.max(0, originalEndX - newStartX)
		} else {
			const newEnd = new Date(dragState.value.originalEnd)
			newEnd.setDate(newEnd.getDate() + dragState.value.currentDays)
			const originalStartX = computeBarX(dragState.value.originalStart)
			const newEndX = computeBarX(newEnd)
			return Math.max(0, newEndX - originalStartX)
		}
	}
	return computeBarWidth(bar)
})

const getBarTextX = computed(() => (bar: GanttBarModel) => {
	return getBarX.value(bar) + 8
})

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
		startDate: dayjs(newStart).startOf('day').toDate(),
		endDate: dayjs(newEnd).endOf('day').toDate(),
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
	// For tasks with actual dates
	if (bar.meta?.hasActualDates) {
		// Use task color if available
		if (bar.meta?.color) {
			return bar.meta.color
		}
		// Default to primary color if no task color
		return 'var(--primary)'
	}
	
	// For tasks without dates, use gray
	return 'var(--grey-100)'
}

function getBarStroke(bar: GanttBarModel) {
	if (!bar.meta?.hasActualDates) {
		return 'var(--grey-300)' // Gray for dashed border
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
	const black = 'var(--grey-800)'
	
	// For tasks without actual dates, use dark text on gray background
	if (!bar.meta?.hasActualDates) {
		return black
	}
	
	// For tasks with actual dates
	if (bar.meta?.color) {
		// Use colorIsDark to determine text color based on background
		return colorIsDark(bar.meta.color) ? black : 'white'
	}
	
	// Default for primary color background (white text)
	return 'white'
}

// Double-click and drag detection
let lastClickTime = 0
let dragStarted = false

function handleBarPointerDown(bar: GanttBarModel, event: PointerEvent) {
	event.preventDefault()
	
	const currentTime = Date.now()
	const timeDiff = currentTime - lastClickTime
	
	// Double-click detection (within 500ms)
	if (timeDiff < 500) {
		openTask(bar)
		lastClickTime = 0
		return
	}
	
	lastClickTime = currentTime
	dragStarted = false
	
	const startX = event.clientX
	const startY = event.clientY
	
	const handleMove = (e: PointerEvent) => {
		const diffX = Math.abs(e.clientX - startX)
		const diffY = Math.abs(e.clientY - startY)
		
		// Start drag if mouse moved more than 5 pixels
		if (!dragStarted && (diffX > 5 || diffY > 5)) {
			dragStarted = true
			document.removeEventListener('pointermove', handleMove)
			document.removeEventListener('pointerup', handleStop)
			startDrag(bar, event)
		}
	}
	
	const handleStop = () => {
		document.removeEventListener('pointermove', handleMove)
		document.removeEventListener('pointerup', handleStop)
		// If no drag was started, this was just a click (do nothing)
	}
	
	document.addEventListener('pointermove', handleMove)
	document.addEventListener('pointerup', handleStop)
}

function startDrag(bar: GanttBarModel, event: PointerEvent) {
	event.preventDefault()
	
	// Initialize reactive drag state
	isDragging.value = true
	dragState.value = {
		barId: bar.id,
		startX: event.clientX,
		originalStart: new Date(bar.start),
		originalEnd: new Date(bar.end),
		currentDays: 0,
	}
	
	// Set grabbing cursor during drag
	document.body.style.setProperty('cursor', 'grabbing', 'important')
	
	const handleMove = (e: PointerEvent) => {
		if (!dragState.value || !isDragging.value) return
		
		const diff = e.clientX - dragState.value.startX
		const days = Math.round(diff / DAY_WIDTH_PIXELS)
		
		// Update reactive state - this will automatically update the template
		if (days !== dragState.value.currentDays) {
			dragState.value.currentDays = days
		}
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
		
		// Only dispatch update when drag is finished
		if (dragState.value && dragState.value.currentDays !== 0) {
			const newStart = new Date(dragState.value.originalStart)
			newStart.setDate(newStart.getDate() + dragState.value.currentDays)
			const newEnd = new Date(dragState.value.originalEnd)
			newEnd.setDate(newEnd.getDate() + dragState.value.currentDays)
			
			updateGanttTask(bar.id, newStart, newEnd)
		}
		
		// Reset drag state
		isDragging.value = false
		dragState.value = null
	}
	
	// Store handlers for cleanup
	dragMoveHandler = handleMove
	dragStopHandler = handleStop
	
	document.addEventListener('pointermove', handleMove)
	document.addEventListener('pointerup', handleStop)
}

function startResize(bar: GanttBarModel, edge: 'start' | 'end', event: PointerEvent) {
	event.preventDefault()
	event.stopPropagation() // Prevent drag from triggering
	
	// Initialize reactive resize state
	isResizing.value = true
	dragState.value = {
		barId: bar.id,
		startX: event.clientX,
		originalStart: new Date(bar.start),
		originalEnd: new Date(bar.end),
		currentDays: 0,
		edge,
	}
	
	// Set col-resize cursor during resize
	document.body.style.setProperty('cursor', 'col-resize', 'important')
	
	const handleMove = (e: PointerEvent) => {
		if (!dragState.value || !isResizing.value) return
		
		const diff = e.clientX - dragState.value.startX
		const days = Math.round(diff / DAY_WIDTH_PIXELS)
		
		// Validate resize bounds
		if (edge === 'start') {
			const newStart = new Date(dragState.value.originalStart)
			newStart.setDate(newStart.getDate() + days)
			if (newStart >= dragState.value.originalEnd) return
		} else {
			const newEnd = new Date(dragState.value.originalEnd)
			newEnd.setDate(newEnd.getDate() + days)
			if (newEnd <= dragState.value.originalStart) return
		}
		
		// Update reactive state - this will automatically update the template
		if (days !== dragState.value.currentDays) {
			dragState.value.currentDays = days
		}
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
		
		// Only dispatch update when resize is finished
		if (dragState.value && dragState.value.currentDays !== 0) {
			if (edge === 'start') {
				const newStart = new Date(dragState.value.originalStart)
				newStart.setDate(newStart.getDate() + dragState.value.currentDays)
				
				// Ensure start doesn't go past end
				if (newStart < dragState.value.originalEnd) {
					updateGanttTask(bar.id, newStart, dragState.value.originalEnd)
				}
			} else {
				const newEnd = new Date(dragState.value.originalEnd)
				newEnd.setDate(newEnd.getDate() + dragState.value.currentDays)
				
				// Ensure end doesn't go before start
				if (newEnd > dragState.value.originalStart) {
					updateGanttTask(bar.id, dragState.value.originalStart, newEnd)
				}
			}
		}
		
		// Reset resize state
		isResizing.value = false
		dragState.value = null
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
	// Reset cursor if component unmounts during drag
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

.gantt-row-bars {
	position: absolute;
	top: 0;
	left: 0;
	pointer-events: none;
	z-index: 4;
	
	:deep(.gantt-bar) {
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
	font-size: .85rem;
	pointer-events: none;
	user-select: none;
}

// Resize handles
:deep(.gantt-resize-handle) {
	cursor: col-resize !important;
	opacity: 0;
	transition: opacity 0.2s ease;
	pointer-events: all; // Ensure they receive pointer events
	
	&:hover {
		opacity: 1;
	}
}

// Show resize handles on bar hover
:deep(g:hover) .gantt-resize-handle {
	opacity: 0.8;
	
	&:hover {
		opacity: 1;
		cursor: inherit; // Use the specific cursor defined above
	}
}
</style>