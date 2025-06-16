import type { Meta, StoryObj } from '@storybook/vue3'
import { ref } from 'vue'
import AutocompleteDropdown from './AutocompleteDropdown.vue'

const meta: Meta<typeof AutocompleteDropdown> = {
    title: 'Input/AutocompleteDropdown',
    component: AutocompleteDropdown,
}
export default meta

type Story = StoryObj<typeof AutocompleteDropdown>

export const Default: Story = {
    render: () => ({
        components: { AutocompleteDropdown },
        setup() {
            const value = ref('')
            const options = ['Apple', 'Banana', 'Cherry']
            return { value, options }
        },
        template: `
            <AutocompleteDropdown v-model="value" :options="options">
                <template #input="{ onUpdateField, onFocusField, onKeydown }">
                    <input class="input" :value="value" @input="onUpdateField" @focus="onFocusField" @keydown="onKeydown" />
                </template>
            </AutocompleteDropdown>
        `,
    }),
}
