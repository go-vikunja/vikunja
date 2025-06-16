import type { Meta, StoryObj } from '@storybook/vue3'
import BaseButton from './BaseButton.vue'
import { createRouter, createMemoryHistory } from 'vue-router'

const meta: Meta<typeof BaseButton> = {
    title: 'Base/Button',
    component: BaseButton,
    decorators: [() => ({
        components: { BaseButton },
        setup() {
            const router = createRouter({
                history: createMemoryHistory(),
                routes: [{ path: '/', name: 'home', component: { render: () => null } }],
            })
            return { router }
        },
        template: '<story />',
    })],
}
export default meta

type Story = StoryObj<typeof BaseButton>

export const Default: Story = {
    render: () => ({
        components: { BaseButton },
        template: '<BaseButton>Hello!</BaseButton>',
    }),
}

export const Disabled: Story = {
    render: () => ({
        components: { BaseButton },
        template: '<BaseButton disabled>Hello!</BaseButton>',
    }),
}

export const Overview: Story = {
    render: () => ({
        components: { BaseButton },
        template: `
            <div class="buttons">
                <BaseButton>Default</BaseButton>
                <BaseButton disabled>Disabled</BaseButton>
            </div>
        `,
    }),
}

