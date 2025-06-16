import type { Meta, StoryObj } from '@storybook/vue3'
import AddTask from './AddTask.vue'
import { useAuthStore } from '@/stores/auth'
import { useTaskStore } from '@/stores/tasks'
import TaskModel from '@/models/task'
import type { ITask } from '@/modelTypes/ITask'

const meta: Meta<typeof AddTask> = {
    title: 'Tasks/AddTask',
    component: AddTask,
}
export default meta

type Story = StoryObj<typeof AddTask>

export const Default: Story = {
    render: () => ({
        components: { AddTask },
        setup() {
            const authStore = useAuthStore()
            Object.assign(authStore.settings, {
                defaultProjectId: 1,
                frontendSettings: { quickAddMagicMode: 'prefix' },
            })

            const taskStore = useTaskStore()
            taskStore.isLoading.value = false
            taskStore.ensureLabelsExist = async () => {}
            taskStore.findProjectId = async () => 1
            taskStore.createNewTask = async (data: Partial<ITask>) => new TaskModel({ id: Date.now(), ...data })
            return {}
        },
        template: '<AddTask />',
    }),
}
