<template>
	<div class="month-grid">
		<div class="month-head">
			<div
				v-for="day in days.slice(0, 7)"
				:key="day.toISOString()"
				class="month-weekday"
			>
				{{ formatWeekday(day) }}
			</div>
		</div>

		<div
			ref="bodyEl"
			class="month-body"
		>
			<div
				v-for="day in days"
				:key="day.toISOString()"
				class="month-cell"
				:class="{
					'is-today': isToday(day),
					'is-outside': !isCurrentMonth(day),
					'is-drop-target': dropDay === dayKey(day),
				}"
				:data-day="dayKey(day)"
				@dragover.prevent="dropDay = dayKey(day)"
				@dragleave="onDragLeave($event, day)"
				@drop="onDrop($event, day)"
				@dblclick="onDblClick($event, day)"
			>
				<div class="cell-head">
					<span class="cell-number">{{ day.getDate() }}</span>
				</div>

				<div class="cell-items">
					<button
						v-for="item in visibleItems(day)"
						:key="item.task.id"
						class="month-chip"
						:class="{'is-done': item.task.done, 'is-ghost': item.isGhost}"
						:style="{'--chip-color': taskColor(item.task), '--chip-text': getTextColor(taskColor(item.task))}"
						:title="chipTitle(item.task)"
						:draggable="!item.isGhost"
						@dragstart="onChipDragStart($event, item.task)"
						@click="emit('openTask', item.task.id)"
					>
						{{ item.task.title }}
					</button>

					<button
						v-if="hiddenCount(day) > 0"
						class="month-more"
						@click="expanded.add(dayKey(day))"
					>
						{{ $t('planner.showMore', {count: hiddenCount(day)}) }}
					</button>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onMounted, reactive, ref, watch} from 'vue'
import {useResizeObserver} from '@vueuse/core'
import dayjs from 'dayjs'

import type {TaskResponse} from '@/client/queries/tasks'
import {getTextColor} from '@/helpers/color/getTextColor'
import {useWeekDayFromDate} from '@/helpers/time/formatDate'
import {useProjects} from '@/composables/useProjects'
import {dayKey, monthItemsByDay, type AllDayItem} from '../helpers/dayLayout'
import {plannerTaskColor} from '../helpers/taskColor'

const props = defineProps<{
	days: Date[]
	tasks: TaskResponse[]
	// Days outside this month are shown greyed; the grid always renders whole weeks.
	month: Date
}>()

const emit = defineEmits<{
	openTask: [taskId: number]
	dropAllDay: [payload: {taskId: number, day: Date}]
	createAllDay: [payload: {day: Date}]
}>()

const projectList = useProjects()
const formatWeekday = useWeekDayFromDate()

const itemsByDay = computed(() => monthItemsByDay(props.tasks, props.days))

// Cleared whenever the visible month changes, so paging doesn't carry a cell's
// expanded state over to whatever day lands in that position next.
const expanded = reactive(new Set<string>())
watch(() => props.month, () => expanded.clear())

function itemsFor(day: Date): AllDayItem[] {
	return itemsByDay.value.get(dayKey(day)) ?? []
}

// Used until a cell has been measured, and if the measurement can't be taken.
const FALLBACK_CAPACITY = 3
const FALLBACK_CHIP_HEIGHT = 18

const bodyEl = ref<HTMLElement>()
// How many chip rows fit in one cell at the current row height. Cells stretch to
// fill the grid, so this is a measurement rather than a fixed cap — a tall month
// shows many more chips than a short one before collapsing the rest.
const capacity = ref(FALLBACK_CAPACITY)

function measureCapacity() {
	const cell = bodyEl.value?.querySelector('.month-cell')
	const items = cell?.querySelector('.cell-items')
	if (!cell || !items) {
		return
	}

	const cellStyle = getComputedStyle(cell)
	const gap = parseFloat(getComputedStyle(items).rowGap) || 0
	const chipHeight = items.querySelector('.month-chip, .month-more')?.getBoundingClientRect().height
		|| FALLBACK_CHIP_HEIGHT

	const available = cell.clientHeight
		- (cell.querySelector('.cell-head')?.getBoundingClientRect().height ?? 0)
		- parseFloat(cellStyle.paddingBlockStart)
		- parseFloat(cellStyle.paddingBlockEnd)
		- gap

	capacity.value = Math.max(Math.floor((available + gap) / (chipHeight + gap)), 1)
}

useResizeObserver(bodyEl, measureCapacity)
onMounted(() => nextTick(measureCapacity))
// Chip height is read off a rendered chip, so re-measure once the data lands.
watch([() => props.days, itemsByDay], () => nextTick(measureCapacity))

