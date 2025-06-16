import type { Meta, StoryObj } from '@storybook/vue3'
import UpdateNotification from './UpdateNotification.vue'
import { useBaseStore } from '@/stores/base'

const meta: Meta<typeof UpdateNotification> = {
    title: 'Home/UpdateNotification',
    component: UpdateNotification,
}
export default meta

type Story = StoryObj<typeof UpdateNotification>

export const Default: Story = {
    render: () => ({
        components: { UpdateNotification },
        setup() {
            const baseStore = useBaseStore()
            baseStore.setUpdateAvailable(true)
            return {}
        },
        template: '<UpdateNotification />',
    }),
}
