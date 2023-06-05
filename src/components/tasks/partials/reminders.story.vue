<script setup lang="ts">
import reminders from './reminders.vue'
import {ref} from 'vue'
import ReminderDetail from '@/components/tasks/partials/reminder-detail.vue'

const reminderNow = ref({reminder: new Date(), relativePeriod: 0, relativeTo: null } )
const relativeReminder = ref({reminder: null, relativePeriod: 1, relativeTo: 'due_date' } )
const newReminder = ref(null)
</script>

<template>
	<Story>
		<Variant title="Default">
			<reminders/>
		</Variant>
		<Variant title="Reminder Detail with fixed date">
			<reminder-detail v-model="reminderNow"/>
		</Variant>
		<Variant title="Reminder Detail with relative date">
			<reminder-detail v-model="relativeReminder"/>
		</Variant>
		<Variant title="New Reminder Detail">
			<reminder-detail v-model="newReminder"/>
		</Variant>
	</Story>
</template>
