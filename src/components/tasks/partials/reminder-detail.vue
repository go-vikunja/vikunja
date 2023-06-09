<template>
	<div>

		{{ reminderText }}

		<div class="options" v-if="showFormSwitch === null">
			<BaseButton
				v-for="p in presets"
			>
				{{ formatReminder(p) }}
			</BaseButton>
			<BaseButton @click="showFormSwitch = 'relative'">
				Custom
			</BaseButton>
			<BaseButton @click="showFormSwitch = 'absolute'">
				Date
			</BaseButton>
		</div>

		<ReminderPeriod
			v-if="showFormSwitch === 'relative'"
			v-model="reminder"
			@update:modelValue="emit('update:modelValue', reminder)"
		/>

		<DatepickerInline
			v-if="showFormSwitch === 'absolute'"
			v-model="reminderDate"
			@update:modelValue="setReminderDate"
		/>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, watch, type PropType} from 'vue'
import {toRef} from '@vueuse/core'
import {SECONDS_A_DAY} from '@/constants/date'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'

import {secondsToPeriod} from '@/helpers/time/period'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {formatDateShort} from '@/helpers/time/formatDate'

import BaseButton from '@/components/base/BaseButton.vue'
import DatepickerInline from '@/components/input/datepickerInline.vue'
import ReminderPeriod from '@/components/tasks/partials/reminder-period.vue'

import TaskReminderModel from '@/models/taskReminder'

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
const reminderDate = ref(null)

const showFormSwitch = ref<null | 'relative' | 'absolute'>(null)

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
	reminder.value.reminder = reminderDate.value === null
		? null
		: new Date(reminderDate.value)
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

	return periodHuman + ' ' + (reminder.relativePeriod <= 0 ? 'before' : 'after') + ' ' + reminder.relativeTo
}
</script>

<style lang="scss" scoped>
.options {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
}
</style>
