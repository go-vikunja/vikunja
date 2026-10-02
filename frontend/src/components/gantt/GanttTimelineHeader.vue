<template>
	<div
		class="gantt-timeline"
		role="columnheader"
		:aria-label="$t('project.gantt.timelineHeader')"
	>
		<!-- Upper timeunit: months, or years at the coarser zooms -->
		<div
			class="gantt-timeline-months"
			role="row"
			:aria-label="$t('project.gantt.monthsRow')"
		>
			<div
				v-for="group in tiers.upper"
				:key="group.key"
				class="timeunit-month"
				:style="{ width: `${group.width}px` }"
				role="columnheader"
				:aria-label="$t('project.gantt.monthLabel', {month: group.label})"
			>
				{{ group.label }}
			</div>
		</div>

		<!-- Lower timeunit for weeks, months or quarters -->
		<div
			v-if="tiers.lower.length > 0"
			class="gantt-timeline-days gantt-timeline-coarse"
			role="row"
		>
			<div
				v-for="cell in tiers.lower"
				:key="cell.key"
				class="timeunit-coarse"
				:class="{'today': coarseCellIsToday(cell)}"
				:style="{ width: `${cell.width}px` }"
				role="columnheader"
			>
				{{ cell.label }}
			</div>
		</div>

		<!-- Lower timeunit for days -->
		<div
			v-else
			class="gantt-timeline-days"
			role="row"
			:aria-label="$t('project.gantt.daysRow')"
		>
			<div
				v-for="date in timelineData"
				:key="date.toISOString()"
				class="timeunit"
				:style="{ width: `${dayWidthPixels}px` }"
				role="columnheader"
				:aria-label="dateIsToday(date) 
					? $t('project.gantt.dayLabelToday', {
						date: date.toLocaleDateString(),
						weekday: weekDayFromDate(date)
					})
					: $t('project.gantt.dayLabel', {
						date: date.toLocaleDateString(),
						weekday: weekDayFromDate(date)
					})"
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
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useGlobalNow} from '@/composables/useGlobalNow'
import {useWeekDayFromDate} from '@/helpers/time/formatDate'
import dayjs from 'dayjs'
import {buildTimelineTiers, type GanttZoom, type TierUnit, type TimelineCell} from '@/helpers/ganttZoom'

const props = withDefaults(defineProps<{
    timelineData: Date[]
    dayWidthPixels: number
    zoom?: GanttZoom
}>(), {
	zoom: 'day',
})

const weekDayFromDate = useWeekDayFromDate()
const { now: today } = useGlobalNow()

const dateIsToday = computed(() => {
	const todayStr = today.value.toDateString()
	return (date: Date) => date.toDateString() === todayStr
})

function formatCell(date: Date, unit: TierUnit): string {
	switch (unit) {
		case 'year': return dayjs(date).format('YYYY')
		case 'quarter': return `Q${Math.floor(date.getMonth() / 3) + 1}`
		case 'month': return dayjs(date).format(props.zoom === 'month' ? 'MMM' : 'MMMM YYYY')
		case 'week': return dayjs(date).format('D MMM')
	}
}

const tiers = computed(() => buildTimelineTiers(props.timelineData, props.dayWidthPixels, props.zoom, formatCell))

// The cell that contains today gets the same highlight the day cell has.
function coarseCellIsToday(cell: TimelineCell): boolean {
	const now = today.value
	const start = cell.start.getTime()
	const end = start + (cell.width / props.dayWidthPixels) * 24 * 60 * 60 * 1000
	return now.getTime() >= start && now.getTime() < end
}
</script>

<style scoped lang="scss">
// The task table next to the chart uses the same two heights, keep them in sync with
// GANTT_HEADER_TIER_PX in GanttTaskPane.vue.
.gantt-timeline {
	background: var(--white);
	border-block-end: 1px solid var(--grey-200);
	position: sticky;
	inset-block-start: 0;
	z-index: 10;
}

.gantt-timeline-months {
	display: flex;
	block-size: 32px;

	.timeunit-month {
		background: var(--white);
		font-family: $vikunja-font;
		font-weight: bold;
		border-inline-end: 1px solid var(--grey-200);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		white-space: nowrap;
		font-size: 1rem;
		color: var(--grey-800);
	}
}

.gantt-timeline-days {
	display: flex;
	block-size: 52px;

	.timeunit {
		.timeunit-wrapper {
			padding: 0.35rem 0;
			font-size: 1rem;
			display: flex;
			flex-direction: column;
			align-items: center;
			inline-size: 100%;
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

	.timeunit-coarse {
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		white-space: nowrap;
		font-family: $vikunja-font;
		font-size: 0.85rem;
		color: var(--grey-700);
		border-inline-end: 1px solid var(--grey-200);

		&.today {
			background: var(--primary);
			color: var(--white);
			font-weight: bold;
		}
	}
}
</style>
