<template>
	<div class="reminder-detail">
		<ReminderPeriod v-if="showRelativeReminder()" v-model="reminder" @update:modelValue="() => updateData()"></ReminderPeriod>
		<Datepicker
				v-if="showAbsoluteReminder()"
				v-model="reminderDate"
				:disabled="disabled"
				@close-on-change="() => setReminderDate()"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref, watch, type PropType } from 'vue'

import Datepicker from '@/components/input/datepicker.vue'
import ReminderPeriod from '@/components/tasks/partials/reminder-period.vue'
import TaskReminderModel from '@/models/taskReminder'
import type { ITaskReminder } from '@/modelTypes/ITaskReminder'

const props = defineProps({
	modelValue: {
		type: Object as PropType<ITaskReminder>,
		required: false,
	},
	disabled: {
		default: false,
	},
})

const emit = defineEmits(['update:modelValue', 'update:Reminder', 'close', 'close-on-change'])

const reminder = ref<ITaskReminder>()
const reminderDate = ref()


watch(
	() => props.modelValue,
	(value) => {
			console.log('reminder-detail.watch', value)
			reminder.value = value
			if (reminder.value && reminder.value.reminder) {
				reminderDate.value = new Date(reminder.value.reminder)
			}
		},
		{immediate: true},
)

function setReminderDate() {
	console.log('reminder-detail.setReminderDate', reminderDate.value)
	console.log('reminder-detail.setReminderDate.reminder', reminder.value)
	if (!reminderDate.value) {
		return
	}
	if (!reminder.value) {
		reminder.value = new TaskReminderModel()
	}
	reminder.value.reminder = new Date(reminderDate.value)
	updateData()
}


function updateData() {
	console.log('reminder-detail.updateData', reminder.value)
	emit('update:modelValue', reminder.value)
}

function showAbsoluteReminder() {
	return !reminder.value || !reminder.value?.relativeTo
}

function showRelativeReminder() {
	console.log('showRelativeReminder', reminder.value)
	return !reminder.value || reminder.value?.relativeTo
}

</script>

<style lang="scss" scoped>
.reminders {
	.reminder-input {
		display: flex;
		align-items: center;

		&.overdue :deep(.datepicker .show) {
			color: var(--danger);
		}

		&:last-child {
			margin-bottom: 0.75rem;
		}

		.remove {
			color: var(--danger);
			padding-left: .5rem;
		}
	}
}

</style>