function visibleItems(day: Date): AllDayItem[] {
	const items = itemsFor(day)
	if (expanded.has(dayKey(day)) || items.length <= capacity.value) {
		return items
	}
	// The "+N more" toggle occupies a row of its own.
	return items.slice(0, Math.max(capacity.value - 1, 1))
}

function hiddenCount(day: Date): number {
	return itemsFor(day).length - visibleItems(day).length
}

function isToday(day: Date): boolean {
	return dayjs(day).isSame(dayjs(), 'day')
}

function isCurrentMonth(day: Date): boolean {
	return dayjs(day).isSame(dayjs(props.month), 'month')
}

function taskColor(task: TaskResponse): string {
	return plannerTaskColor(task.hex_color, projectList.projects[task.project_id]?.hex_color)
}

// Chips are single-line and often clipped; surface the project in the tooltip.
function chipTitle(task: TaskResponse): string {
	const project = projectList.projects[task.project_id]?.title
	return project ? `${task.title} · ${project}` : task.title
}

const dropDay = ref<string | null>(null)

function onDragLeave(event: DragEvent, day: Date) {
	// Moving onto a child of the same cell still counts as being over it.
	const next = event.relatedTarget
	if (next instanceof Node && (event.currentTarget as HTMLElement).contains(next)) {
		return
	}
	if (dropDay.value === dayKey(day)) {
		dropDay.value = null
	}
}

function onDrop(event: DragEvent, day: Date) {
	dropDay.value = null
	const taskId = Number(event.dataTransfer?.getData('text/plain'))
	if (taskId) {
		emit('dropAllDay', {taskId, day})
	}
}

function onChipDragStart(event: DragEvent, task: TaskResponse) {
	event.dataTransfer?.setData('text/plain', String(task.id))
	if (event.dataTransfer) {
		event.dataTransfer.effectAllowed = 'move'
	}
}

// Only an empty part of the cell creates; double-clicking a chip opens it.
function onDblClick(event: MouseEvent, day: Date) {
	const target = event.target
	if (target instanceof HTMLElement && target.closest('.month-chip, .month-more')) {
		return
	}
	emit('createAllDay', {day})
}
</script>

<style lang="scss" scoped>
.month-grid {
	display: flex;
	flex-direction: column;
	flex: 1;
	min-block-size: 0;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	overflow: hidden;
	background: var(--white);
}

.month-head {
	display: grid;
	grid-template-columns: repeat(7, 1fr);
	border-block-end: 1px solid var(--grey-200);
	background: var(--grey-50);
}

.month-weekday {
	padding: .35rem .5rem;
	font-size: .75rem;
	font-weight: 600;
	color: var(--grey-500);
	text-align: center;
	text-transform: uppercase;
}

.month-body {
	display: grid;
	grid-template-columns: repeat(7, 1fr);
	grid-auto-rows: 1fr;
	flex: 1;
	min-block-size: 0;
	overflow-y: auto;
}

.month-cell {
	display: flex;
	flex-direction: column;
	gap: .15rem;
	min-block-size: 5.5rem;
	padding: .25rem;
	border-inline-end: 1px solid var(--grey-200);
	border-block-end: 1px solid var(--grey-200);
	overflow: hidden;

	&:nth-child(7n) {
		border-inline-end: 0;
	}

	&.is-outside {
		background: var(--grey-50);

		.cell-number {
			color: var(--grey-400);
		}
	}

	&.is-today .cell-number {
		background: var(--primary);
		color: var(--white);
		border-radius: 100%;
	}

	&.is-drop-target {
		background: var(--primary-light);
	}
}

.cell-head {
	display: flex;
	justify-content: flex-end;
}

.cell-number {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	min-inline-size: 1.35rem;
	block-size: 1.35rem;
	font-size: .8rem;
	font-weight: 600;
	color: var(--grey-700);
}

.cell-items {
	display: flex;
	flex-direction: column;
	gap: .15rem;
	overflow: hidden;
}

.month-chip {
	display: block;
	inline-size: 100%;
	padding: .1rem .3rem;
	border: 0;
	border-radius: $radius;
	background: var(--chip-color);
	color: var(--chip-text);
	font-size: .7rem;
	line-height: 1.3;
	text-align: start;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
	cursor: pointer;

	&.is-done {
		text-decoration: line-through;
		opacity: .65;
	}

	// Projected occurrences of a recurring task are read-only.
	&.is-ghost {
		opacity: .5;
		cursor: default;
	}
}

.month-more {
	padding: .05rem .3rem;
	border: 0;
	background: transparent;
	color: var(--grey-500);
	font-size: .7rem;
	text-align: start;
	cursor: pointer;

	&:hover {
		color: var(--grey-700);
		text-decoration: underline;
	}
}
</style>
