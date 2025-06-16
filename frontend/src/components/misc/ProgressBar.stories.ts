import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import ProgressBar from './ProgressBar.vue'

const meta: Meta<typeof ProgressBar> = {
    title: 'Misc/ProgressBar',
    component: ProgressBar,
}
export default meta

type Story = StoryObj<typeof ProgressBar>

export const Default: Story = {
    render: () => ({
        components: { ProgressBar },
        setup() {
            const value = ref(50)
            return { value }
        },
        template: '<ProgressBar :value="value" />',
    }),
}
