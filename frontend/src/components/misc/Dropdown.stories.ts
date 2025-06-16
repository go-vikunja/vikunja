import type { Meta, StoryObj } from '@storybook/vue3'
import Dropdown from './Dropdown.vue'
import DropdownItem from './DropdownItem.vue'
import BaseButton from '@/components/base/BaseButton.vue'

const meta: Meta<typeof Dropdown> = {
    title: 'Misc/Dropdown',
    component: Dropdown,
}
export default meta

type Story = StoryObj<typeof Dropdown>

export const Default: Story = {
    render: () => ({
        components: { Dropdown, DropdownItem, BaseButton },
        template: `
            <Dropdown>
                <template #trigger="{ toggleOpen }">
                    <BaseButton @click="toggleOpen">Open</BaseButton>
                </template>
                <DropdownItem>One</DropdownItem>
                <DropdownItem>Two</DropdownItem>
            </Dropdown>
        `,
    }),
}
