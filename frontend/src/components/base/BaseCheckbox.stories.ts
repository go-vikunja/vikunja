import type { Meta, StoryObj } from '@storybook/vue3'
import BaseCheckbox from './BaseCheckbox.vue'

const meta: Meta<typeof BaseCheckbox> = {
    title: 'Base/Checkbox',
    component: BaseCheckbox,
}
export default meta

type Story = StoryObj<typeof BaseCheckbox>

export const Default: Story = {
    render: () => ({
        components: { BaseCheckbox },
        data() {
            return { checked: false }
        },
        template: '<BaseCheckbox v-model="checked">Check me</BaseCheckbox>',
    }),
}
