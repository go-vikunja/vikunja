<template>
	<div>

		{{ reminderText }}

		<div class="presets">
			<BaseButton
				v-for="p in presets"
			>
				{{ formatReminder(p) }}
			</BaseButton>
			<BaseButton>
				Custom
			</BaseButton>
		</div>

		<ReminderPeriod
			v-if="showRelativeReminder"
			v-model="reminder"
			:disabled="disabled"
			@update:modelValue="emit('update:modelValue', reminder.value)"
		/>

		<Datepicker
			v-if="showAbsoluteReminder"
			v-model="reminderDate"
			:disabled="disabled"
			@close-on-change="setReminderDate"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, type PropType} from 'vue'
import {toRef} from '@vueuse/core'
import {SECONDS_A_DAY} from '@/constants/date'
import {secondsToPeriod} from '@/helpers/time/period'

import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {formatDateShort} from '@/helpers/time/formatDate'

import Datepicker from '@/components/input/datepicker.vue'
import ReminderPeriod from '@/components/tasks/partials/reminder-period.vue'
import TaskReminderModel from '@/models/taskReminder'
import BaseButton from '@/components/base/BaseButton.vue'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'

const props = defineProps({
	modelValue: {
		type: Object as PropType<ITaskReminder>,
		required: false,
	},
	disabled: {
		default: false,
	},
})

const emit = defineEmits(['update:modelValue'])

const reminder = ref<ITaskReminder>(new TaskReminderModel())

const presets: TaskReminderModel[] = [
	{relativePeriod: SECONDS_A_DAY, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE},
	{relativePeriod: SECONDS_A_DAY * 3, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE},
	{relativePeriod: SECONDS_A_DAY * 7, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE},
	{relativePeriod: SECONDS_A_DAY * 30, relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE},
]
const reminderDate = computed({
	get() {
		return reminder.value?.reminder
	},
	set(newReminderDate) {
		if (!reminderDate.value) {
			return
		}
		reminder.value.reminder = new Date(reminderDate.value)
	},
})

const showAbsoluteReminder = computed(() => !reminder.value || !reminder.value?.relativeTo)
const showRelativeReminder = computed(() => !reminder.value || reminder.value?.relativeTo)

const reminderText = computed(() => {

	if (reminder.value.reminder !== null) {
		return formatDateShort(reminder.value.reminder)
	}

	if (reminder.value.relativeTo !== null) {
		return formatReminder(reminder.value)
	}

	return 'Add a reminder…'
})

const modelValue = toRef(props, 'modelValue')
watch(
	modelValue,
	(newReminder) => {
		reminder.value = newReminder || new TaskReminderModel()
	},
	{immediate: true},
)

function setReminderDate() {
	if (!reminderDate.value) {
		return
	}
	reminder.value.reminder = new Date(reminderDate.value)
	emit('update:modelValue', reminder.value)
}

function formatReminder(reminder: TaskReminderModel) {

	const period = secondsToPeriod(reminder.relativePeriod)
	let periodHuman = ''

	if (period.days > 0) {
		periodHuman = period.days + ' days'
	}

	if (period.days === 1) {
		periodHuman = period.days + ' day'
	}

	return periodHuman + ' ' + (reminder.relativePeriod > 0 ? 'before' : 'after') + ' ' + reminder.relativeTo
}
</script>

<style lang="scss" scoped>
.presets {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
}
</style>
