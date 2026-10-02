<template>
	<div
		class="gantt-pane"
		:style="{inlineSize: `${width}px`}"
		role="table"
		:aria-label="$t('project.gantt.pane.label')"
	>
		<div class="gantt-pane-inner">
			<div
				class="gantt-pane-header"
				role="row"
			>
				<span
					v-for="column in columns"
					:key="column.key"
					:class="['gantt-pane-cell', `col-${column.key}`]"
					role="columnheader"
				>{{ column.label }}</span>
			</div>

			<div
				v-for="row in rows"
				:key="row.id"
				class="gantt-pane-row"
				:class="{'is-critical': row.isCritical, 'is-done': row.isDone, 'is-parent': row.isParent}"
				role="row"
			>
				<span class="gantt-pane-cell col-wbs">{{ row.wbs }}</span>

				<span
					class="gantt-pane-cell col-name"
					:style="{paddingInlineStart: `${row.indent * 14 + 6}px`}"
				>
					<span
						v-if="row.isMilestone"
						class="gantt-pane-diamond"
						aria-hidden="true"
					>&#9670;</span>
					<input
						class="gantt-pane-input name"
						type="text"
						:value="row.title"
						:disabled="!editable"
						:aria-label="$t('project.gantt.pane.name')"
						@change="onTitle(row, $event)"
						@keydown.enter="($event.target as HTMLInputElement).blur()"
					>
					<BaseButton
						class="gantt-pane-open"
						:aria-label="$t('project.gantt.pane.open')"
						@click="emit('openTask', row.id)"
					>
						<Icon icon="arrow-up-right-from-square" />
					</BaseButton>
				</span>

				<span class="gantt-pane-cell col-duration">
					{{ row.duration === null ? '' : $t('project.gantt.pane.days', {count: row.duration}) }}
				</span>

				<span class="gantt-pane-cell col-start">
					<input
						class="gantt-pane-input"
						type="date"
						:value="toDateInput(row.start)"
						:disabled="!editable || row.isMilestone || row.isParent"
						:aria-label="$t('project.gantt.pane.start')"
						@change="onDate(row, 'start', $event)"
					>
				</span>

				<span class="gantt-pane-cell col-finish">
					<input
						class="gantt-pane-input"
						type="date"
						:value="toDateInput(row.end)"
						:disabled="!editable || row.isParent"
						:aria-label="$t('project.gantt.pane.finish')"
						@change="onDate(row, 'end', $event)"
					>
				</span>

				<span class="gantt-pane-cell col-percent">
					<input
						class="gantt-pane-input percent"
						type="number"
						min="0"
						max="100"
						step="5"
						:value="row.percent"
						:disabled="!editable || row.isParent || row.isMilestone"
						:aria-label="$t('project.gantt.pane.percent')"
						@change="onPercent(row, $event)"
					>
				</span>

				<span class="gantt-pane-cell col-predecessors">
					<span
						v-for="pred in row.predecessors"
						:key="pred.id"
						class="gantt-pane-chip"
					>
						{{ pred.label }}
						<BaseButton
							v-if="editable"
							class="gantt-pane-chip-remove"
							:aria-label="$t('project.gantt.removeDependency')"
							@click="emit('deleteDependency', pred.id, row.id)"
						>&times;</BaseButton>
					</span>
				</span>

				<span class="gantt-pane-cell col-assignees">{{ row.assignees }}</span>

				<span
					v-if="showVariance"
					class="gantt-pane-cell col-variance"
					:class="{'is-late': (row.variance ?? 0) > 0, 'is-early': (row.variance ?? 0) < 0}"
				>
					{{ row.variance === null ? '' : varianceLabel(row.variance) }}
				</span>
			</div>
		</div>

		<div
			class="gantt-pane-resizer"
			role="separator"
			aria-orientation="vertical"
			:aria-label="$t('project.gantt.pane.resize')"
			@pointerdown.prevent="startResize"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'

