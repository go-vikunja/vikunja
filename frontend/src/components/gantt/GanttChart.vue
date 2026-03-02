<template>
	<Loading
		v-if="(isLoading && !ganttBars.length) || dayjsLanguageLoading"
		class="gantt-container"
	/>
	<div
		v-else
		ref="ganttContainer"
		class="gantt-container"
		role="application"
		:aria-label="$t('project.gantt.chartLabel')"
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

			<GanttChartBody
				ref="ganttChartBodyRef"
				:rows="ganttRows"
				:cells-by-row="cellsByRow"
				@update:focused="handleFocusChange"
				@enterPressed="handleEnterPressed"
			>
				<template #default="{ focusedRow, focusedCell }">
					<div class="gantt-rows">
						<GanttRow
							v-for="(rowId, index) in ganttRows"
							:id="rowId"
							:key="rowId"
							:index="index"
						>
							<div class="gantt-row-content">
								<GanttRowBars
									:bars="ganttBars[index] ?? []"
									:total-width="totalWidth"
									:date-from-date="dateFromDate"
									:date-to-date="dateToDate"
									:day-width-pixels="DAY_WIDTH_PIXELS"
									:is-dragging="isDragging"
									:is-resizing="isResizing"
									:drag-state="dragState"
									:focused-row="focusedRow ?? null"
									:focused-cell="focusedCell"
									:row-id="rowId"
									@barPointerDown="handleBarPointerDown"
									@startResize="startResize"
									@updateTask="updateGanttTask"
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
import dayjs from 'dayjs'
import {useDayjsLanguageSync} from '@/i18n/useDayjsLanguageSync'

import {getHexColor} from '@/models/task'
import {buildGanttTaskTree, type GanttTaskTreeNode} from '@/helpers/ganttTaskTree'

import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'
import type {DateISO} from '@/types/DateISO'
import type {GanttFilters} from '@/views/project/helpers/useGanttFilters'
import type {GanttBarModel, GanttBarDateType} from '@/composables/useGanttBar'

import GanttChartBody from '@/components/gantt/GanttChartBody.vue'
import GanttRow from '@/components/gantt/GanttRow.vue'
import GanttRowBars from '@/components/gantt/GanttRowBars.vue'
import GanttVerticalGridLines from '@/components/gantt/GanttVerticalGridLines.vue'
import GanttTimelineHeader from '@/components/gantt/GanttTimelineHeader.vue'
import Loading from '@/components/misc/Loading.vue'

import {MILLISECONDS_A_DAY} from '@/constants/date'
import {roundToNaturalDayBoundary} from '@/helpers/time/roundToNaturalDayBoundary'

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

const DAY_WIDTH_PIXELS = 30

const {tasks, filters} = toRefs(props)

const dayjsLanguageLoading = useDayjsLanguageSync(dayjs)
const ganttContainer = ref(null)
const ganttChartBodyRef = ref<InstanceType<typeof GanttChartBody> | null>(null)
const router = useRouter()

const isDragging = ref(false)
const isResizing = ref(false)

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

let dragMoveHandler: ((e: PointerEvent) => void) | null = null
let dragStopHandler: (() => void) | null = null

const dateFromDate = computed(() => dayjs(filters.value.dateFrom).startOf('day').toDate())
const dateToDate = computed(() => dayjs(filters.value.dateTo).endOf('day').toDate())

const totalWidth = computed(() => {
	const dateDiff = Math.ceil((dateToDate.value.valueOf() - dateFromDate.value.valueOf()) / MILLISECONDS_A_DAY)
	return dateDiff * DAY_WIDTH_PIXELS
})

const timelineData = computed(() => {
	const dates: Date[] = []
	const currentDate = new Date(dateFromDate.value)
	
	while (currentDate <= dateToDate.value) {
		dates.push(new Date(currentDate))
		currentDate.setDate(currentDate.getDate() + 1)
	}
	
	return dates
})

