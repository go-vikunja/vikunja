import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import ColorPicker from './ColorPicker.vue'

const meta: Meta<typeof ColorPicker> = {
    title: 'Input/ColorPicker',
    component: ColorPicker,
}
export default meta

type Story = StoryObj<typeof ColorPicker>

export const Default: Story = {
    render: () => ({
        components: { ColorPicker },
        setup() {
            const color = ref('#f2f2f2')
            return { color }
        },
        template: '<ColorPicker v-model="color" />',
    }),
}
