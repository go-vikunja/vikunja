<template>
	<Loading
		v-if="props.isLoading && !ganttBars.length || dayjsLanguageLoading"
		class="gantt-container"
	/>
	<div ref="ganttContainer" class="gantt-container" v-else>
		<GGanttChart
			:date-format="DAYJS_ISO_DATE_FORMAT"
			:chart-start="isoToKebabDate(filters.dateFrom)"
			:chart-end="isoToKebabDate(filters.dateTo)"
			precision="day"
			bar-start="startDate"
			bar-end="endDate"
			:grid="true"
			@dragend-bar="updateGanttTask"
			@dblclick-bar="openTask"
			:width="ganttChartWidth"
		>
			<template #timeunit="{value, date}">
				<div
					class="timeunit-wrapper"
					:class="{'today': dateIsToday(date)}"
				>
					<span>{{ value }}</span>
					<span class="weekday">
						{{ weekDayFromDate(date) }}
					</span>
				</div>
			</template>
			<GGanttRow
				v-for="(bar, k) in ganttBars"
				:key="k"
				label=""
				:bars="bar"
			/>
		</GGanttChart>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, toRefs, onActivated} from 'vue'
import {useRouter} from 'vue-router'

import {getHexColor} from '@/models/task'

import {colorIsDark} from '@/helpers/color/colorIsDark'
import {isoToKebabDate} from '@/helpers/time/isoToKebabDate'
import {parseKebabDate} from '@/helpers/time/parseKebabDate'

import type {ITask, ITaskPartialWithId} from '@/modelTypes/ITask'
import type {DateISO} from '@/types/DateISO'
import type {GanttFilters} from '@/views/project/helpers/useGanttFilters'

import {
	extendDayjs,
	GGanttChart,
	GGanttRow,
	type GanttBarObject,
} from '@infectoone/vue-ganttastic'

import Loading from '@/components/misc/loading.vue'
import {MILLISECONDS_A_DAY} from '@/constants/date'
import {useWeekDayFromDate} from '@/helpers/time/formatDate'

export interface GanttChartProps {
	isLoading: boolean,
	filters: GanttFilters,
	tasks: Map<ITask['id'], ITask>,
	defaultTaskStartDate: DateISO
	defaultTaskEndDate: DateISO
}

const DAYJS_ISO_DATE_FORMAT = 'YYYY-MM-DD'

const props = defineProps<GanttChartProps>()

const emit = defineEmits<{
  (e: 'update:task', task: ITaskPartialWithId): void
}>()

const {tasks, filters} = toRefs(props)

// setup dayjs for vue-ganttastic
const dayjsLanguageLoading = ref(false)
// const dayjsLanguageLoading = useDayjsLanguageSync(dayjs)
extendDayjs()

const ganttContainer = ref(null)

const router = useRouter()

const dateFromDate = computed(() => new Date(new Date(filters.value.dateFrom).setHours(0,0,0,0)))
const dateToDate = computed(() => new Date(new Date(filters.value.dateTo).setHours(23,59,0,0)))

const DAY_WIDTH_PIXELS = 30
const ganttChartWidth = computed(() => {

	const ganttContainerReference = ganttContainer?.value
	const ganttContainerWidth = ganttContainerReference ? (ganttContainerReference['clientWidth'] ?? 0) : 0

	const dateDiff = Math.floor((dateToDate.value.valueOf() - dateFromDate.value.valueOf()) / MILLISECONDS_A_DAY)
	const calculatedWidth = dateDiff * DAY_WIDTH_PIXELS

	return (calculatedWidth > ganttContainerWidth) ? calculatedWidth + 'px' : '100%'

})

const ganttBars = ref<GanttBarObject[][]>([])

/**
 * Update ganttBars when tasks change
 */
watch(
	tasks,
	() => {
		ganttBars.value = []
		tasks.value.forEach(t => ganttBars.value.push(transformTaskToGanttBar(t)))
	},
	{deep: true, immediate: true},
)

function transformTaskToGanttBar(t: ITask) {
	const black = 'var(--grey-800)'
	return [{
		startDate: isoToKebabDate(t.startDate ? t.startDate.toISOString() : props.defaultTaskStartDate),
		endDate: isoToKebabDate(t.endDate ? t.endDate.toISOString() : props.defaultTaskEndDate),
		ganttBarConfig: {
			id: String(t.id),
			label: t.title,
			hasHandles: true,
			style: {
				color: t.startDate ? (colorIsDark(getHexColor(t.hexColor)) ? black : 'white') : black,
				backgroundColor: t.startDate ? getHexColor(t.hexColor) : 'var(--grey-100)',
				border: t.startDate ? '' : '2px dashed var(--grey-300)',
				'text-decoration': t.done ? 'line-through' : null,
			},
		},
	} as GanttBarObject]
}

async function updateGanttTask(e: {
	bar: GanttBarObject;
	e: MouseEvent;
	datetime?: string | undefined;
}) {
	emit('update:task', {
		id: Number(e.bar.ganttBarConfig.id),
		startDate: new Date(parseKebabDate(e.bar.startDate).setHours(0,0,0,0)),
		endDate: new Date(parseKebabDate(e.bar.endDate).setHours(23,59,0,0)),
	})
}

function openTask(e: {
    bar: GanttBarObject;
    e: MouseEvent;
    datetime?: string | undefined;
}) {
	router.push({
		name: 'task.detail',
		params: {id: e.bar.ganttBarConfig.id},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

const weekDayFromDate = useWeekDayFromDate()

const today = ref(new Date())
onActivated(() => today.value = new Date())
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
}
</style>
	

<style lang="scss">
// Not scoped because we need to style the elements inside the gantt chart component
.g-gantt-chart {
	width: max-content;
}

.g-gantt-row-label {
	display: none !important;
}

.g-upper-timeunit, .g-timeunit {
	background: var(--white) !important;
	font-family: $vikunja-font;
}

.g-upper-timeunit {
	font-weight: bold;
	border-right: 1px solid var(--grey-200);
	padding: .5rem 0;
}

.g-timeunit .timeunit-wrapper {
	padding: 0.5rem 0;
	font-size: 1rem !important;
	display: flex;
	flex-direction: column;
	align-items: center;
	width: 100%;

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

.g-timeaxis {
	height: auto !important;
	box-shadow: none !important;
}

.g-gantt-row > .g-gantt-row-bars-container {
	border-bottom: none !important;
	border-top: none !important;
}

.g-gantt-row:nth-child(odd) {
	background: hsla(var(--grey-100-hsl), .5);
}

.g-gantt-bar {
	border-radius: $radius * 1.5;
	overflow: visible;
	font-size: .85rem;

	&-handle-left,
	&-handle-right {
		width: 6px !important;
		height: 75% !important;
		opacity: .75 !important;
		border-radius: $radius !important;
		margin-top: 4px;
	}
}
</style>