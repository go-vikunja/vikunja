<template>
	<div
		class="reminder-period control"
	>
		<input
			v-model.number="period.duration"
			class="input"
			type="number"
			min="0"
			@change="updateData"
		>

		<div class="select">
			<select
				v-model="period.durationUnit"
				:aria-label="$t('task.reminder.periodUnit')"
				@change="updateData"
			>
				<option value="minutes">
					{{ $t('time.units.minutes', period.duration) }}
				</option>
				<option value="hours">
					{{ $t('time.units.hours', period.duration) }}
				</option>
				<option value="days">
					{{ $t('time.units.days', period.duration) }}
				</option>
				<option value="weeks">
					{{ $t('time.units.weeks', period.duration) }}
				</option>
			</select>
		</div>

		<div class="select">
			<select
				v-model.number="period.sign"
				:aria-label="$t('task.reminder.periodDirection')"
				@change="updateData"
			>
				<option value="-1">
					{{ $t('task.reminder.beforeShort') }}
				</option>
				<option value="1">
					{{ $t('task.reminder.afterShort') }}
				</option>
			</select>
		</div>

		<div
			v-if="lockRelativeTo === null"
			class="select"
		>
			<select
				v-model="period.relative_to"
				:aria-label="$t('task.reminder.periodRelativeTo')"
				@change="updateData"
			>
				<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE">
					{{ $t('task.attributes.dueDate') }}
				</option>
				<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE">
					{{ $t('task.attributes.startDate') }}
				</option>
				<option :value="REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE">
					{{ $t('task.attributes.endDate') }}
				</option>
			</select>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'

import {periodToSeconds, type PeriodUnit, secondsToPeriod} from '@/helpers/time/period'
import {createReminderDraft} from '@/helpers/task'


import type {TaskReminder as ITaskReminder} from '@/client/generated'
import {type IReminderPeriodRelativeTo, REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'

const props = withDefaults(defineProps<{
	modelValue: ITaskReminder,
	lockRelativeTo?: IReminderPeriodRelativeTo | null,
}>(), {
	lockRelativeTo: null,
})

const emit = defineEmits<{
	'update:modelValue': [ITaskReminder]
}>()

const reminder = ref<ITaskReminder>(createReminderDraft())

interface PeriodInput {
	duration: number,
	durationUnit: PeriodUnit,
	relative_to: string,
	sign: -1 | 1,
}

const period = ref<PeriodInput>({
	duration: 0,
	durationUnit: 'hours',
	relative_to: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
	sign: -1,
})

watch(
	() => props.modelValue,
	(value) => {
		const p = secondsToPeriod(value?.relative_period ?? 0)
		period.value.durationUnit = p.unit
		period.value.duration = Math.abs(p.amount)
		period.value.relative_to = props.lockRelativeTo || value?.relative_to || REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE
	},
	{
		immediate: true,
		deep: true,
	},
)

watch(
	() => period.value.duration,
	value => {
		if (value < 0) {
			period.value.duration = value * -1
		}
	},
)

function updateData() {
	reminder.value.relative_period = period.value.sign * periodToSeconds(Math.abs(period.value.duration), period.value.durationUnit)
	reminder.value.relative_to = period.value.relative_to
	reminder.value.reminder = ''

	emit('update:modelValue', reminder.value)
}
</script>

<style lang="scss" scoped>
.reminder-period {
	display: flex;
	flex-direction: column;
	gap: .25rem;
	padding: .5rem .5rem 0;

	.input, .select select {
		inline-size: 100% !important;
		block-size: auto;
	}
}
</style>
