<template>
	<div>
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

import type {ITaskReminder} from '@/modelTypes/ITaskReminder'

import Datepicker from '@/components/input/datepicker.vue'
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
const reminderDate = computed({
	get() {
		return reminder.value?.reminder
	},
	set(newReminderDate) {
		if (!reminderDate.value) {
			return
		}
		reminder.value.reminder = new Date(reminderDate.value)
	}
})

const showAbsoluteReminder = computed(() => !reminder.value || !reminder.value?.relativeTo)
const showRelativeReminder = computed(() => !reminder.value || reminder.value?.relativeTo)

watch(
		props.modelValue,
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
</script>
