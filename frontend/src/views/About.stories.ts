import type { Meta, StoryObj } from '@storybook/vue3'
import About from './About.vue'
import { useConfigStore } from '@/stores/config'

const meta: Meta<typeof About> = {
    title: 'Views/About',
    component: About,
}
export default meta

type Story = StoryObj<typeof About>

export const Overview: Story = {
    render: () => ({
        components: { About },
        setup() {
            const configStore = useConfigStore()
            configStore.version = 'v1.0.0'
            return {}
        },
        template: '<About />',
    }),
}
