import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import Popup from './Popup.vue'
import BaseButton from '@/components/base/BaseButton.vue'

const meta: Meta<typeof Popup> = {
    title: 'Misc/Popup',
    component: Popup,
}
export default meta

type Story = StoryObj<typeof Popup>

export const Default: Story = {
    render: () => ({
        components: { Popup, BaseButton },
        setup() {
            const open = ref(false)
            return { open }
        },
        template: `
            <Popup :open="open" @update:open="open = $event">
                <template #trigger="{ toggle }">
                    <BaseButton @click="toggle">Toggle</BaseButton>
                </template>
                <template #content>
                    <div class="p-2">Content</div>
                </template>
            </Popup>
        `,
    }),
}
