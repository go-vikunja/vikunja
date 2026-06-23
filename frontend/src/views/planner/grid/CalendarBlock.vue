<template>
	<div
		ref="blockEl"
		class="calendar-block"
		:class="{
			'is-ghost': occurrence.isGhost,
			'is-done': occurrence.task.done,
			'is-dragging': isInteracting,
			'is-moving': isMoving,
		}"
		:style="blockStyle"
		@pointerdown="onMovePointerDown"
	>
		<span class="block-time">{{ timeLabel }}</span>
		<span class="block-title">{{ occurrence.task.title }}</span>
		<div
			v-if="!occurrence.isGhost"
			class="resize-handle"
			@pointerdown.stop="onResizePointerDown"
		>
			<span class="grip-dot" />
			<span class="grip-dot" />
			<span class="grip-dot" />
		</div>
	</div>

	<Teleport to="body">
		<div
			v-if="isMoving"
			class="calendar-block drag-preview"
			:style="previewStyle"
		>
			<span class="block-time">{{ timeLabel }}</span>
			<span class="block-title">{{ occurrence.task.title }}</span>
		</div>
	</Teleport>
</template>

<script setup lang="ts">
import {computed, ref} from 'vue'
import dayjs from 'dayjs'

import {useProjectStore} from '@/stores/projects'
import {useTimeFormat} from '@/composables/useTimeFormat'
import {TIME_FORMAT} from '@/constants/timeFormat'
import type {PlannedOccurrence} from '../helpers/types'

const props = withDefaults(defineProps<{
	occurrence: PlannedOccurrence
	col: number
	cols: number
	topMinutes: number
	durationMinutes: number
	pxPerMinute: number
	slotMinutes: number
	originMinutes?: number
}>(), {
	originMinutes: 0,
})

const emit = defineEmits<{
	open: [taskId: number]
	update: [payload: {taskId: number, start: Date | null, end: Date | null}]
}>()

const projectStore = useProjectStore()
const {store: timeFormat} = useTimeFormat()
const blockEl = ref<HTMLElement | null>(null)

const resizeDeltaMinutes = ref(0)
const isInteracting = ref(false)
const isMoving = ref(false)
const grabOffset = ref({x: 0, y: 0})
const previewPos = ref({x: 0, y: 0})
const previewSize = ref({w: 0, h: 0})

const color = computed(() => {
	const project = projectStore.projects[props.occurrence.task.projectId]
	const hex = project?.hexColor || props.occurrence.task.hexColor
	if (!hex) {
		return 'var(--primary)'
	}
	return hex.startsWith('#') ? hex : `#${hex}`
})

const effectiveTop = computed(() => (props.topMinutes - props.originMinutes) * props.pxPerMinute)
const effectiveHeight = computed(() => Math.max(
	(props.durationMinutes + resizeDeltaMinutes.value) * props.pxPerMinute,
	props.slotMinutes * props.pxPerMinute,
))

const blockStyle = computed(() => ({
	top: `${effectiveTop.value}px`,
	height: `${effectiveHeight.value}px`,
	insetInlineStart: `${(props.col / props.cols) * 100}%`,
	inlineSize: `${(1 / props.cols) * 100}%`,
	'--block-color': color.value,
}))

// A floating clone teleported to <body> follows the cursor, so it isn't clipped
// by the grid's scroll container when dragged over the sidebar/all-day row.
const previewStyle = computed(() => ({
	left: `${previewPos.value.x}px`,
	top: `${previewPos.value.y}px`,
	inlineSize: `${previewSize.value.w}px`,
	blockSize: `${previewSize.value.h}px`,
	'--block-color': color.value,
}))

const timeLabel = computed(() => dayjs(props.occurrence.start)
	.format(timeFormat.value === TIME_FORMAT.HOURS_24 ? 'HH:mm' : 'h:mm A'))

function snap(deltaPx: number): number {
	const minutes = deltaPx / props.pxPerMinute
	return Math.round(minutes / props.slotMinutes) * props.slotMinutes
}

