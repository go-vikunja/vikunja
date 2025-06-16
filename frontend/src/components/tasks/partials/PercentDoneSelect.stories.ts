import type { Meta, StoryObj } from '@storybook/vue3'
import PercentDoneSelect from './PercentDoneSelect.vue'
import { ref } from 'vue'

const meta: Meta<typeof PercentDoneSelect> = {
    title: 'Tasks/PercentDoneSelect',
    component: PercentDoneSelect,
}
export default meta

type Story = StoryObj<typeof PercentDoneSelect>

export const Default: Story = {
    render: () => ({
        components: { PercentDoneSelect },
        setup() {
            const percent = ref(0.3)
            return { percent }
        },
        template: '<PercentDoneSelect v-model="percent" />',
    }),
}
