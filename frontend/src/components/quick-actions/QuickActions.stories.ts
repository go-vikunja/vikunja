import type { Meta, StoryObj } from '@storybook/vue3'
import QuickActions from './QuickActions.vue'
import { useBaseStore } from '@/stores/base'

const meta: Meta<typeof QuickActions> = {
    title: 'QuickActions/QuickActions',
    component: QuickActions,
}
export default meta

type Story = StoryObj<typeof QuickActions>

export const Default: Story = {
    render: () => ({
        components: { QuickActions },
        setup() {
            const baseStore = useBaseStore()
            baseStore.setQuickActionsActive(true)
            return {}
        },
        template: '<QuickActions />',
    }),
}
