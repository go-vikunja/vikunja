import type { Meta, StoryObj } from '@storybook/vue3'
import PriorityLabel from './PriorityLabel.vue'

const meta: Meta<typeof PriorityLabel> = {
    title: 'Tasks/PriorityLabel',
    component: PriorityLabel,
}
export default meta

type Story = StoryObj<typeof PriorityLabel>

export const Default: Story = {
    render: () => ({
        components: { PriorityLabel },
        template: '<PriorityLabel :priority="3" />',
    }),
}
