import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import Modal from './Modal.vue'
import BaseButton from '@/components/base/BaseButton.vue'

const meta: Meta<typeof Modal> = {
    title: 'Misc/Modal',
    component: Modal,
}
export default meta

type Story = StoryObj<typeof Modal>

export const Overview: Story = {
    render: () => ({
        components: { Modal, BaseButton },
        setup() {
            const open = ref(false)
            return { open }
        },
        template: `
            <Modal :enabled="open" @close="open = false">
                <template #default>
                    <div class="p-4">This is a modal</div>
                </template>
            </Modal>
            <BaseButton @click="open = true">Open Modal</BaseButton>
        `,
    }),
}
