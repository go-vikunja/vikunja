<template>
	<div class="print-page">
		<header class="print-header">
			<div>
				<h1>{{ $t('export.printTitle', {project: projectTitle, view: viewTitle}) }}</h1>
				<p class="print-meta">
					<span>{{ $t('export.printedOn', {date: printedOn}) }}</span>
					<span v-if="filterText"> · {{ $t('export.printFilter', {filter: filterText}) }}</span>
				</p>
			</div>
			<button
				type="button"
				class="button is-primary no-print"
				@click="printNow"
			>
				{{ $t('export.printNow') }}
			</button>
		</header>

		<p
			v-if="loading"
			class="print-note"
		>
			{{ $t('export.printLoading') }}
		</p>
		<p
			v-else-if="failed"
			class="print-note has-text-danger"
		>
			{{ failed }}
		</p>
		<template v-else>
			<p
				v-if="truncated"
				class="print-note"
			>
				{{ $t('export.printTruncated', {count: MAX_EXPORT_TASKS}) }}
			</p>

			<!-- Table and list -->
			<table
				v-if="table"
				class="print-table"
			>
				<thead>
					<tr>
						<th
							v-for="header in table.headers"
							:key="header"
						>
							{{ header }}
						</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="(row, index) in table.rows"
						:key="index"
						:class="{'is-done': table.doneRows[index]}"
					>
						<td
							v-for="(cell, cellIndex) in row"
							:key="cellIndex"
						>
							{{ cell }}
						</td>
					</tr>
				</tbody>
			</table>

			<!-- Kanban -->
			<div
				v-else-if="board"
				class="print-board"
			>
				<section
					v-for="column in board.columns"
					:key="column.title"
					class="print-bucket"
				>
					<h2>{{ column.title }} <span class="count">{{ column.tasks.length }}</span></h2>
					<article
						v-for="(card, index) in column.tasks"
						:key="index"
						class="print-card"
						:class="{'is-done': card.done}"
					>
						<strong>{{ card.title }}</strong>
						<span
							v-if="card.details.length"
							class="details"
						>{{ card.details.join(' · ') }}</span>
					</article>
				</section>
			</div>

			<!-- Gantt: the task table on the left, the whole date range as one drawing on the right -->
			<div
				v-else-if="gantt"
				class="print-gantt"
			>
				<table class="print-table gantt-table">
					<thead>
						<tr>
							<th
								class="head-spacer"
								:style="{height: `${HEADER_HEIGHT}px`}"
							>
								{{ $t('project.gantt.pane.wbs') }}
							</th>
							<th>{{ $t('project.gantt.pane.name') }}</th>
							<th>{{ $t('project.gantt.pane.duration') }}</th>
							<th>{{ $t('project.gantt.pane.start') }}</th>
							<th>{{ $t('project.gantt.pane.finish') }}</th>
							<th>{{ $t('project.gantt.pane.percent') }}</th>
							<th>{{ $t('project.gantt.pane.predecessors') }}</th>
							<th>{{ $t('project.gantt.pane.assignees') }}</th>
						</tr>
					</thead>
					<tbody>
						<tr
							v-for="row in gantt.rows"
							:key="row.id"
							:style="{height: `${ROW_HEIGHT}px`}"
							:class="{'is-critical': row.critical, 'is-summary': row.kind === 'summary'}"
						>
							<td>{{ row.wbs }}</td>
							<td :style="{paddingInlineStart: `${row.indent * 12 + 4}px`}">
								{{ row.title }}
							</td>
							<td>{{ row.duration === null ? '' : $t('project.gantt.pane.days', {count: row.duration}) }}</td>
							<td>{{ row.start ? formatDay(row.start) : '' }}</td>
							<td>{{ row.end ? formatDay(row.end) : '' }}</td>
							<td>{{ row.percent }}%</td>
							<td>{{ row.predecessors }}</td>
							<td>{{ row.assignees }}</td>
						</tr>
					</tbody>
				</table>

				<svg
					class="gantt-drawing"
					:width="gantt.width"
					:height="HEADER_HEIGHT + gantt.rows.length * ROW_HEIGHT"
					xmlns="http://www.w3.org/2000/svg"
				>
					<g
						v-for="cell in gantt.upper"
						:key="`u-${cell.x}`"
					>
						<rect
							:x="cell.x"
							y="0"
							:width="cell.width"
							:height="UPPER_HEIGHT"
							class="head-cell"
						/>
						<text
							:x="cell.x + cell.width / 2"
							:y="UPPER_HEIGHT / 2 + 4"
							text-anchor="middle"
							class="head-text"
						>{{ cell.label }}</text>
					</g>
					<g
						v-for="cell in gantt.lower"
						:key="`l-${cell.x}`"
					>
						<rect
							:x="cell.x"
							:y="UPPER_HEIGHT"
							:width="cell.width"
							:height="HEADER_HEIGHT - UPPER_HEIGHT"
							class="head-cell"
						/>
						<text
							v-if="cell.width >= 9"
							:x="cell.x + cell.width / 2"
							:y="UPPER_HEIGHT + (HEADER_HEIGHT - UPPER_HEIGHT) / 2 + 3"
							text-anchor="middle"
							class="head-text small"
						>{{ cell.label }}</text>
					</g>

					<line
						v-for="x in gantt.lines"
						:key="`g-${x}`"
						:x1="x"
						:x2="x"
						:y1="HEADER_HEIGHT"
						:y2="HEADER_HEIGHT + gantt.rows.length * ROW_HEIGHT"
						class="grid-line"
					/>
					<line
						v-for="(row, index) in gantt.rows"
						:key="`r-${row.id}`"
						x1="0"
						:x2="gantt.width"
						:y1="HEADER_HEIGHT + (index + 1) * ROW_HEIGHT"
						:y2="HEADER_HEIGHT + (index + 1) * ROW_HEIGHT"
						class="grid-line"
					/>

					<line
						v-if="gantt.todayX !== null"
						:x1="gantt.todayX"
						:x2="gantt.todayX"
						:y1="HEADER_HEIGHT"
						:y2="HEADER_HEIGHT + gantt.rows.length * ROW_HEIGHT"
						class="today-line"
					/>

					<template
						v-for="(row, index) in gantt.rows"
						:key="`b-${row.id}`"
					>
						<polygon
							v-if="row.bar && row.kind === 'milestone'"
							:points="diamond(row.bar.x, rowCenter(index))"
							class="bar milestone"
							:class="{critical: row.critical}"
						/>
						<g v-else-if="row.bar">
							<rect
								:x="row.bar.x"
								:y="rowCenter(index) - (row.kind === 'summary' ? 4 : 8)"
								:width="row.bar.width"
								:height="row.kind === 'summary' ? 8 : 16"
								class="bar"
								:class="[row.kind, {critical: row.critical}]"
							/>
							<rect
								v-if="row.percent > 0 && row.kind !== 'summary'"
								:x="row.bar.x"
								:y="rowCenter(index) - 8"
								:width="row.bar.width * Math.min(row.percent, 100) / 100"
								height="16"
								class="bar progress"
							/>
						</g>
					</template>

					<path
						v-for="(dep, index) in depPaths"
						:key="`d-${index}`"
						:d="dep.d"
						class="dep"
						:class="{critical: dep.critical}"
					/>
				</svg>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import {computed, nextTick, onMounted, ref} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import dayjs from 'dayjs'

