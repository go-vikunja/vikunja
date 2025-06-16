import type { Meta, StoryObj } from '@storybook/vue3'
import Card from './Card.vue'

const meta: Meta<typeof Card> = {
    title: 'Misc/Card',
    component: Card,
}
export default meta

type Story = StoryObj<typeof Card>

export const Default: Story = {
    render: () => ({
        components: { Card },
        template: '<Card>Card content</Card>',
    }),
}
