import type { Meta, StoryObj } from '@storybook/vue3'
import FilterInput from '@/components/project/partials/FilterInput.vue'

const meta: Meta<typeof FilterInput> = {
    title: 'Project/FilterInput',
    component: FilterInput,
}
export default meta

type Story = StoryObj<typeof FilterInput>

export const WithDateValues: Story = {
    render: () => ({
        components: { FilterInput },
        setup() {
            const value = 'dueDate < now && done = false && dueDate > now/w+1w'
            return { value }
        },
        template: '<FilterInput v-model="value" />',
    }),
}