import {allTasksQuery, type TaskFilterParams, type TaskResponse} from '@/client/queries/tasks'
import {queryClient} from '@/client/queryClient'
import {ensureProject} from '@/client/queries/projects'
import {useProjects} from '@/composables/useProjects'
import {useAuthStore} from '@/stores/auth'
import {isGanttZoom, type GanttZoom} from '@/helpers/ganttZoom'
import {getErrorText} from '@/message'
import {capTasks, MAX_EXPORT_TASKS, sortByBoard, type ExportColumnKey} from '@/helpers/export/taskExport'
import {loadTasksForExport, type SortMap} from '@/composables/useViewExport'
import {
	buildPrintBoard,
	buildPrintGantt,
	buildPrintTable,
	type PrintBoard,
	type PrintFormatters,
	type PrintGantt,
	type PrintLabels,
	type PrintTable,
} from '@/helpers/export/printModel'
import {ALL_EXPORT_COLUMNS} from '@/helpers/export/taskExport'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

defineOptions({name: 'ProjectPrint'})

const props = defineProps<{
	projectId: number
	viewId: number
}>()

const route = useRoute()
const {t} = useI18n({useScope: 'global'})
const authStore = useAuthStore()
const projectList = useProjects()

const UPPER_HEIGHT = 20
const HEADER_HEIGHT = 40
const ROW_HEIGHT = 22

