import type { Meta, StoryObj } from '@storybook/vue3'
import Labels from './Labels.vue'
import LabelModel from '@/models/label'

const meta: Meta<typeof Labels> = {
    title: 'Tasks/Labels',
    component: Labels,
}
export default meta

type Story = StoryObj<typeof Labels>

export const Default: Story = {
    render: () => ({
        components: { Labels },
        setup() {
            const labels = [
                new LabelModel({ id: 1, title: 'Bug', hexColor: '#ff0000', textColor: '#ffffff' }),
                new LabelModel({ id: 2, title: 'Feature', hexColor: '#00ff00', textColor: '#000000' }),
            ]
            return { labels }
        },
        template: '<Labels :labels="labels" />',
    }),
}
