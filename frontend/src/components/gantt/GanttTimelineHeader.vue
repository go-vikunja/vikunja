<template>
	<div class="gantt-timeline">
		<!-- Upper timeunit for months -->
		<div class="gantt-timeline-months">
			<div
				v-for="monthGroup in monthGroups"
				:key="monthGroup.key"
				class="timeunit-month"
				:style="{ width: `${monthGroup.width}px` }"
			>
				{{ monthGroup.label }}
			</div>
		</div>
        
		<!-- Lower timeunit for days -->
		<div class="gantt-timeline-days">
			<div
				v-for="date in timelineData"
				:key="date.toISOString()"
				class="timeunit"
				:style="{ width: `${dayWidthPixels}px` }"
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
import { computed } from 'vue'
import { useGlobalNow } from '@/composables/useGlobalNow'
import { useWeekDayFromDate } from '@/helpers/time/formatDate'
import dayjs from 'dayjs'

const props = defineProps<{
    timelineData: Date[]
    dayWidthPixels: number
}>()

const weekDayFromDate = useWeekDayFromDate()
const { now: today } = useGlobalNow()

const dateIsToday = computed(() => (date: Date) => {
	return (
		date.getDate() === today.value.getDate() &&
        date.getMonth() === today.value.getMonth() &&
        date.getFullYear() === today.value.getFullYear()
	)
})

const monthGroups = computed(() => {
	const groups: Array<{key: string; label: string; width: number}> = []
	let currentMonth = -1
	let currentYear = -1
	let dayCount = 0
    
	props.timelineData.forEach((date, index) => {
		const month = date.getMonth()
		const year = date.getFullYear()
        
		if (month !== currentMonth || year !== currentYear) {
			// Finish previous group
			if (currentMonth !== -1) {
				groups[groups.length - 1].width = dayCount * props.dayWidthPixels
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
		if (index === props.timelineData.length - 1) {
			groups[groups.length - 1].width = dayCount * props.dayWidthPixels
		}
	})
    
	return groups
})
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
