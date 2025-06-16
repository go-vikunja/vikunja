import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import DatepickerWithValues from './DatepickerWithValues.vue'
import BaseButton from '@/components/base/BaseButton.vue'

const meta: Meta<typeof DatepickerWithValues> = {
    title: 'Date/DatepickerWithValues',
    component: DatepickerWithValues,
}
export default meta

type Story = StoryObj<typeof DatepickerWithValues>

export const Default: Story = {
    render: () => ({
        components: { DatepickerWithValues, BaseButton },
        setup() {
            const date = ref<string | Date | null>(null)
            const open = ref(true)
            return { date, open }
        },
        template: `
            <DatepickerWithValues v-model="date" :open="open" @update:open="open = $event">
                <template #trigger="{ toggle }">
                    <BaseButton @click="toggle">Pick Date</BaseButton>
                </template>
            </DatepickerWithValues>
        `,
    }),
}