const ganttBars = ref<GanttBarModel[][]>([])
const ganttRows = ref<string[]>([])
const cellsByRow = ref<Record<string, string[]>>({})

// Hierarchy state
const collapsedTaskIds = ref(new Set<number>())
const allNodes = ref<GanttTaskTreeNode[]>([])

const visibleNodes = computed(() => {
	const result: GanttTaskTreeNode[] = []
	const hiddenParents = new Set<number>()

	for (const node of allNodes.value) {
		const parents = node.task.relatedTasks?.parenttask ?? []
		const isHidden = parents.some(p =>
			collapsedTaskIds.value.has(p.id) || hiddenParents.has(p.id),
		)

		if (isHidden) {
			hiddenParents.add(node.task.id)
			continue
		}

		result.push(node)
	}

	return result
})

// Used in Task 8 for arrow re-routing when children are collapsed
const _hiddenToAncestor = computed(() => {
	const map = new Map<number, number>()
	const hiddenParents = new Set<number>()

	for (const node of allNodes.value) {
		const parents = node.task.relatedTasks?.parenttask ?? []
		const collapsedParent = parents.find(p =>
			collapsedTaskIds.value.has(p.id),
		)

		if (collapsedParent && tasks.value.has(collapsedParent.id)) {
			map.set(node.task.id, collapsedParent.id)
			hiddenParents.add(node.task.id)
		} else {
			const hiddenAncestor = parents.find(p => hiddenParents.has(p.id))
			if (hiddenAncestor) {
				const ancestorTarget = map.get(hiddenAncestor.id) ?? hiddenAncestor.id
				map.set(node.task.id, ancestorTarget)
				hiddenParents.add(node.task.id)
			}
		}
	}

	return map
})

// Used in Task 5 for collapse/expand toggle
function _toggleCollapse(taskId: number) {
	const newSet = new Set(collapsedTaskIds.value)
	if (newSet.has(taskId)) {
		newSet.delete(taskId)
	} else {
		newSet.add(taskId)
	}
	collapsedTaskIds.value = newSet
}

function getRoundedDate(value: string | Date | undefined, fallback: Date | string, isStart: boolean) {
	return roundToNaturalDayBoundary(value ? new Date(value) : new Date(fallback), isStart)
}

function transformTaskToGanttBar(node: GanttTaskTreeNode): GanttBarModel {
	const t = node.task
	const DEFAULT_SPAN_DAYS = 7

	// Use derived dates for dateless parents
	const effectiveEndDate = t.endDate || t.dueDate || (node.hasDerivedDates ? node.derivedEndDate : null)
	const effectiveStartDate = t.startDate || (node.hasDerivedDates ? node.derivedStartDate : null)

	let startDate: Date
	let endDate: Date
	let dateType: GanttBarDateType

	if (effectiveStartDate && effectiveEndDate) {
		startDate = getRoundedDate(effectiveStartDate, effectiveStartDate, true)
		endDate = getRoundedDate(effectiveEndDate, effectiveEndDate, false)
		dateType = 'both'
	} else if (effectiveStartDate && !effectiveEndDate) {
		startDate = getRoundedDate(effectiveStartDate, effectiveStartDate, true)
		const defaultEnd = new Date(startDate)
		defaultEnd.setDate(defaultEnd.getDate() + DEFAULT_SPAN_DAYS)
		endDate = getRoundedDate(defaultEnd, defaultEnd, false)
		dateType = 'startOnly'
	} else if (!effectiveStartDate && effectiveEndDate) {
		endDate = getRoundedDate(effectiveEndDate, effectiveEndDate, false)
		const defaultStart = new Date(endDate)
		defaultStart.setDate(defaultStart.getDate() - DEFAULT_SPAN_DAYS)
		startDate = getRoundedDate(defaultStart, defaultStart, true)
		dateType = 'endOnly'
	} else {
		startDate = getRoundedDate(undefined, props.defaultTaskStartDate, true)
		endDate = getRoundedDate(undefined, props.defaultTaskEndDate, false)
		dateType = 'both'
	}

	const taskColor = getHexColor(t.hexColor)

	return {
		id: String(t.id),
		start: startDate,
		end: endDate,
		meta: {
			label: t.title,
			task: t,
			color: taskColor,
			hasActualDates: Boolean(t.startDate && (t.endDate || t.dueDate)),
			dateType,
			isDone: t.done,
			isParent: node.isParent,
			hasDerivedDates: node.hasDerivedDates,
			indentLevel: node.indentLevel,
		},
	}
}

