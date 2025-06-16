import type { Meta, StoryObj } from '@storybook/vue3'
import TaskForm from './TaskForm.vue'

const meta: Meta<typeof TaskForm> = {
    title: 'Tasks/TaskForm',
    component: TaskForm,
}
export default meta

type Story = StoryObj<typeof TaskForm>

export const Default: Story = {
    render: () => ({
        components: { TaskForm },
        template: '<TaskForm />',
    }),
}
