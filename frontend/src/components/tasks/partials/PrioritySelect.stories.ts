import type { Meta, StoryObj } from '@storybook/vue3'
import PrioritySelect from './PrioritySelect.vue'
import { ref } from 'vue'

const meta: Meta<typeof PrioritySelect> = {
    title: 'Tasks/PrioritySelect',
    component: PrioritySelect,
}
export default meta

type Story = StoryObj<typeof PrioritySelect>

export const Default: Story = {
    render: () => ({
        components: { PrioritySelect },
        setup() {
            const priority = ref(2)
            return { priority }
        },
        template: '<PrioritySelect v-model="priority" />',
    }),
}
