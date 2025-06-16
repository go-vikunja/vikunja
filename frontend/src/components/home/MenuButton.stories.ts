import type { Meta, StoryObj } from '@storybook/vue3'
import MenuButton from './MenuButton.vue'
import { useBaseStore } from '@/stores/base'

const meta: Meta<typeof MenuButton> = {
    title: 'Home/MenuButton',
    component: MenuButton,
}
export default meta

type Story = StoryObj<typeof MenuButton>

export const Default: Story = {
    render: () => ({
        components: { MenuButton },
        setup() {
            const baseStore = useBaseStore()
            baseStore.menuActive = false
            return {}
        },
        template: '<MenuButton />',
    }),
}
