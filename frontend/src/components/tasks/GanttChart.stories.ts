import type { Meta, StoryObj } from '@storybook/vue3'
import GanttChart from './GanttChart.vue'
import TaskModel from '@/models/task'
import type { GanttFilters } from '@/views/project/helpers/useGanttFilters'

const meta: Meta<typeof GanttChart> = {
    title: 'Tasks/GanttChart',
    component: GanttChart,
}
export default meta

type Story = StoryObj<typeof GanttChart>

export const Default: Story = {
    render: () => ({
        components: { GanttChart },
        setup() {
            const filters: GanttFilters = {
                projectId: 1,
                viewId: 1,
                dateFrom: new Date().toISOString(),
                dateTo: new Date(Date.now() + 7 * 86400000).toISOString(),
                showTasksWithoutDates: true,
            }
            const tasks = new Map<number, TaskModel>()
            tasks.set(1, new TaskModel({ id: 1, title: 'Task 1', startDate: new Date(), endDate: new Date(Date.now() + 2 * 86400000) }))
            tasks.set(2, new TaskModel({ id: 2, title: 'Task 2', startDate: new Date(Date.now() + 86400000), endDate: new Date(Date.now() + 5 * 86400000) }))
            return { filters, tasks }
        },
        template: '<GanttChart :is-loading="false" :filters="filters" :tasks="tasks" default-task-start-date="1970-01-01" default-task-end-date="1970-01-01" />',
    }),
}
