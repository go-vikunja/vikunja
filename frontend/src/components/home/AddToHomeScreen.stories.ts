import type { Meta, StoryObj } from '@storybook/vue3'
import AddToHomeScreen from './AddToHomeScreen.vue'
import { useBaseStore } from '@/stores/base'

const meta: Meta<typeof AddToHomeScreen> = {
    title: 'Home/AddToHomeScreen',
    component: AddToHomeScreen,
}
export default meta

type Story = StoryObj<typeof AddToHomeScreen>

export const Default: Story = {
    render: () => ({
        components: { AddToHomeScreen },
        setup() {
            const baseStore = useBaseStore()
            baseStore.setUpdateAvailable(false)
            localStorage.setItem('hideAddToHomeScreenMessage', 'false')
            return {}
        },
        template: '<AddToHomeScreen />',
    }),
}
