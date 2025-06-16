import type { Meta, StoryObj } from '@storybook/vue3'
import DatemathHelp from './DatemathHelp.vue'

const meta: Meta<typeof DatemathHelp> = {
    title: 'Date/DatemathHelp',
    component: DatemathHelp,
}
export default meta

type Story = StoryObj<typeof DatemathHelp>

export const Default: Story = {
    render: () => ({
        components: { DatemathHelp },
        template: '<DatemathHelp />',
    }),
}
