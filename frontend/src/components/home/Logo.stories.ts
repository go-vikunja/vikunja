import type { Meta, StoryObj } from '@storybook/vue3'
import Logo from './Logo.vue'

const meta: Meta<typeof Logo> = {
    title: 'Home/Logo',
    component: Logo,
}
export default meta

type Story = StoryObj<typeof Logo>

export const Default: Story = {
    render: () => ({
        components: { Logo },
        template: '<Logo />',
    }),
}