import type {GanttPaneRow} from './ganttPaneTypes'
import {GANTT_HEADER_HEIGHT_PX, GANTT_ROW_HEIGHT_PX} from './ganttPaneTypes'

const props = withDefaults(defineProps<{
	rows: GanttPaneRow[]
	editable?: boolean
	width: number
	showVariance?: boolean
}>(), {
	editable: false,
	showVariance: false,
})

const emit = defineEmits<{
	(e: 'update:width', width: number): void
	(e: 'update:task', update: {id: number, title?: string, percent_done?: number}): void
	(e: 'changeDates', id: number, start: Date | null, end: Date | null): void
	(e: 'deleteDependency', predecessorId: number, successorId: number): void
	(e: 'openTask', id: number): void
}>()

const {t} = useI18n({useScope: 'global'})

const MIN_WIDTH = 260
const MAX_WIDTH = 900

const headerPx = `${GANTT_HEADER_HEIGHT_PX}px`
const rowPx = `${GANTT_ROW_HEIGHT_PX}px`

const columns = computed(() => [
	{key: 'wbs', label: t('project.gantt.pane.wbs')},
	{key: 'name', label: t('project.gantt.pane.name')},
	{key: 'duration', label: t('project.gantt.pane.duration')},
	{key: 'start', label: t('project.gantt.pane.start')},
	{key: 'finish', label: t('project.gantt.pane.finish')},
	{key: 'percent', label: t('project.gantt.pane.percent')},
	{key: 'predecessors', label: t('project.gantt.pane.predecessors')},
	{key: 'assignees', label: t('project.gantt.pane.assignees')},
	...(props.showVariance ? [{key: 'variance', label: t('project.gantt.pane.variance')}] : []),
])

function pad(n: number) {
	return String(n).padStart(2, '0')
}

// A date input works on local calendar days.
function toDateInput(date: Date | null): string {
	return date ? `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` : ''
}

function fromDateInput(value: string): Date | null {
	const [y, m, d] = value.split('-').map(Number)
	if (!y || !m || !d) {
		return null
	}
	return new Date(y, m - 1, d)
}

function onTitle(row: GanttPaneRow, event: Event) {
	const value = (event.target as HTMLInputElement).value.trim()
	if (value === '' || value === row.title) {
		;(event.target as HTMLInputElement).value = row.title
		return
	}
	emit('update:task', {id: row.id, title: value})
}

function onPercent(row: GanttPaneRow, event: Event) {
	const raw = Number((event.target as HTMLInputElement).value)
	const percent = Math.min(100, Math.max(0, Number.isFinite(raw) ? Math.round(raw) : row.percent))
	;(event.target as HTMLInputElement).value = String(percent)
	if (percent !== row.percent) {
		emit('update:task', {id: row.id, percent_done: percent / 100})
	}
}

// Changing one date keeps the other, and never lets the start end up after the finish.
function onDate(row: GanttPaneRow, which: 'start' | 'end', event: Event) {
	const picked = fromDateInput((event.target as HTMLInputElement).value)
	if (!picked) {
		return
	}
	let start = row.start
	let end = row.end
	if (which === 'start') {
		start = picked
		if (end && end < picked) {
			end = picked
		}
	} else {
		end = picked
		if (row.isMilestone) {
			start = picked
		} else if (start && start > picked) {
			start = picked
		}
	}
	emit('changeDates', row.id, start, end)
}

function varianceLabel(days: number): string {
	if (days === 0) {
		return t('project.gantt.pane.onPlan')
	}
	return days > 0
		? t('project.gantt.pane.late', {count: days})
		: t('project.gantt.pane.early', {count: Math.abs(days)})
}

let cleanup: (() => void) | null = null

