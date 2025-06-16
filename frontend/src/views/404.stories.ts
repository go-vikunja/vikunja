import type { Meta, StoryObj } from '@storybook/vue3'
import NotFound from './404.vue'

const meta: Meta<typeof NotFound> = {
    title: 'Views/404',
    component: NotFound,
}
export default meta

type Story = StoryObj<typeof NotFound>

export const Overview: Story = {
    render: () => ({
        components: { NotFound },
        template: '<NotFound />',
    }),
}
