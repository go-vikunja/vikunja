import type { Meta, StoryObj } from '@storybook/vue3'
import XButton from './Button.vue'

const meta: Meta<typeof XButton> = {
    title: 'Input/Button',
    component: XButton,
}
export default meta

type Story = StoryObj<typeof XButton>

export const Primary: Story = {
    render: () => ({
        components: { XButton },
        template: '<XButton variant="primary">Order pizza!</XButton>',
    }),
}

export const Secondary: Story = {
    render: () => ({
        components: { XButton },
        template: '<XButton variant="secondary">Order spaghetti!</XButton>',
    }),
}

export const Tertiary: Story = {
    render: () => ({
        components: { XButton },
        template: '<XButton variant="tertiary">Order tortellini!</XButton>',
    }),
}
