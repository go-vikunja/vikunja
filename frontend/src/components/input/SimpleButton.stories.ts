import type { Meta, StoryObj } from '@storybook/vue3'
import SimpleButton from './SimpleButton.vue'

const meta: Meta<typeof SimpleButton> = {
    title: 'Input/SimpleButton',
    component: SimpleButton,
}
export default meta

type Story = StoryObj<typeof SimpleButton>

export const Default: Story = {
    render: () => ({
        components: { SimpleButton },
        template: '<SimpleButton>Click me</SimpleButton>',
    }),
}
