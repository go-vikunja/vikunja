<template>
	<div
		class="gantt-timeline"
		role="columnheader"
		:aria-label="$t('project.gantt.timelineHeader')"
	>
		<div
			class="gantt-timeline-months"
			role="row"
			:aria-label="units.upper === 'year' ? $t('project.gantt.yearsRow') : $t('project.gantt.monthsRow')"
		>
			<div
				v-for="group in upperGroups"
				:key="group.key"
				class="timeunit-month"
				:style="{ width: `${group.dayCount * dayWidthPixels}px` }"
				role="columnheader"
				:aria-label="units.upper === 'year'
					? $t('project.gantt.yearLabel', {year: group.label})
					: $t('project.gantt.monthLabel', {month: group.label})"
			>
				{{ group.label }}
			</div>
		</div>

		<div
			v-if="units.lower === 'day'"
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
					<span
						v-if="units.showWeekday"
						class="weekday"
					>
						{{ weekDayFromDate(date) }}
					</span>
				</div>
			</div>
		</div>

		<div
			v-else
			class="gantt-timeline-days"
			role="row"
			:aria-label="units.lower === 'week' ? $t('project.gantt.weeksRow') : $t('project.gantt.monthsRow')"
		>
			<div
				v-for="group in lowerGroups"
				:key="group.key"
				class="timeunit"
				:style="{ width: `${group.dayCount * dayWidthPixels}px` }"
				role="columnheader"
				:aria-label="group.ariaLabel"
			>
				<div
					class="timeunit-wrapper is-grouped"
					:class="{'today': groupContainsToday(group)}"
				>
					<span>{{ group.label }}</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import dayjs from 'dayjs'

import {useGlobalNow} from '@/composables/useGlobalNow'
import {useWeekDayFromDate} from '@/helpers/time/formatDate'
import {
	groupTimelineDates,
	isoWeekNumber,
	timelineUnitsForDayWidth,
	type GanttTimelineGroup,
	type GanttTimelineUnit,
} from '@/helpers/time/ganttTimeline'

const props = defineProps<{
	timelineData: Date[]
	dayWidthPixels: number
}>()

const {t} = useI18n({useScope: 'global'})
const weekDayFromDate = useWeekDayFromDate()
const {now: today} = useGlobalNow()

const units = computed(() => timelineUnitsForDayWidth(props.dayWidthPixels))

const dateIsToday = computed(() => {
	const todayStr = today.value.toDateString()
	return (date: Date) => date.toDateString() === todayStr
})

function groupContainsToday(group: GanttTimelineGroup) {
	const start = dayjs(group.start).startOf('day')
	const end = start.add(group.dayCount, 'day')
	return !dayjs(today.value).isBefore(start) && dayjs(today.value).isBefore(end)
}

function labelFor(group: GanttTimelineGroup, unit: GanttTimelineUnit): string {
	switch (unit) {
		case 'week':
			return t('project.gantt.weekShort', {week: isoWeekNumber(group.start)})
		case 'month':
			return dayjs(group.start).format(units.value.upper === 'year' ? 'MMM' : 'MMMM YYYY')
		case 'year':
			return dayjs(group.start).format('YYYY')
		case 'day':
			return String(group.start.getDate())
	}
}

function ariaLabelFor(group: GanttTimelineGroup, unit: GanttTimelineUnit): string {
	if (unit === 'week') {
		return t('project.gantt.weekLabel', {
			week: isoWeekNumber(group.start),
			start: group.start.toLocaleDateString(),
		})
	}
	return t('project.gantt.monthLabel', {month: dayjs(group.start).format('MMMM YYYY')})
}

const upperGroups = computed(() => groupTimelineDates(props.timelineData, units.value.upper)
	.map(group => ({...group, label: labelFor(group, units.value.upper)})))

const lowerGroups = computed(() => groupTimelineDates(props.timelineData, units.value.lower)
	.map(group => ({
		...group,
		label: labelFor(group, units.value.lower),
		ariaLabel: ariaLabelFor(group, units.value.lower),
	})))
</script>

<style scoped lang="scss">
.gantt-timeline {
	background: var(--white);
	border-block-end: 1px solid var(--grey-200);
	position: sticky;
	inset-block-start: 0;
	z-index: 10;
}

.gantt-timeline-months {
	display: flex;

	.timeunit-month {
		background: var(--white);
		font-family: $vikunja-font;
		font-weight: bold;
		border-inline-end: 1px solid var(--grey-200);
		padding: 0.5rem 0;
		text-align: center;
		font-size: 1rem;
		color: var(--grey-800);
		overflow: hidden;
		white-space: nowrap;
	}
}

.gantt-timeline-days {
	display: flex;

	.timeunit {
		.timeunit-wrapper {
			padding: 0.5rem 0;
			font-size: 1rem;
			display: flex;
			flex-direction: column;
			align-items: center;
			inline-size: 100%;
			font-family: $vikunja-font;

			&.is-grouped {
				font-size: 0.8rem;
				overflow: hidden;
				white-space: nowrap;
				border-inline-end: 1px solid var(--grey-200);
			}

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
</style>