// Build the task tree when tasks change
watch(
	[tasks, filters],
	() => {
		allNodes.value = buildGanttTaskTree(tasks.value)
	},
	{deep: true, immediate: true},
)

// Derive bars, rows, and cells from visible nodes
watch(
	[visibleNodes, filters],
	() => {
		const bars: GanttBarModel[] = []
		const rows: string[] = []
		const cells: Record<string, string[]> = {}

		visibleNodes.value.forEach((node, index) => {
			const bar = transformTaskToGanttBar(node)

			// Check if task is visible in the current date range
			const hasAnyDate = Boolean(node.task.startDate || node.task.endDate || node.task.dueDate || node.hasDerivedDates)
			if (!filters.value.showTasksWithoutDates && !hasAnyDate) {
				return
			}
			if (bar.start > dateToDate.value || bar.end < dateFromDate.value) {
				return
			}

			bars.push(bar)

			const rowId = `row-${index}`
			rows.push(rowId)

			const rowCells: string[] = []
			timelineData.value.forEach((_, dayIndex) => {
				rowCells.push(`${rowId}-cell-${dayIndex}`)
			})
			cells[rowId] = rowCells
		})

		ganttBars.value = bars.map(bar => [bar])
		ganttRows.value = rows
		cellsByRow.value = cells
	},
	{deep: true, immediate: true},
)

