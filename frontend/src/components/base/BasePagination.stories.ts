import type { Meta, StoryObj } from '@storybook/vue3'
import BasePagination from './BasePagination.vue'

const meta: Meta<typeof BasePagination> = {
    title: 'Base/Pagination',
    component: BasePagination,
}
export default meta

type Story = StoryObj<typeof BasePagination>

export const Default: Story = {
    render: () => ({
        components: { BasePagination },
        template: `
            <BasePagination :total-pages="5" :current-page="2">
                <template #previous="{ disabled }">
                    <button class="pagination-previous" :disabled="disabled">Prev</button>
                </template>
                <template #next="{ disabled }">
                    <button class="pagination-next" :disabled="disabled">Next</button>
                </template>
                <template #page-link="{ page, isCurrent }">
                    <button class="pagination-link" :class="{ 'is-current': isCurrent }">{{ page.number }}</button>
                </template>
            </BasePagination>
        `,
    }),
}
