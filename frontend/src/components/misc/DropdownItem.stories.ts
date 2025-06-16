import type { Meta, StoryObj } from '@storybook/vue3'
import DropdownItem from './DropdownItem.vue'

const meta: Meta<typeof DropdownItem> = {
    title: 'Misc/DropdownItem',
    component: DropdownItem,
}
export default meta

type Story = StoryObj<typeof DropdownItem>

export const Default: Story = {
    render: () => ({
        components: { DropdownItem },
        template: '<DropdownItem>Item</DropdownItem>',
    }),
}