function updateGanttTask(id: string, newStart: Date, newEnd: Date) {
	const task = tasks.value.get(Number(id))
	if (!task) return

	const update: ITaskPartialWithId = {
		id: Number(id),
	}

	const hasStartDate = Boolean(task.startDate)
	const hasEndDate = Boolean(task.endDate)
	const hasDueDate = Boolean(task.dueDate)

	if (hasStartDate && hasEndDate) {
		// Both dates exist — update both
		update.startDate = roundToNaturalDayBoundary(newStart, true)
		update.endDate = roundToNaturalDayBoundary(newEnd)
	} else if (hasStartDate && !hasEndDate && hasDueDate) {
		// startDate + dueDate (no endDate) — treat as fully dated
		update.startDate = roundToNaturalDayBoundary(newStart, true)
		update.dueDate = roundToNaturalDayBoundary(newEnd)
	} else if (hasStartDate && !hasEndDate) {
		// startOnly — only update startDate, don't persist the synthetic end
		update.startDate = roundToNaturalDayBoundary(newStart, true)
	} else if (!hasStartDate && (hasEndDate || hasDueDate)) {
		// endOnly / dueOnly — only update the end side
		if (hasEndDate) {
			update.endDate = roundToNaturalDayBoundary(newEnd)
		}
		if (hasDueDate) {
			update.dueDate = roundToNaturalDayBoundary(newEnd)
		}
	} else {
		// No dates at all — update both (existing behavior for dateless tasks)
		update.startDate = roundToNaturalDayBoundary(newStart, true)
		update.endDate = roundToNaturalDayBoundary(newEnd)
	}

	emit('update:task', update)
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

const DOUBLE_CLICK_THRESHOLD_MS = 500
const DRAG_THRESHOLD_PIXELS = 5

function handleBarPointerDown(bar: GanttBarModel, event: PointerEvent) {
	event.preventDefault()
	
	const barIndex = ganttBars.value.findIndex(barGroup => barGroup.some(b => b.id === bar.id))
	if (barIndex !== -1 && ganttRows.value[barIndex]) {
		focusTaskBar(ganttRows.value[barIndex])
	}
	
	const currentTime = Date.now()
	const timeDiff = currentTime - lastClickTime
	
	if (timeDiff < DOUBLE_CLICK_THRESHOLD_MS) {	
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
		
		// Start drag if mouse moved more than threshhold
		if (!dragStarted && (diffX > DRAG_THRESHOLD_PIXELS || diffY > DRAG_THRESHOLD_PIXELS)) {	
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

function setCursor(cursor: string, barElement?: Element | null) {
	document.body.style.setProperty('cursor', cursor, 'important')
	if (barElement) {
		(barElement as HTMLElement).style.setProperty('cursor', cursor, 'important')
	}
}

function clearCursor(barElement?: Element | null) {
	document.body.style.removeProperty('cursor')
	if (barElement) {
		(barElement as HTMLElement).style.removeProperty('cursor')
	}
}

function startDrag(bar: GanttBarModel, event: PointerEvent) {
	event.preventDefault()
	
	isDragging.value = true
	dragState.value = {
		barId: bar.id,
		startX: event.clientX,
		originalStart: new Date(bar.start),
		originalEnd: new Date(bar.end),
		currentDays: 0,
	}
	
	const barGroup = (event.target as Element).closest('g')
	const barElement = barGroup?.querySelector('.gantt-bar')
	setCursor('grabbing', barElement)
	
	const handleMove = (e: PointerEvent) => {
		if (!dragState.value || !isDragging.value) return
		
		const diff = e.clientX - dragState.value.startX
		const days = Math.round(diff / DAY_WIDTH_PIXELS)
		
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
		
		clearCursor(barElement)
		
		if (dragState.value && dragState.value.currentDays !== 0) {
			const newStart = new Date(dragState.value.originalStart)
			newStart.setDate(newStart.getDate() + dragState.value.currentDays)
			const newEnd = new Date(dragState.value.originalEnd)
			newEnd.setDate(newEnd.getDate() + dragState.value.currentDays)
			
			updateGanttTask(bar.id, newStart, newEnd)
		}
		
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
	
	isResizing.value = true
	dragState.value = {
		barId: bar.id,
		startX: event.clientX,
		originalStart: new Date(bar.start),
		originalEnd: new Date(bar.end),
		currentDays: 0,
		edge,
	}
	
	const barGroup = (event.target as Element).closest('g')
	const barElement = barGroup?.querySelector('.gantt-bar')
	setCursor('col-resize', barElement)
	
	const handleMove = (e: PointerEvent) => {
		if (!dragState.value || !isResizing.value) return
		
		const diff = e.clientX - dragState.value.startX
		const days = Math.round(diff / DAY_WIDTH_PIXELS)
		
		if (edge === 'start') {
			const newStart = new Date(dragState.value.originalStart)
			newStart.setDate(newStart.getDate() + days)
			if (newStart >= dragState.value.originalEnd) return
		} else {
			const newEnd = new Date(dragState.value.originalEnd)
			newEnd.setDate(newEnd.getDate() + days)
			if (newEnd <= dragState.value.originalStart) return
		}
		
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
		
		clearCursor(barElement)
		
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

function handleEnterPressed(payload: { row: string; cell: number }) {
	const rowIndex = ganttRows.value.indexOf(payload.row)
	if (rowIndex !== -1 && ganttBars.value[rowIndex]?.[0]) {
		const bar = ganttBars.value[rowIndex][0]
		openTask(bar)
	}
}

function focusTaskBar(rowId: string) {
	setTimeout(() => {
		const taskBarElement = document.querySelector(`[data-row-id="${rowId}"] [role="slider"]`) as HTMLElement
		if (taskBarElement) {
			taskBarElement.focus()
		}
	}, 0)
}

onUnmounted(() => {
	if (dragMoveHandler) {
		document.removeEventListener('pointermove', dragMoveHandler)
		dragMoveHandler = null
	}
	if (dragStopHandler) {
		document.removeEventListener('pointerup', dragStopHandler)
		dragStopHandler = null
	}
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
