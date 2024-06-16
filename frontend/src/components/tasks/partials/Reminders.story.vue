<script setup lang="ts">
import reminders from './Reminders.vue'
import {ref} from 'vue'
import ReminderDetail from '@/components/tasks/partials/ReminderDetail.vue'

const reminderNow = ref({reminder: new Date(), relativePeriod: 0, relativeTo: null } )
const relativeReminder = ref({reminder: null, relativePeriod: 1, relativeTo: 'due_date' } )
const newReminder = ref(null)
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