const loading = ref(true)
const failed = ref('')
const truncated = ref(false)
const table = ref<PrintTable | null>(null)
const board = ref<PrintBoard | null>(null)
const gantt = ref<PrintGantt | null>(null)

const kind = computed(() => {
	const value = String(route.query.kind ?? 'table')
	return (['table', 'list', 'kanban', 'gantt'] as const).find(k => k === value) ?? 'table'
})

const projectTitle = computed(() => projectList.projects[props.projectId]?.title ?? String(props.projectId))
const viewTitle = computed(() => t(`export.views.${kind.value}`))
const printedOn = computed(() => dayjs().format('L LT'))
const filterText = computed(() => String(route.query.filter ?? '').trim())

const fmt: PrintFormatters = {
	date: d => dayjs(d).format('L'),
	dateTime: d => dayjs(d).format('L LT'),
}

const labels = computed<PrintLabels>(() => ({
	yes: t('export.yes'),
	no: t('export.no'),
	columns: Object.fromEntries(ALL_EXPORT_COLUMNS.map(c => [c.key, t(`export.columns.${c.key}`)])) as PrintLabels['columns'],
}))

function formatDay(d: Date) {
	return fmt.date(d)
}

function rowCenter(index: number) {
	return HEADER_HEIGHT + index * ROW_HEIGHT + ROW_HEIGHT / 2
}

function diamond(x: number, y: number) {
	const h = 7
	return `${x - h},${y} ${x},${y - h} ${x + h},${y} ${x},${y + h}`
}

// Dependency line from the end of one bar to the start of the next, right angles only.
const depPaths = computed(() => {
	const g = gantt.value
	if (!g) return []
	const out: {d: string, critical: boolean}[] = []
	for (const dep of g.deps) {
		const from = g.rows[dep.fromRow]?.bar
		const to = g.rows[dep.toRow]?.bar
		if (!from || !to) continue
		const x1 = from.x + from.width
		const y1 = rowCenter(dep.fromRow)
		const x2 = to.x
		const y2 = rowCenter(dep.toRow)
		const midX = Math.max(x1 + 6, Math.min(x2 - 6, (x1 + x2) / 2))
		out.push({d: `M ${x1} ${y1} H ${midX} V ${y2} H ${x2}`, critical: dep.critical})
	}
	return out
})

function parseSort(raw: unknown): SortMap | undefined {
	if (typeof raw !== 'string' || raw === '') return undefined
	try {
		const parsed = JSON.parse(raw)
		return parsed && typeof parsed === 'object' ? parsed as SortMap : undefined
	} catch {
		return undefined
	}
}

function parseBuckets(raw: unknown): {id: number, title: string}[] {
	if (typeof raw !== 'string' || raw === '') return []
	try {
		const parsed = JSON.parse(raw)
		return Array.isArray(parsed)
			? parsed.filter(b => b && typeof b.id === 'number').map(b => ({id: b.id as number, title: String(b.title ?? '')}))
			: []
	} catch {
		return []
	}
}

function parseColumns(raw: unknown): ExportColumnKey[] {
	const known = new Set<string>(ALL_EXPORT_COLUMNS.map(c => c.key))
	const wanted = String(raw ?? '').split(',').filter(key => known.has(key)) as ExportColumnKey[]
	return wanted.length > 0 ? wanted : ['identifier', 'title', 'assignees', 'due', 'done']
}

