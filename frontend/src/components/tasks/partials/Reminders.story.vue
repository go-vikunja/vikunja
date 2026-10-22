<script setup lang="ts">
import reminders from './Reminders.vue'
import type {TaskReminder as ITaskReminder} from '@/client/generated'
import {ref} from 'vue'
import ReminderDetail from '@/components/tasks/partials/ReminderDetail.vue'
import {createReminderDraft} from '@/helpers/task'

const reminderNow = ref(createReminderDraft({reminder: new Date().toISOString()}))
const relativeReminder = ref(createReminderDraft({relative_period: 1, relative_to: 'due_date'}))
const newReminder = ref<ITaskReminder>()
</script>

<template>
	<Story>
		<Variant title="Default">
			<reminders />
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
