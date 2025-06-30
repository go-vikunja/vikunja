<script setup lang="ts">
import reminders from './Reminders.vue'
import {ref} from 'vue'
import ReminderDetail from '@/components/tasks/partials/ReminderDetail.vue'
import type {ITask} from '@/modelTypes/ITask'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'

const reminderNow = ref({reminder: new Date(), relativePeriod: 0, relativeTo: null, maxRight: null } as ITaskReminder)
const relativeReminder = ref({reminder: null, relativePeriod: 1, relativeTo: 'due_date', maxRight: null } as ITaskReminder)
const newReminder = ref(null as ITaskReminder | null)
const mockTask = ref({} as ITask)
</script>

<template>
	<Story>
		<Variant title="Default">
			<reminders :model-value="mockTask" />
		</Variant>
		<Variant title="Reminder Detail with fixed date">
			<ReminderDetail v-model="reminderNow" />
		</Variant>
		<Variant title="Reminder Detail with relative date">
			<ReminderDetail v-model="relativeReminder" />
		</Variant>
		<Variant title="New Reminder Detail">
			<ReminderDetail v-model="newReminder" />
		</Variant>
	</Story>
</template>
