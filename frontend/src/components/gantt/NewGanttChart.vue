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
			<GanttTimelineHeader
				:timeline-data="timelineData"
				:day-width-pixels="DAY_WIDTH_PIXELS"
			/>

			<GanttVerticalGridLines
				:timeline-data="timelineData"
				:total-width="totalWidth"
				:height="ganttRows.length * 40"
				:day-width-pixels="DAY_WIDTH_PIXELS"
			/>

			<!-- Gantt Chart Body -->
			<GanttChartBody
				ref="ganttChartBodyRef"
				:rows="ganttRows"
				:cells-by-row="cellsByRow"
				@update:focused="handleFocusChange"
			>
				<template #default="{ focusedRow, focusedCell }">
					<div class="gantt-rows">
						<GanttRow
							v-for="(rowId, index) in ganttRows"
							:id="rowId"
							:key="rowId"
							:index="index"
						>
							<!-- Row content with relative positioning -->
							<div class="gantt-row-content">
								<GanttRowBars
									:bars="ganttBars[index]"
									:total-width="totalWidth"
									:date-from-date="dateFromDate"
									:date-to-date="dateToDate"
									:day-width-pixels="DAY_WIDTH_PIXELS"
									:is-dragging="isDragging"
									:is-resizing="isResizing"
									:drag-state="dragState"
									:focused-row="focusedRow"
									:focused-cell="focusedCell"
									:row-id="rowId"
									@barPointerDown="handleBarPointerDown"
									@startResize="startResize"
								/>
							</div>
						</GanttRow>
					</div>
				</template>
			</GanttChartBody>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, toRefs, onUnmounted} from 'vue'
import {useRouter} from 'vue-router'

import {getHexColor} from '@/models/task'

import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'
import type {DateISO} from '@/types/DateISO'
import type {GanttFilters} from '@/views/project/helpers/useGanttFilters'
import type {GanttBarModel} from '@/composables/useGanttBar'

import GanttChartBody from '@/components/gantt/GanttChartBody.vue'
import GanttRow from '@/components/gantt/GanttRow.vue'
import GanttRowBars from '@/components/gantt/GanttRowBars.vue'
import GanttVerticalGridLines from '@/components/gantt/GanttVerticalGridLines.vue'
import GanttTimelineHeader from '@/components/gantt/GanttTimelineHeader.vue'
import Loading from '@/components/misc/Loading.vue'

import {MILLISECONDS_A_DAY} from '@/constants/date'
import dayjs from 'dayjs'
import {useDayjsLanguageSync} from '@/i18n/useDayjsLanguageSync'

const props = defineProps<{
	isLoading: boolean,
	filters: GanttFilters,
	tasks: Map<ITask['id'], ITask>,
	defaultTaskStartDate: DateISO
	defaultTaskEndDate: DateISO
}>()

const emit = defineEmits<{
  (e: 'update:task', task: ITaskPartialWithId): void
}>()

const {tasks, filters} = toRefs(props)

const dayjsLanguageLoading = useDayjsLanguageSync(dayjs)
const ganttContainer = ref(null)
const ganttChartBodyRef = ref<InstanceType<typeof GanttChartBody> | null>(null)
const router = useRouter()

// Reactive drag state
const isDragging = ref(false)
const isResizing = ref(false)

// Focus state
const currentFocusedRow = ref<string | null>(null)
const currentFocusedCell = ref<number | null>(null)

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

// Double-click and drag detection
let lastClickTime = 0
let dragStarted = false

function handleBarPointerDown(bar: GanttBarModel, event: PointerEvent) {
	event.preventDefault()
	
	// Set focus to the clicked task's row
	const barIndex = ganttBars.value.findIndex(barGroup => barGroup.some(b => b.id === bar.id))
	if (barIndex !== -1 && ganttRows.value[barIndex]) {
		setFocusToRow(ganttRows.value[barIndex])
	}
	
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
	
	// Find the bar element and set cursor directly for SVG elements
	const barGroup = (event.target as Element).closest('g')
	const barElement = barGroup?.querySelector('.gantt-bar')
	if (barElement) {
		(barElement as HTMLElement).style.setProperty('cursor', 'grabbing', 'important')
	}
	
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
		if (barElement) {
			(barElement as HTMLElement).style.removeProperty('cursor')
		}
		
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
	
	// Find the bar element and set cursor directly for SVG elements
	const barGroup = (event.target as Element).closest('g')
	const barElement = barGroup?.querySelector('.gantt-bar')
	if (barElement) {
		(barElement as HTMLElement).style.setProperty('cursor', 'col-resize', 'important')
	}
	
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
		if (barElement) {
			(barElement as HTMLElement).style.removeProperty('cursor')
		}
		
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

function handleFocusChange(payload: { row: string | null; cell: number | null }) {
	currentFocusedRow.value = payload.row
	currentFocusedCell.value = payload.cell
}

function setFocusToRow(rowId: string) {
	ganttChartBodyRef.value?.setFocus(rowId, 0)
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
</script>

<style scoped lang="scss">
.gantt-container {
	overflow-x: auto;
}

.gantt-chart-wrapper {
	inline-size: max-content;
	min-inline-size: 100%;
	position: relative;
}

.gantt-rows {
	position: relative;
	z-index: 2;
}

.gantt-row-content {
	position: relative;
	min-block-size: 40px;
	inline-size: 100%;
}
</style>
