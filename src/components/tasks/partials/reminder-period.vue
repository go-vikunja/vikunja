<template>
	<div
		class="reminder-period control"
	>
		<input
			class="input"
			v-model.number="period.duration"
			type="number"
			min="0"
			@change="updateData"
		/>

		<div class="select">
			<select v-model="period.durationUnit" @change="updateData">
				<option value="minutes">{{ $t('time.units.minutes', period.duration) }}</option>
				<option value="hours">{{ $t('time.units.hours', period.duration) }}</option>
				<option value="days">{{ $t('time.units.days', period.duration) }}</option>
				<option value="weeks">{{ $t('time.units.weeks', period.duration) }}</option>
			</select>
		</div>

		<div class="select">
			<select v-model.number="period.sign" @change="updateData">
				<option value="-1">
					{{ $t('task.reminder.beforeShort') }}
				</option>
				<option value="1">
					{{ $t('task.reminder.afterShort') }}
				</option>
			</select>
		</div>

		<div class="select">
			<select v-model="period.relativeTo" @change="updateData">
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
import {ref, watch, type PropType} from 'vue'
import {useI18n} from 'vue-i18n'
import {toRef} from '@vueuse/core'

import {periodToSeconds, PeriodUnit, secondsToPeriod} from '@/helpers/time/period'

import TaskReminderModel from '@/models/taskReminder'

import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES, type IReminderPeriodRelativeTo} from '@/types/IReminderPeriodRelativeTo'

const {t} = useI18n({useScope: 'global'})

const props = defineProps({
	modelValue: {
		type: Object as PropType<ITaskReminder>,
		required: false,
	},
	disabled: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits(['update:modelValue'])

const reminder = ref<ITaskReminder>(new TaskReminderModel())

const showForm = ref(false)

interface PeriodInput {
	duration: number,
	durationUnit: PeriodUnit,
	relativeTo: IReminderPeriodRelativeTo,
	sign: -1 | 1,
}

const period = ref<PeriodInput>({
	duration: 0,
	durationUnit: 'hours',
	relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
	sign: -1,
})

const modelValue = toRef(props, 'modelValue')
watch(
	modelValue,
	(value) => {
		const p = secondsToPeriod(value?.relativePeriod)
		period.value.durationUnit = p.unit
		period.value.duration = p.amount
		period.value.relativeTo = value?.relativeTo || REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE
	},
	{immediate: true},
)

function updateData() {
	reminder.value.relativePeriod = period.value.sign * periodToSeconds(period.value.duration, period.value.durationUnit)
	reminder.value.relativeTo = period.value.relativeTo
	reminder.value.reminder = null

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
		width: 100% !important;
		height: auto;
	}
}
</style>