function projectTitleOf(id: number) {
	return projectList.projects[id]?.title ?? ''
}

async function load() {
	const q = route.query
	const params: TaskFilterParams = {
		filter: String(q.filter ?? ''),
		q: String(q.q ?? ''),
		filter_include_nulls: q.filter_include_nulls === 'true',
		filter_timezone: authStore.settings.timezone,
	}

	if (kind.value === 'gantt') {
		await loadGantt(params)
		return
	}

	const loaded = await loadTasksForExport(props.projectId, props.viewId, kind.value, params, parseSort(q.sort))
	const capped = capTasks(loaded)
	truncated.value = capped.truncated

	if (kind.value === 'kanban') {
		const buckets = parseBuckets(q.buckets)
		const titles = new Map(buckets.map(b => [b.id, b.title]))
		const lookups = {projectTitle: projectTitleOf, bucketTitle: (id: number) => titles.get(id) ?? ''}
		board.value = buildPrintBoard(sortByBoard(capped.tasks, buckets), buckets, lookups, fmt)
		return
	}

	table.value = buildPrintTable(capped.tasks, parseColumns(q.cols), {projectTitle: projectTitleOf}, fmt, labels.value)
}

async function loadGantt(params: TaskFilterParams) {
	const q = route.query
	const zoom: GanttZoom = isGanttZoom(q.zoom) ? q.zoom : 'fit'
	const from = parseDateOrNull(String(q.dateFrom ?? '')) ?? new Date()
	const to = parseDateOrNull(String(q.dateTo ?? '')) ?? dayjs(from).add(60, 'day').toDate()
	const fromDay = dayjs(from).format('YYYY-MM-DD')
	const toDay = dayjs(to).format('YYYY-MM-DD')

	// The same selection the chart makes: whatever overlaps the date range.
	const ganttParams: TaskFilterParams = {
		...params,
		sort_by: ['start_date', 'done', 'id'],
		order_by: ['asc', 'asc', 'desc'],
		filter: '(' +
			`(start_date >= "${fromDay}" && start_date <= "${toDay}") || ` +
			`(end_date >= "${fromDay}" && end_date <= "${toDay}") || ` +
			`(due_date >= "${fromDay}" && due_date <= "${toDay}") || ` +
			`(start_date <= "${fromDay}" && end_date >= "${toDay}")` +
			')',
		filter_include_nulls: q.withoutDates === 'true',
		expand: ['subtasks'],
	}

	const loaded: TaskResponse[] = await queryClient.fetchQuery({
		...allTasksQuery({project: props.projectId, view: props.viewId, params: ganttParams}),
		staleTime: 0,
	})
	const capped = capTasks(loaded)
	truncated.value = capped.truncated

	const unit = (date: Date, tier: 'year' | 'quarter' | 'month' | 'week') => {
		switch (tier) {
			case 'year': return dayjs(date).format('YYYY')
			case 'quarter': return `Q${Math.floor(date.getMonth() / 3) + 1}`
			case 'month': return dayjs(date).format(zoom === 'month' ? 'MMM' : 'MMMM YYYY')
			case 'week': return dayjs(date).format('D MMM')
		}
	}
	gantt.value = buildPrintGantt(
		new Map(capped.tasks.map(task => [task.id, task])),
		from,
		to,
		zoom,
		q.critical === 'true',
		unit,
		new Date(),
	)
}

function printNow() {
	window.print()
}

onMounted(async () => {
	try {
		await ensureProject(props.projectId)
		await load()
	} catch (e) {
		failed.value = getErrorText(e)
	} finally {
		loading.value = false
	}

	// Print once the data is drawn. The button stays for another go.
	if (!failed.value) {
		await nextTick()
		requestAnimationFrame(() => setTimeout(printNow, 300))
	}
})
</script>

<style lang="scss" scoped>
.print-page {
	padding: 1rem 1.5rem;
	background: #fff;
	color: #000;
	min-block-size: 100vh;
}

