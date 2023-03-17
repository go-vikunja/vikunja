<template>
	<div>
		<ReminderPeriod v-if="showRelativeReminder" v-model="reminder" :disabled="disabled"
										@update:modelValue="() => updateData()"></ReminderPeriod>
		<Datepicker
				v-if="showAbsoluteReminder"
				v-model="reminderDate"
				:disabled="disabled"
				@close-on-change="() => setReminderDate()"
		/>
	</div>
</template>

<script setup lang="ts">
import Datepicker from '@/components/input/datepicker.vue'
import ReminderPeriod from '@/components/tasks/partials/reminder-period.vue'
import TaskReminderModel from '@/models/taskReminder'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import {computed, ref, watch, type PropType} from 'vue'

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

const reminder = ref<ITaskReminder>()
const reminderDate = ref()

const showAbsoluteReminder = computed(() => !reminder.value || !reminder.value?.relativeTo)
const showRelativeReminder = computed(() => !reminder.value || reminder.value?.relativeTo)

watch(
		() => props.modelValue,
		(value) => {
			reminder.value = value
			if (reminder.value && reminder.value.reminder) {
				reminderDate.value = new Date(reminder.value.reminder)
			}
		},
		{immediate: true},
)

function updateData() {
	emit('update:modelValue', reminder.value)
}

function setReminderDate() {
	if (!reminderDate.value) {
		return
	}
	if (!reminder.value) {
		reminder.value = new TaskReminderModel()
	}
	reminder.value.reminder = new Date(reminderDate.value)
	updateData()
}
</script>
