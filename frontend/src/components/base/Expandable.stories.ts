import type { Meta, StoryObj } from '@storybook/vue3'
import Expandable from './Expandable.vue'
import BaseButton from './BaseButton.vue'

const meta: Meta<typeof Expandable> = {
    title: 'Base/Expandable',
    component: Expandable,
}
export default meta

type Story = StoryObj<typeof Expandable>

export const Default: Story = {
    render: () => ({
        components: { Expandable, BaseButton },
        data() {
            return { open: false }
        },
        template: `
            <div>
                <BaseButton @click="open = !open">Toggle</BaseButton>
                <Expandable :open="open" :initial-height="60">
                    <p>This is some long text that will be expanded when the button is clicked. Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
                </Expandable>
            </div>
        `,
    }),
}