.print-header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	gap: 1rem;
	margin-block-end: 1rem;

	h1 {
		font-size: 1.4rem;
		margin: 0;
	}
}

.print-meta,
.print-note {
	margin: 0.25rem 0 0;
	font-size: 0.85rem;
	color: #444;
}

.print-note {
	margin-block-end: 0.75rem;
}

.print-table {
	inline-size: 100%;
	border-collapse: collapse;
	font-size: 0.8rem;

	th,
	td {
		border: 1px solid #bbb;
		padding: 3px 6px;
		text-align: start;
		vertical-align: top;
	}

	th {
		background: #eee;
	}

	// A table that runs over a page repeats its header and never cuts a row in two.
	thead {
		display: table-header-group;
	}

	tr {
		break-inside: avoid;
		page-break-inside: avoid;
	}

	tr.is-done td {
		color: #666;
		text-decoration: line-through;
	}
}

.print-board {
	display: flex;
	gap: 0.75rem;
	align-items: flex-start;
	flex-wrap: wrap;
}

.print-bucket {
	flex: 1 1 14rem;
	min-inline-size: 12rem;
	border: 1px solid #bbb;
	border-radius: 4px;
	padding: 0.5rem;

	h2 {
		font-size: 0.95rem;
		margin: 0 0 0.5rem;

		.count {
			color: #666;
			font-weight: normal;
		}
	}
}

.print-card {
	display: flex;
	flex-direction: column;
	gap: 2px;
	border: 1px solid #ccc;
	border-radius: 3px;
	padding: 4px 6px;
	margin-block-end: 0.4rem;
	font-size: 0.8rem;
	break-inside: avoid;
	page-break-inside: avoid;

	.details {
		color: #555;
		font-size: 0.75rem;
	}

	&.is-done strong {
		text-decoration: line-through;
		color: #666;
	}
}

.print-gantt {
	display: flex;
	align-items: flex-start;
	// A Gantt is one drawing, it is scaled to the page rather than split.
	break-inside: avoid;
	page-break-inside: avoid;
	overflow: visible;

	.gantt-table {
		inline-size: auto;
		flex: none;
		max-inline-size: 48%;

		.head-spacer {
			vertical-align: middle;
		}

		td,
		th {
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
			max-inline-size: 16rem;
		}

		tr.is-summary td {
			font-weight: bold;
		}

		tr.is-critical td:first-child {
			color: #d00;
			font-weight: bold;
		}
	}
}

.gantt-drawing {
	flex: none;
	overflow: visible;

	.head-cell {
		fill: #eee;
		stroke: #bbb;
		stroke-width: 0.5;
	}

	.head-text {
		font-size: 10px;
		fill: #222;

		&.small {
			font-size: 8px;
		}
	}

	.grid-line {
		stroke: #ddd;
		stroke-width: 0.5;
	}

	.today-line {
		stroke: #d00;
		stroke-width: 1;
		stroke-dasharray: 3 2;
	}

	.bar {
		fill: #3b82f6;

		&.summary {
			fill: #374151;
		}

		&.milestone {
			fill: #f59e0b;
		}

		&.critical {
			fill: #ef4444;
		}

		&.progress {
			fill: rgba(0, 0, 0, 0.3);
		}
	}

	.dep {
		fill: none;
		stroke: #666;
		stroke-width: 0.8;

		&.critical {
			stroke: #d00;
			stroke-width: 1.2;
		}
	}
}

@media print {
	.no-print {
		display: none !important;
	}

	.print-page {
		padding: 0;
		min-block-size: 0;
	}

	// Keep the bar colours on paper.
	.gantt-drawing,
	.print-table th {
		print-color-adjust: exact;
		-webkit-print-color-adjust: exact;
	}
}
</style>

<style lang="scss">
// The page size can only be set globally. Landscape, with a narrow margin, for tables and charts.
@media print {
	@page {
		size: landscape;
		margin: 10mm;
	}
}
</style>