function startResize(event: PointerEvent) {
	const startX = event.clientX
	const startWidth = props.width

	// In a right-to-left layout the pane grows to the other side.
	const direction = getComputedStyle(document.documentElement).direction === 'rtl' ? -1 : 1
	const move = (e: PointerEvent) => {
		const next = startWidth + (e.clientX - startX) * direction
		emit('update:width', Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, Math.round(next))))
	}
	const stop = () => {
		document.removeEventListener('pointermove', move)
		document.removeEventListener('pointerup', stop)
		cleanup = null
	}
	document.addEventListener('pointermove', move)
	document.addEventListener('pointerup', stop)
	cleanup = stop
}

onBeforeUnmount(() => cleanup?.())
</script>

<style scoped lang="scss">
.gantt-pane {
	position: relative;
	flex: none;
	overflow-x: auto;
	border-inline-end: 1px solid var(--grey-300);
	background: var(--white);
	font-size: 0.85rem;
}

.gantt-pane-inner {
	min-inline-size: 760px;
}

.gantt-pane-header,
.gantt-pane-row {
	display: flex;
	align-items: center;
	box-sizing: border-box;
}

.gantt-pane-header {
	block-size: v-bind(headerPx);
	position: sticky;
	inset-block-start: 0;
	z-index: 10;
	background: var(--white);
	border-block-end: 1px solid var(--grey-200);
	font-weight: bold;
	color: var(--grey-700);
}

.gantt-pane-row {
	block-size: v-bind(rowPx);
	border-block-end: 1px solid var(--grey-100);

	&.is-parent {
		font-weight: 600;
		background: var(--grey-50);
	}

	&.is-done {
		color: var(--grey-500);

		.name {
			text-decoration: line-through;
		}
	}

	&.is-critical .col-wbs {
		color: var(--danger);
		font-weight: bold;
	}
}

.gantt-pane-cell {
	box-sizing: border-box;
	padding: 0 6px;
	overflow: hidden;
	white-space: nowrap;
	text-overflow: ellipsis;
	flex: none;
}

.col-wbs { inline-size: 48px; }
.col-name { flex: 1 1 auto; min-inline-size: 170px; display: flex; align-items: center; gap: 4px; }
.col-duration { inline-size: 62px; }
.col-start,
.col-finish { inline-size: 124px; }
.col-percent { inline-size: 64px; }
.col-predecessors { inline-size: 110px; }
.col-assignees { inline-size: 110px; }
.col-variance { inline-size: 84px; }

.gantt-pane-input {
	inline-size: 100%;
	min-inline-size: 0;
	box-sizing: border-box;
	padding: 2px 4px;
	font: inherit;
	color: inherit;
	background: transparent;
	border: 1px solid transparent;
	border-radius: 3px;

	&:hover:not(:disabled) {
		border-color: var(--grey-300);
	}

	&:focus {
		outline: none;
		border-color: var(--primary);
		background: var(--white);
	}

	&:disabled {
		cursor: default;
	}

	&.percent {
		text-align: end;
	}
}

.gantt-pane-diamond {
	color: var(--primary);
	flex: none;
}

.gantt-pane-open {
	flex: none;
	opacity: 0;
	color: var(--grey-500);
}

.gantt-pane-row:hover .gantt-pane-open,
.gantt-pane-open:focus {
	opacity: 1;
}

.gantt-pane-chip {
	display: inline-flex;
	align-items: center;
	gap: 2px;
	padding: 0 4px;
	margin-inline-end: 3px;
	background: var(--grey-100);
	border-radius: 3px;
	font-size: 0.78rem;
}

.gantt-pane-chip-remove {
	line-height: 1;
	color: var(--grey-500);

	&:hover {
		color: var(--danger);
	}
}

.is-late { color: var(--danger); }
.is-early { color: var(--success); }

.gantt-pane-resizer {
	position: absolute;
	inset-block: 0;
	inset-inline-end: 0;
	inline-size: 5px;
	cursor: col-resize;
	z-index: 11;

	&:hover {
		background: var(--primary);
		opacity: 0.4;
	}
}
</style>