function onMovePointerDown(event: PointerEvent) {
	if (props.occurrence.isGhost) {
		// Ghosts are read-only, but still let the user open the underlying task.
		emit('open', props.occurrence.task.id)
		return
	}

	const startY = event.clientY
	const startX = event.clientX
	const rect = blockEl.value?.getBoundingClientRect()
	grabOffset.value = {x: rect ? startX - rect.left : 0, y: rect ? startY - rect.top : 0}
	previewSize.value = {w: rect?.width ?? 0, h: rect?.height ?? 0}
	previewPos.value = {x: rect?.left ?? startX, y: rect?.top ?? startY}
	let moved = false

	const onMove = (e: PointerEvent) => {
		previewPos.value = {x: e.clientX - grabOffset.value.x, y: e.clientY - grabOffset.value.y}
		if (Math.abs(e.clientX - startX) > 3 || Math.abs(e.clientY - startY) > 3) {
			moved = true
			isInteracting.value = true
			isMoving.value = true
		}
	}

	const onUp = (e: PointerEvent) => {
		document.removeEventListener('pointermove', onMove)
		document.removeEventListener('pointerup', onUp)

		const taskId = props.occurrence.task.id
		// Hit-test from the preview block's top-centre (what the user visually
		// aligns), not the cursor, which sits at the grab offset further down.
		const hitX = previewPos.value.x + previewSize.value.w / 2
		const hitY = previewPos.value.y
		const dropEl = document.elementFromPoint(hitX, hitY)
		const overSidebar = dropEl?.closest('.planner-sidebar')
		const allDayCell = dropEl?.closest<HTMLElement>('.all-day-cell')
		const dayColumn = dropEl?.closest<HTMLElement>('.day-column')
		const targetDay = dayColumn?.dataset.day ?? null

		if (overSidebar) {
			// Drop on the sidebar → unschedule (back to the unscheduled list).
			emit('update', {taskId, start: null, end: null})
		} else if (allDayCell?.dataset.day) {
			// Drop on the all-day row → make it an all-day task on that day.
			const day = dayjs(allDayCell.dataset.day).startOf('day').toDate()
			emit('update', {taskId, start: day, end: day})
		} else {
			const dayChanged = targetDay !== null && !dayjs(targetDay).isSame(props.occurrence.start, 'day')
			if (!moved && !dayChanged) {
				emit('open', taskId)
			} else {
				const origStart = dayjs(props.occurrence.start)
				const minutesFromMidnight = origStart.diff(origStart.startOf('day'), 'minute')
				const newMinutes = Math.min(
					Math.max(minutesFromMidnight + snap(e.clientY - startY), 0),
					24 * 60 - props.durationMinutes,
				)
				const baseDay = targetDay ? dayjs(targetDay).startOf('day') : origStart.startOf('day')
				const newStart = baseDay.add(newMinutes, 'minute')
				emit('update', {
					taskId,
					start: newStart.toDate(),
					end: newStart.add(props.durationMinutes, 'minute').toDate(),
				})
			}
		}
		isInteracting.value = false
		isMoving.value = false
	}

	document.addEventListener('pointermove', onMove)
	document.addEventListener('pointerup', onUp)
}

function onResizePointerDown(event: PointerEvent) {
	const startY = event.clientY

	const onMove = (e: PointerEvent) => {
		isInteracting.value = true
		resizeDeltaMinutes.value = snap(e.clientY - startY)
	}

	const onUp = () => {
		document.removeEventListener('pointermove', onMove)
		document.removeEventListener('pointerup', onUp)

		const newDuration = Math.max(props.durationMinutes + resizeDeltaMinutes.value, props.slotMinutes)
		if (newDuration !== props.durationMinutes) {
			emit('update', {
				taskId: props.occurrence.task.id,
				start: new Date(props.occurrence.start),
				end: dayjs(props.occurrence.start).add(newDuration, 'minute').toDate(),
			})
		}
		resizeDeltaMinutes.value = 0
		isInteracting.value = false
	}

	document.addEventListener('pointermove', onMove)
	document.addEventListener('pointerup', onUp)
}
</script>

<style lang="scss" scoped>
.calendar-block {
	position: absolute;
	overflow: hidden;
	padding: 2px 6px;
	border-radius: 4px;
	border-inline-start: 3px solid var(--block-color);
	background-color: var(--block-color);
	color: var(--white);
	cursor: grab;
	user-select: none;
	font-size: .75rem;
	line-height: 1.1;
	box-shadow: 0 1px 2px hsla(0, 0%, 0%, .15);

	&.is-dragging {
		cursor: grabbing;
	}

	// While moving, dim the original in place and let elementFromPoint see what's
	// underneath (the floating preview is what the user actually follows).
	&.is-moving {
		opacity: .3;
		pointer-events: none;
	}

	&.is-ghost {
		opacity: .45;
		cursor: pointer;
		background-image: repeating-linear-gradient(
			45deg,
			hsla(0, 0%, 100%, .15),
			hsla(0, 0%, 100%, .15) 4px,
			transparent 4px,
			transparent 8px
		);
	}

	&.is-done {
		opacity: .6;

		.block-title {
			text-decoration: line-through;
		}
	}
}

.drag-preview {
	position: fixed;
	z-index: 100;
	pointer-events: none;
	opacity: .95;
	box-shadow: 0 4px 12px hsla(0, 0%, 0%, .3);
}

.block-time {
	display: block;
	font-weight: 700;
	opacity: .85;
}

.block-title {
	display: block;
	white-space: normal;
	overflow-wrap: anywhere;
}

.resize-handle {
	position: absolute;
	inset-block-end: 0;
	inset-inline: 0;
	block-size: 9px;
	cursor: ns-resize;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 3px;
}

.grip-dot {
	inline-size: 3px;
	block-size: 3px;
	border-radius: 50%;
	background-color: hsla(0, 0%, 100%, .8);
}
</style>
