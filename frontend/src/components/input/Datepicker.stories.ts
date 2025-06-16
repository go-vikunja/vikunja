import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import Datepicker from './Datepicker.vue'

const meta: Meta<typeof Datepicker> = {
    title: 'Input/Datepicker',
    component: Datepicker,
}
export default meta

type Story = StoryObj<typeof Datepicker>

export const Default: Story = {
    render: () => ({
        components: { Datepicker },
        setup() {
            const date = ref<Date | null>(null)
            return { date }
        },
        template: '<Datepicker v-model="date" />',
    }),
}
