import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import Reminders from './Reminders.vue'
import ReminderDetail from '@/components/tasks/partials/ReminderDetail.vue'

const meta: Meta<typeof Reminders> = {
    title: 'Tasks/Reminders',
    component: Reminders,
}
export default meta

type Story = StoryObj<typeof Reminders>

export const Default: Story = {
    render: () => ({
        components: { Reminders },
        template: '<Reminders />',
    }),
}

export const WithDetails: Story = {
    render: () => ({
        components: { ReminderDetail },
        setup() {
            const reminderNow = ref({ reminder: new Date(), relativePeriod: 0, relativeTo: null })
            const relativeReminder = ref({ reminder: null, relativePeriod: 1, relativeTo: 'due_date' })
            const newReminder = ref(null)
            return { reminderNow, relativeReminder, newReminder }
        },
        template: '<ReminderDetail v-model="reminderNow" />\n<ReminderDetail v-model="relativeReminder" />\n<ReminderDetail v-model="newReminder" />',
    }),
}
