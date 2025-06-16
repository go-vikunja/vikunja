import type { Meta, StoryObj } from '@storybook/vue3'
import Label from './Label.vue'
import LabelModel from '@/models/label'

const meta: Meta<typeof Label> = {
    title: 'Tasks/Label',
    component: Label,
}
export default meta

type Story = StoryObj<typeof Label>

export const Default: Story = {
    render: () => ({
        components: { Label },
        setup() {
            const label = new LabelModel({
                id: 1,
                title: 'Important',
                hexColor: '#ffd700',
                textColor: '#000000',
            })
            return { label }
        },
        template: '<Label :label="label" />',
    }),
}
