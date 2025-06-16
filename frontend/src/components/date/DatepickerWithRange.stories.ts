import type { Meta, StoryObj } from '@storybook/vue3'
import { reactive } from 'vue'
import DatepickerWithRange from './DatepickerWithRange.vue'
import BaseButton from '@/components/base/BaseButton.vue'

const meta: Meta<typeof DatepickerWithRange> = {
    title: 'Date/DatepickerWithRange',
    component: DatepickerWithRange,
}
export default meta

type Story = StoryObj<typeof DatepickerWithRange>

export const Default: Story = {
    render: () => ({
        components: { DatepickerWithRange, BaseButton },
        setup() {
            const range = reactive({ dateFrom: '', dateTo: '' })
            return { range }
        },
        template: `
            <DatepickerWithRange v-model="range">
                <template #trigger="{ toggle }">
                    <BaseButton @click="toggle">Select Range</BaseButton>
                </template>
            </DatepickerWithRange>
        `,
    }),
}
